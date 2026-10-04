package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/auth"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/config"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/output"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/update"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/version"
)

const repoSlug = "piyush-gambhir/jira-cli"

// groupDispatcherKey marks command groups whose only job is to dispatch to a
// sub-command (or print help). PersistentPreRunE uses it to skip client setup
// for them — set by enforceGroupNoArgs, never on a command with its own RunE.
const groupDispatcherKey = "jira.groupDispatcher"

var (
	// Global flags
	outputFormat   string
	profileFlag    string
	siteFlag       string
	emailFlag      string
	tokenFlag      string
	userFlag       string
	apiVersionFlag string
	insecureFlag   bool
	noColorFlag    bool
	verboseFlag    bool
	readOnlyFlag   bool
	noInputFlag    bool
	quietFlag      bool

	// Shared state set during PersistentPreRunE
	cfg               *config.Config
	activeProfileName string
	jiraClient        *client.Client
	outFormat         output.Format

	// OutputFormat is the format resolved in PersistentPreRunE; ErrorFormat
	// uses it to format top-level errors.
	OutputFormat string

	// updateResult receives the background release check's latest version;
	// nil when no check was started for this command.
	updateResult  chan string
	updateChecker *update.Checker
	// updateNoticeWait is how long PersistentPostRunE waits for a background
	// check that this run started and that has not answered yet.
	updateNoticeWait = updateNoticeGrace
)

// updateNoticeGrace bounds the wait for this run's own release check. The check
// is recorded before GitHub is asked, so an answer lost to a fast command would
// hide the notice for a day; waiting up to a second (at most once a day, only
// in an interactive terminal) keeps it. A cached answer never waits.
const updateNoticeGrace = time.Second

var rootCmd = &cobra.Command{
	Use:   "jira",
	Short: "Jira CLI — manage Jira from the command line",
	Long: `A command-line interface for Jira (Cloud and Server/Data Center).

Manage issues, comments, worklogs, attachments, links, transitions, JQL search,
projects, users, boards and sprints from the terminal. Designed for both humans
and coding agents.

Quick start:
  jira auth login --type api_token  # authenticate with a Cloud API token
  jira whoami                       # confirm who you are
  jira issue search --jql "assignee = currentUser() AND statusCategory != Done"
  jira issue create -p PROJ --type Task --summary "Title" -d "Description"

Supports every Jira auth method (API token, scoped token, OAuth 2.0 3LO, PAT,
username/password) — see "jira auth login --help" and docs/CREDENTIALS.md.

All list/get commands support -o json and -o yaml for machine-readable output.

Full command reference (for agents/LLMs): https://github.com/piyush-gambhir/jira-cli/blob/main/docs/llms.txt
Claude Code skill: https://github.com/piyush-gambhir/jira-cli/blob/main/jira/SKILL.md`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Env fallbacks for the agent-friendly flags. An explicit flag, including
		// --quiet=false or --no-input=false, wins over the environment.
		if !cmd.Flags().Changed("no-input") {
			if v, ok := os.LookupEnv("JIRA_NO_INPUT"); ok && truthy(v) {
				noInputFlag = true
			}
		}
		if !cmd.Flags().Changed("quiet") {
			if v, ok := os.LookupEnv("JIRA_QUIET"); ok && truthy(v) {
				quietFlag = true
			}
		}

		updateResult = nil
		if updateNotifierEnabled(cmd) {
			startBackgroundUpdateCheck()
		}

		// Parse output format early (also used by main.go error handling).
		var err error
		outFormat, err = output.ParseFormat(outputFormat)
		if err != nil {
			return err
		}
		OutputFormat = outputFormat

		// Commands that never need an authenticated client. They are matched by
		// their top-level command (so `completion bash` and Cobra's hidden
		// __complete are covered), never by leaf name: `project update` needs one.
		if isClientlessCommand(cmd) {
			return nil
		}
		if cmd.Parent() != nil && cmd.Parent().Name() == "auth" {
			return loadConfigOnly()
		}
		// Pure dispatcher groups (issue, project, ...) only show help or reject an
		// unknown sub-command, so they need no client. Commands that have their own
		// run function still get one even if they also have sub-commands — e.g.
		// `jira status`, which reports the live connection.
		if cmd.Annotations[groupDispatcherKey] == "true" {
			return nil
		}
		if !cmd.Runnable() {
			return nil
		}

		if err := loadConfigOnly(); err != nil {
			return err
		}
		if outputFormat == "" && cfg.Defaults.Output != "" {
			outFormat, _ = output.ParseFormat(cfg.Defaults.Output)
			OutputFormat = cfg.Defaults.Output
		}

		profile, err := resolveProfile(cmd)
		if err != nil {
			return err
		}
		if err := checkReadOnly(cmd, profile); err != nil {
			return err
		}

		persist := func(updated config.Profile) error {
			return config.PersistProfile(activeProfileName, updated)
		}
		authr, err := auth.New(profile, activeProfileName, persist)
		if err != nil {
			return err
		}
		jiraClient = client.NewClient(authr, profile.EffectiveAPIVersion(), profile.Insecure, verboseFlag).WithContext(cmd.Context())
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if updateResult == nil {
			return nil
		}
		// A cached answer is already waiting; a check this run started gets
		// updateNoticeWait to answer before the command exits without a notice.
		var latest string
		select {
		case latest = <-updateResult:
		default:
			select {
			case latest = <-updateResult:
			case <-time.After(updateNoticeWait):
				return nil
			}
		}
		printUpdateNotice(cmd.ErrOrStderr(), updateChecker, latest)
		return nil
	},
}

