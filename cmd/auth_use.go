package cmd

import (
	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/internal/config"
)

func newAuthUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <profile>",
		Short: "Set the active profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.SetCurrentProfile(cfg, args[0]); err != nil {
				return err
			}
			if err := config.Save(cfg); err != nil {
				return err
			}
			info("Active profile is now %q", args[0])
			return nil
		},
	}
}
