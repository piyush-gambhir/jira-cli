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
  jira issue create -p PROJ -t Task -s "Title" -d "Description"

Supports every Jira auth method (API token, scoped token, OAuth 2.0 3LO, PAT,
username/password) — see "jira auth login --help" and docs/CREDENTIALS.md.

All list/get commands support -o json and -o yaml for machine-readable output.`,
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

		// Commands that never need an authenticated client.
		if cmdName == "version" || cmdName == "help" || cmdName == "update" || cmdName == "completion" {
			return nil
		}
		if cmd.Parent() != nil && cmd.Parent().Name() == "auth" {
			return loadConfigOnly()
		}
		// Parent/group commands (no RunE of their own) don't need a client.
		if cmd.Runnable() && cmd.RunE == nil && cmd.Run == nil {
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

	activeProfileName = profileFlag
	if activeProfileName == "" {
		activeProfileName = cfg.CurrentProfile
	}
	if activeProfileName == "" {
		activeProfileName = "default"
	}

	profile, err := config.ResolveAuth(flags, os.LookupEnv, cfg, profileFlag)
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
}
