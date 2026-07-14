package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/output"
)

func newGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "group",
		Aliases: []string{"groups"},
		Short:   "List, inspect, and manage user groups",
		Long: `List, search, and manage Jira user groups and their members.

Create/delete and add/remove-user require the manage:jira-configuration scope /
site-admin permission.`,
	}
	cmd.AddCommand(
		newGroupListCmd(),
		newGroupFindCmd(),
		newGroupMembersCmd(),
		newGroupAddUserCmd(),
		newGroupRemoveUserCmd(),
		newGroupCreateCmd(),
		newGroupDeleteCmd(),
	)
	return cmd
}

func groupTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"NAME", "GROUP ID"},
		RowFunc: func(item interface{}) []string {
			g := item.(client.Group)
			return []string{dash(g.Name), dash(g.GroupID)}
		},
	}
}

func groupMemberTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"ACCOUNT ID", "NAME", "EMAIL", "ACTIVE"},
		RowFunc: func(item interface{}) []string {
			u := item.(client.User)
			active := "no"
			if u.Active {
				active = "yes"
			}
			return []string{dash(u.AccountID), dash(u.DisplayName), dash(u.EmailAddress), active}
		},
	}
}

func newGroupListCmd() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List groups",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			groups, err := jiraClient.ListGroups(limit)
			if err != nil {
				return err
			}
			return render(groups, groupTable())
		},
	}
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Maximum groups to return")
	return cmd
}

func newGroupFindCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "find <name>",
		Short: "Search for groups by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groups, err := jiraClient.FindGroups(args[0])
			if err != nil {
				return err
			}
			return render(groups, groupTable())
		},
	}
}

func newGroupMembersCmd() *cobra.Command {
	var groupID string
	var limit int
	var inactive bool
	cmd := &cobra.Command{
		Use:   "members <name>",
		Short: "List the members of a group",
		Long: `List the members of a group. Identify the group by name (positional arg)
or, to disambiguate same-named groups, by --group-id.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			members, err := jiraClient.GroupMembers(args[0], groupID, limit, inactive)
			if err != nil {
				return err
			}
			return render(members, groupMemberTable())
		},
	}
	cmd.Flags().StringVar(&groupID, "group-id", "", "Group id (preferred over name when groups share a name)")
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Maximum members to return")
	cmd.Flags().BoolVar(&inactive, "inactive", false, "Include inactive (deactivated) users")
	return cmd
}

func newGroupAddUserCmd() *cobra.Command {
	var userRef string
	cmd := &cobra.Command{
		Use:         "add-user <groupname>",
		Short:       "Add a user to a group",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if userRef == "" {
				return fmt.Errorf("--user is required")
			}
			acct, err := jiraClient.ResolveUser(userRef)
			if err != nil {
				return err
			}
			if err := jiraClient.AddGroupUser(args[0], acct); err != nil {
				return err
			}
			info("Added user %s to group %s", acct, args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&userRef, "user", "", "User to add (email, name, @me, id:<accountId>) (required)")
	return cmd
}

func newGroupRemoveUserCmd() *cobra.Command {
	var userRef string
	cmd := &cobra.Command{
		Use:         "remove-user <groupname>",
		Short:       "Remove a user from a group",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if userRef == "" {
				return fmt.Errorf("--user is required")
			}
			acct, err := jiraClient.ResolveUser(userRef)
			if err != nil {
				return err
			}
			if err := jiraClient.RemoveGroupUser(args[0], acct); err != nil {
				return err
			}
			info("Removed user %s from group %s", acct, args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&userRef, "user", "", "User to remove (email, name, @me, id:<accountId>) (required)")
	return cmd
}

func newGroupCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "create <name>",
		Short:       "Create a group",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g, err := jiraClient.CreateGroup(args[0])
			if err != nil {
				return err
			}
			info("Created group %s", args[0])
			return render(*g, groupTable())
		},
	}
	return cmd
}

func newGroupDeleteCmd() *cobra.Command {
	var swapGroup string
	var yes bool
	cmd := &cobra.Command{
		Use:         "delete <name>",
		Short:       "Delete a group",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		Long: `Delete a group. Use --swap-group to transfer any restrictions or permissions
that referenced the deleted group to another group.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				if noInputFlag {
					return fmt.Errorf("refusing to delete group %s without --yes", args[0])
				}
				ans, _ := prompt(fmt.Sprintf("Delete group %s? [y/N]: ", args[0]))
				if !strings.EqualFold(strings.TrimSpace(ans), "y") {
					info("Aborted.")
					return nil
				}
			}
			if err := jiraClient.DeleteGroup(args[0], swapGroup); err != nil {
				return err
			}
			info("Deleted group %s", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&swapGroup, "swap-group", "", "Group to receive restrictions/permissions of the deleted group")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")
	return cmd
}