func loadConfigOnly() error {
	var err error
	cfg, err = config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	return nil
}

func resolveProfile(cmd *cobra.Command) (config.Profile, error) {
	flags := config.FlagValues{
		Site:        siteFlag,
		Email:       emailFlag,
		Token:       tokenFlag,
		User:        userFlag,
		Insecure:    insecureFlag,
		SiteSet:     cmd.Flags().Changed("site"),
		EmailSet:    cmd.Flags().Changed("email"),
		TokenSet:    cmd.Flags().Changed("token"),
		UserSet:     cmd.Flags().Changed("user"),
		InsecureSet: cmd.Flags().Changed("insecure"),
	}

	effProfile := profileFlag
	if effProfile == "" {
		effProfile = os.Getenv("JIRA_PROFILE")
	}
	activeProfileName = effProfile
	if activeProfileName == "" {
		activeProfileName = cfg.CurrentProfile
	}
	if activeProfileName == "" {
		activeProfileName = "default"
	}

	profile, err := config.ResolveAuth(flags, os.LookupEnv, cfg, effProfile)
	if err != nil {
		return config.Profile{}, fmt.Errorf("resolving auth: %w", err)
	}
	if apiVersionFlag != "" {
		profile.APIVersion = apiVersionFlag
	}
	if profile.Site == "" {
		return config.Profile{}, fmt.Errorf("no Jira site configured. Run 'jira auth login' or set JIRA_SITE")
	}
	return profile, nil
}

func checkReadOnly(cmd *cobra.Command, profile config.Profile) error {
	effective := profile.ReadOnly
	if readOnlyFlag {
		effective = true
	}
	if effective && cmd.Annotations != nil && cmd.Annotations["mutates"] == "true" {
		return fmt.Errorf("command '%s' is blocked in read-only mode; remove read_only from the profile or disable the read-only environment setting to permit writes", cmd.CommandPath())
	}
	return nil
}

// isClientlessCommand reports whether cmd belongs to a top-level command that
// needs no Jira client: update, version, completion, help, or __complete*.
func isClientlessCommand(cmd *cobra.Command) bool {
	top := cmd
	for top.HasParent() && top.Parent().HasParent() {
		top = top.Parent()
	}
	switch name := top.Name(); name {
	case "update", "version", "completion", "help":
		return true
	default:
		return strings.HasPrefix(name, "__complete")
	}
}

