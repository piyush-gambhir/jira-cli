package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/internal/client"
	"github.com/piyush-gambhir/jira-cli/internal/output"
)

func newUserCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "user",
		Aliases: []string{"users"},
		Short:   "Find and inspect users",
	}
	cmd.AddCommand(newUserSearchCmd(), newUserGetCmd())
	cmd.AddCommand(newUserListCmd(), newUserBulkCmd(), newUserAssignableCmd(), newUserGroupsCmd())
	return cmd
}

func userTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"ACCOUNT ID", "NAME", "EMAIL", "ACTIVE"},
		RowFunc: func(item interface{}) []string {
			u := item.(client.User)
			return []string{u.AccountID, u.DisplayName, dash(u.EmailAddress), fmt.Sprintf("%v", u.Active)}
		},
	}
}

func newUserSearchCmd() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search users by name or email",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			users, err := jiraClient.SearchUsers(args[0], limit)
			if err != nil {
				return err
			}
			return render(users, userTable())
		},
	}
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Maximum users to return")
	return cmd
}

func newUserGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <accountId>",
		Short: "Get a user by accountId",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			u, err := jiraClient.GetUser(args[0])
			if err != nil {
				return err
			}
			return render(*u, userTable())
		},
	}
}
