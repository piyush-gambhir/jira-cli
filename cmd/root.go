package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/internal/auth"
	"github.com/piyush-gambhir/jira-cli/internal/client"
	"github.com/piyush-gambhir/jira-cli/internal/config"
	"github.com/piyush-gambhir/jira-cli/internal/output"
	"github.com/piyush-gambhir/jira-cli/internal/update"
	"github.com/piyush-gambhir/jira-cli/internal/version"
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
	activeProfile     config.Profile
	activeProfileName string
	jiraClient        *client.Client
	outFormat         output.Format

	// OutputFormat is exported so main.go can format top-level errors.
	OutputFormat string

	updateResult chan *update.UpdateInfo
)

var rootCmd = &cobra.Command{
	Use:   "jira",
	Short: "Jira CLI — manage Jira from the command line",
	Long: `A command-line interface for Jira (Cloud and Server/Data Center).

Manage issues, comments, worklogs, attachments, links, transitions, JQL search,
projects, users, boards and sprints from the terminal. Designed for both humans
and coding agents.

Quick start:
  jira auth login                 # authenticate (Cloud API token by default)
  jira whoami                     # confirm who you are
  jira issue search --jql "assignee = currentUser() AND statusCategory != Done"
  jira issue create -p PROJ --type Task --summary "Title" -d "Description"

Supports every Jira auth method (API token, scoped token, OAuth 2.0 3LO, PAT,
username/password) — see "jira auth login --help" and docs/CREDENTIALS.md.

All list/get commands support -o json and -o yaml for machine-readable output.

Full command reference (for agents/LLMs): https://jira-cli.pages.dev/llms.txt`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Env fallbacks for the agent-friendly flags.
		if !noInputFlag {
			if v, ok := os.LookupEnv("JIRA_NO_INPUT"); ok && truthy(v) {
				noInputFlag = true
			}
		}
		if !quietFlag {
			if v, ok := os.LookupEnv("JIRA_QUIET"); ok && truthy(v) {
				quietFlag = true
			}
		}

		cmdName := cmd.Name()
		if cmdName != "update" && cmdName != "version" {
			startBackgroundUpdateCheck()
		}

		// Parse output format early (also used by main.go error handling).
		var err error
		outFormat, err = output.ParseFormat(outputFormat)
		if err != nil {
			return err
		}
		OutputFormat = outputFormat

		// Commands that never need an authenticated client. version/update/completion
		// are matched ONLY at the top level — otherwise a subcommand of the same name
		// (e.g. `project update`) would wrongly be left without a client.
		isTopLevel := cmd.Parent() == nil || cmd.Parent() == cmd.Root()
		if cmdName == "help" || (isTopLevel && (cmdName == "version" || cmdName == "update" || cmdName == "completion")) {
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
		activeProfile = profile

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
		jiraClient = client.NewClient(authr, profile.EffectiveAPIVersion(), profile.Insecure, verboseFlag)
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		cmdName := cmd.Name()
		if cmdName == "update" || cmdName == "version" || updateResult == nil {
			return nil
		}
		select {
		case info := <-updateResult:
			if info != nil && info.Available {
				update.PrintUpdateNotice(os.Stderr, info)
			}
		case <-time.After(1500 * time.Millisecond):
		}
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
	if cmd.Flags().Changed("read-only") {
		effective = readOnlyFlag
	}
	if effective && cmd.Annotations != nil && cmd.Annotations["mutates"] == "true" {
		return fmt.Errorf("command '%s' is blocked in read-only mode (use --read-only=false or remove read_only from the profile)", cmd.CommandPath())
	}
	return nil
}

func startBackgroundUpdateCheck() {
	updateResult = make(chan *update.UpdateInfo, 1)
	go func() {
		info, _ := update.CheckForUpdate(version.Version, repoSlug, config.ConfigDir(), false)
		updateResult <- info
	}()
}

func truthy(v string) bool { return v != "" && v != "0" && v != "false" }

// RootCmd returns the root command for use in main.go.
func RootCmd() *cobra.Command { return rootCmd }

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		statusCode := 0
		var apiErr *client.APIError
		if errors.As(err, &apiErr) {
			statusCode = apiErr.StatusCode
		}
		output.WriteError(os.Stderr, outFormat, err, statusCode)
		os.Exit(1)
	}
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