// updateNotifierEnabled reports whether this run may check GitHub for a newer
// release in the background. Scripts, CI, and agents are never disturbed: the
// check is skipped (no network, no output) unless stderr is a terminal, and
// also under CI, JIRA_NO_UPDATE_NOTIFIER, NO_UPDATE_NOTIFIER, --quiet or
// JIRA_QUIET, for dev builds, and for update/version/completion/help.
func updateNotifierEnabled(cmd *cobra.Command) bool {
	if isClientlessCommand(cmd) {
		return false
	}
	return !quietFlag &&
		!update.NotifierDisabledByEnv("JIRA", os.Getenv) &&
		update.IsRelease(version.Version) &&
		stderrIsTerminal()
}

// startBackgroundUpdateCheck answers from a fresh cache right away (a small
// local read, so even a fast command can show the notice) and otherwise asks
// GitHub in a goroutine that PersistentPostRunE waits for at most
// updateNoticeWait.
func startBackgroundUpdateCheck() {
	ch := make(chan string, 1)
	checker := newUpdateChecker(3 * time.Second)
	updateResult, updateChecker = ch, checker
	if latest, ok := checker.Cached(); ok {
		ch <- latest
		return
	}
	go func() {
		latest, _ := checker.Latest(context.Background(), false)
		ch <- latest
	}()
}

func truthy(v string) bool {
	v = strings.TrimSpace(v)
	return strings.EqualFold(v, "true") || v == "1"
}

// RootCmd returns the root command for use in main.go.
func RootCmd() *cobra.Command { return rootCmd }

// Execute runs the root command and reports a failure on stderr in the
// selected output format (a structured object for -o json), exiting 1.
func Execute() {
	if runtime.GOOS == "windows" {
		// A Windows update leaves the replaced binary at jira.exe.old.
		if exe, err := executablePath(); err == nil {
			update.RemoveStaleOld(exe)
		}
	}
	if err := rootCmd.Execute(); err != nil {
		statusCode := 0
		var apiErr *client.APIError
		if errors.As(err, &apiErr) {
			statusCode = apiErr.StatusCode
		}
		output.WriteError(os.Stderr, ErrorFormat(os.Args[1:]), err, statusCode)
		os.Exit(1)
	}
}

// ErrorFormat picks the output format for a top-level error. PersistentPreRunE
// resolves OutputFormat (including the config default), but Cobra rejects bad
// arguments and unknown flags before that hook runs, so fall back to the
// -o/--output value in args.
func ErrorFormat(args []string) output.Format {
	name := OutputFormat
	if name == "" {
		name = outputFlagFromArgs(args)
	}
	format, err := output.ParseFormat(name)
	if err != nil {
		return output.FormatTable
	}
	return format
}

