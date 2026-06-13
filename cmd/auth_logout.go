package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/internal/config"
)

func newAuthLogoutCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove a saved profile (and its stored credentials)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			target := name
			if target == "" {
				target = cfg.CurrentProfile
			}
			if target == "" {
				return fmt.Errorf("no profile to remove; pass --name")
			}
			if err := config.DeleteProfile(cfg, target); err != nil {
				return err
			}
			if err := config.Save(cfg); err != nil {
				return err
			}
			info("Removed profile %q", target)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Profile to remove (defaults to the current profile)")
	return cmd
}
