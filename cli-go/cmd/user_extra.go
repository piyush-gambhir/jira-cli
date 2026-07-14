package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/output"
)

// userGroupTable renders the {name,groupId} groups a user belongs to.
func userGroupTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"NAME", "GROUP ID"},
		RowFunc: func(item interface{}) []string {
			g := item.(client.Group)
			return []string{dash(g.Name), dash(g.GroupID)}
		},
	}
}

func newUserListCmd() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all users on the site",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			users, err := jiraClient.ListAllUsers(limit)
			if err != nil {
				return err
			}
			return render(users, userTable())
		},
	}
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Maximum users to return")
	return cmd
}

func newUserBulkCmd() *cobra.Command {
	var accountIDs []string
	cmd := &cobra.Command{
		Use:   "bulk",
		Short: "Get multiple users by accountId in one request",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(accountIDs) == 0 {
				return fmt.Errorf("at least one --account-id is required")
			}
			users, err := jiraClient.BulkUsers(accountIDs)
			if err != nil {
				return err
			}
			return render(users, userTable())
		},
	}
	cmd.Flags().StringArrayVar(&accountIDs, "account-id", nil, "Account id to look up (repeatable)")
	return cmd
}

func newUserAssignableCmd() *cobra.Command {
	var project, issueKey string
	var limit int
	cmd := &cobra.Command{
		Use:   "assignable [query]",
		Short: "List users assignable to issues (the correct assign/create picker)",
		Long: `List users assignable to issues, optionally filtered by a name/email query.

Scope the search with --project or --issue (--issue wins when both are set).
This is the correct picker to use before 'issue assign' or 'issue create'.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if project == "" && issueKey == "" {
				return fmt.Errorf("--project or --issue is required")
			}
			var query string
			if len(args) == 1 {
				query = args[0]
			}
			users, err := jiraClient.AssignableUsers(query, project, issueKey, limit)
			if err != nil {
				return err
			}
			return render(users, userTable())
		},
	}
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project key to scope assignable users")
	cmd.Flags().StringVar(&issueKey, "issue", "", "Issue key to scope assignable users")
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Maximum users to return")
	return cmd
}

func newUserGroupsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "groups <accountId>",
		Short: "List the groups a user belongs to",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groups, err := jiraClient.UserGroups(args[0])
			if err != nil {
				return err
			}
			return render(groups, userGroupTable())
		},
	}
}
