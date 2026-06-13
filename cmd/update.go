package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/internal/config"
	"github.com/piyush-gambhir/jira-cli/internal/update"
	"github.com/piyush-gambhir/jira-cli/internal/version"
)

func newUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Check for a newer release of the CLI",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			info, err := update.CheckForUpdate(version.Version, repoSlug, config.ConfigDir(), true)
			if err != nil {
				return fmt.Errorf("checking for updates: %w", err)
			}
			if info.Available {
				fmt.Printf("A new version is available: %s -> %s\n", info.CurrentVersion, info.LatestVersion)
				if info.URL != "" {
					fmt.Printf("  %s\n", info.URL)
				}
				fmt.Printf("Upgrade with: go install %s@latest\n", "github.com/"+repoSlug)
			} else {
				fmt.Printf("jira %s is up to date.\n", version.Version)
			}
			return nil
		},
	}
}