// outputFlagFromArgs returns the -o/--output value in args (-o json, -ojson,
// -o=json, --output json, --output=json). It walks args with the flag
// definitions of the command they select, so a value glued to another short
// flag (-shttps://x) is skipped whole instead of being read as a cluster of
// one-letter flags. Parsing of an argument stops at an unknown flag.
func outputFlagFromArgs(args []string) string {
	cmd, _, err := rootCmd.Find(args)
	if err != nil || cmd == nil {
		cmd = rootCmd
	}
	cmd.InheritedFlags() // merges the persistent flags into cmd.Flags()
	flags := cmd.Flags()

	format := ""
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--":
			return format
		case strings.HasPrefix(a, "--"):
			name, value, hasValue := strings.Cut(a[2:], "=")
			f := flags.Lookup(name)
			if f == nil {
				continue
			}
			if !hasValue && f.NoOptDefVal == "" && i+1 < len(args) {
				i++
				value = args[i]
			}
			if f.Name == "output" {
				format = value
			}
		case strings.HasPrefix(a, "-") && len(a) > 1:
			for j := 1; j < len(a); j++ {
				f := flags.ShorthandLookup(a[j : j+1])
				if f == nil {
					break
				}
				if f.NoOptDefVal != "" { // a boolean: the cluster continues
					continue
				}
				value := strings.TrimPrefix(a[j+1:], "=")
				if j+1 == len(a) && i+1 < len(args) {
					i++
					value = args[i]
				}
				if f.Name == "output" {
					format = value
				}
				break
			}
		}
	}
	return format
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.StringVarP(&outputFormat, "output", "o", "", "Output format: table, json, yaml")
	pf.StringVar(&profileFlag, "profile", "", "Configuration profile to use")
	pf.StringVarP(&siteFlag, "site", "s", "", "Jira site URL override (https://your.atlassian.net or https://jira.host)")
	pf.StringVarP(&emailFlag, "email", "e", "", "Atlassian account email override (Cloud)")
	pf.StringVarP(&tokenFlag, "token", "t", "", "API token / PAT override")
	pf.StringVarP(&userFlag, "user", "u", "", "Username override (Server/DC basic auth)")
	pf.StringVar(&apiVersionFlag, "api-version", "", "Platform REST API version override: 3 (Cloud) or 2 (Server/DC)")
	pf.BoolVarP(&insecureFlag, "insecure", "k", false, "Skip TLS certificate verification")
	pf.BoolVar(&noColorFlag, "no-color", false, "Disable color output")
	pf.BoolVarP(&verboseFlag, "verbose", "v", false, "Verbose HTTP logging to stderr")
	pf.BoolVar(&readOnlyFlag, "read-only", false, "Block write operations (safety mode for agents)")
	pf.BoolVar(&noInputFlag, "no-input", false, "Disable all interactive prompts (for CI/agents)")
	pf.BoolVarP(&quietFlag, "quiet", "q", false, "Suppress informational output")

	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newUpdateCmd())
	rootCmd.AddCommand(newAuthCmd())
	rootCmd.AddCommand(newWhoamiCmd())
	rootCmd.AddCommand(newStatusCmd())
	rootCmd.AddCommand(newIssueCmd())
	rootCmd.AddCommand(newProjectCmd())
	rootCmd.AddCommand(newUserCmd())
	rootCmd.AddCommand(newFieldCmd())
	rootCmd.AddCommand(newBoardCmd())
	rootCmd.AddCommand(newSprintCmd())
	rootCmd.AddCommand(newEpicCmd())

	// Expanded API coverage (added groups)
	rootCmd.AddCommand(newComponentCmd())
	rootCmd.AddCommand(newReleaseCmd())
	rootCmd.AddCommand(newFilterCmd())
	rootCmd.AddCommand(newDashboardCmd())
	rootCmd.AddCommand(newIssueTypeCmd())
	rootCmd.AddCommand(newPriorityCmd())
	rootCmd.AddCommand(newResolutionCmd())
	rootCmd.AddCommand(newLabelCmd())
	rootCmd.AddCommand(newGroupCmd())
	rootCmd.AddCommand(newPermissionCmd())
	rootCmd.AddCommand(newJQLCmd())
	rootCmd.AddCommand(newWebhookCmd())
	rootCmd.AddCommand(newBacklogCmd())
	rootCmd.AddCommand(newServerInfoCmd())

	// Make command groups reject unknown sub-commands (e.g. `jira issue bogus`)
	// with a non-zero exit instead of silently printing help and exiting 0.
	enforceGroupNoArgs(rootCmd)
}

// enforceGroupNoArgs walks the command tree and, for every command that groups
// sub-commands but has no run function of its own, installs a RunE that returns
// an "unknown command" error (exit 1) when given an unrecognized sub-command,
// and otherwise prints help. Cobra short-circuits a non-runnable parent to help
// (exit 0) before arg validation runs, so a typo like `jira issue bogus` would
// otherwise look like success — bad for scripts and agents.
func enforceGroupNoArgs(c *cobra.Command) {
	if c.HasSubCommands() && c.Run == nil && c.RunE == nil {
		if c.Annotations == nil {
			c.Annotations = map[string]string{}
		}
		c.Annotations[groupDispatcherKey] = "true"
		c.RunE = func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("unknown command %q for %q\nRun '%s --help' for available commands", args[0], cmd.CommandPath(), cmd.CommandPath())
			}
			return cmd.Help()
		}
	}
	for _, sub := range c.Commands() {
		enforceGroupNoArgs(sub)
	}
}
