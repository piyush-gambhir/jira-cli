package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/update"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/version"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the CLI version",
		Long: `Print the CLI version, commit, and build time.

When a release check from the last 24 hours is cached (see "jira update --help"),
two more lines report the latest release and whether an update is available.
jira version never contacts the network.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()
			fmt.Fprintln(w, version.Info())
			if !update.IsRelease(version.Version) {
				return nil
			}
			if latest, ok := newUpdateChecker(0).CachedLatest(); ok {
				fmt.Fprintf(w, "latest: %s\n", latest)
				fmt.Fprintf(w, "update_available: %t\n", update.Newer(latest, version.Version))
			}
			return nil
		},
	}
}
