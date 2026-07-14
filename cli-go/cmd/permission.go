package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/output"
)

// permissionDefaultKeys is the sensible common set queried by `permission mine`
// when --keys is omitted (the permissions query param is now required by Jira).
var permissionDefaultKeys = []string{
	"BROWSE_PROJECTS",
	"CREATE_ISSUES",
	"EDIT_ISSUES",
	"ASSIGN_ISSUES",
	"TRANSITION_ISSUES",
	"ADD_COMMENTS",
	"DELETE_ISSUES",
	"ADMINISTER",
	"ADMINISTER_PROJECTS",
}

func newPermissionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "permission",
		Aliases: []string{"permissions", "perm"},
		Short:   "Inspect permissions and permitted projects",
		Long: `Check which permissions you hold (globally or scoped to a project/issue),
list the permissions the instance defines, and find the projects in which you
hold a given set of permissions.`,
	}
	cmd.AddCommand(
		newPermissionMineCmd(),
		newPermissionPermittedProjectsCmd(),
		newPermissionListCmd(),
	)
	return cmd
}

// permissionHaveTable renders KEY / NAME / HAVE for the caller's permissions.
func permissionHaveTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"KEY", "NAME", "HAVE"},
		RowFunc: func(item interface{}) []string {
			p := item.(client.Permission)
			return []string{p.Key, dash(p.Name), permissionYesNo(p.HavePermission)}
		},
	}
}

// permissionListTable renders KEY / TYPE / NAME for the instance permission set.
func permissionListTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"KEY", "TYPE", "NAME"},
		RowFunc: func(item interface{}) []string {
			p := item.(client.Permission)
			return []string{p.Key, dash(p.Type), dash(p.Name)}
		},
	}
}

func permissionYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// permissionSortByKey gives stable output (the API returns an unordered map).
func permissionSortByKey(perms []client.Permission) {
	sort.Slice(perms, func(i, j int) bool { return perms[i].Key < perms[j].Key })
}

func newPermissionMineCmd() *cobra.Command {
	var (
		project string
		issue   string
		keys    string
	)
	cmd := &cobra.Command{
		Use:   "mine",
		Short: "Show which permissions you hold (optionally scoped to a project/issue)",
		Long: `Show the permissions the current user holds. Scope the check to a project
with --project or to an issue with --issue.

The permissions query parameter is required by Jira Cloud, so when --keys is
omitted a sensible common set is used (BROWSE_PROJECTS, CREATE_ISSUES,
EDIT_ISSUES, ASSIGN_ISSUES, TRANSITION_ISSUES, ADD_COMMENTS, DELETE_ISSUES,
ADMINISTER, ADMINISTER_PROJECTS).

Examples:
  jira permission mine
  jira permission mine --project ABC
  jira permission mine --issue ABC-123 --keys EDIT_ISSUES,DELETE_ISSUES -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			want := splitCSV(keys)
			if len(want) == 0 {
				want = permissionDefaultKeys
			}
			perms, err := jiraClient.MyPermissions(want, project, issue)
			if err != nil {
				return err
			}
			permissionSortByKey(perms)
			return render(perms, permissionHaveTable())
		},
	}
	cmd.Flags().StringVarP(&project, "project", "p", "", "Scope the check to a project key")
	cmd.Flags().StringVar(&issue, "issue", "", "Scope the check to an issue key")
	cmd.Flags().StringVar(&keys, "keys", "", "Comma-separated permission keys to check (default: a common set)")
	return cmd
}

func newPermissionPermittedProjectsCmd() *cobra.Command {
	var keys string
	cmd := &cobra.Command{
		Use:   "permitted-projects",
		Short: "List project ids where you hold all of the given permissions",
		Long: `Return the ids of the projects in which the current user holds ALL of the
given permissions. --keys is required.

Examples:
  jira permission permitted-projects --keys EDIT_ISSUES
  jira permission permitted-projects --keys CREATE_ISSUES,ADMINISTER_PROJECTS -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			want := splitCSV(keys)
			if len(want) == 0 {
				return fmt.Errorf("--keys is required (comma-separated permission keys)")
			}
			ids, err := jiraClient.PermittedProjects(want)
			if err != nil {
				return err
			}
			sort.Strings(ids)
			def := &output.TableDef{
				Headers: []string{"PROJECT_ID"},
				RowFunc: func(item interface{}) []string {
					return []string{item.(string)}
				},
			}
			return render(ids, def)
		},
	}
	cmd.Flags().StringVar(&keys, "keys", "", "Comma-separated permission keys (required)")
	return cmd
}

func newPermissionListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all permissions the instance defines",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			perms, err := jiraClient.ListPermissions()
			if err != nil {
				return err
			}
			permissionSortByKey(perms)
			return render(perms, permissionListTable())
		},
	}
}
