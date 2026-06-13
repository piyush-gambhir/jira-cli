package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/internal/client"
	"github.com/piyush-gambhir/jira-cli/internal/output"
)

func newWhoamiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show the authenticated user",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			me, err := jiraClient.Myself()
			if err != nil {
				return err
			}
			info("Authenticated via %s", jiraClient.AuthDescription())
			def := &output.TableDef{
				Headers: []string{"ACCOUNT ID", "NAME", "EMAIL", "ACTIVE", "TIMEZONE"},
				RowFunc: func(item interface{}) []string {
					u := item.(*client.User)
					return []string{u.AccountID, u.DisplayName, dash(u.EmailAddress), fmt.Sprintf("%v", u.Active), dash(u.TimeZone)}
				},
			}
			return render(me, def)
		},
	}
}
