package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/internal/client"
	"github.com/piyush-gambhir/jira-cli/internal/output"
)

// projectStatusRow is a flattened ISSUETYPE/STATUS pair for tabular rendering.
type projectStatusRow struct {
	IssueType      string `json:"issueType"`
	Status         string `json:"status"`
	StatusCategory string `json:"statusCategory,omitempty"`
}

func projectStatusTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"ISSUETYPE", "STATUS", "CATEGORY"},
		RowFunc: func(item interface{}) []string {
			r := item.(projectStatusRow)
			return []string{r.IssueType, r.Status, dash(r.StatusCategory)}
		},
	}
}

func newProjectStatusesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "statuses <idOrKey>",
		Short: "List the statuses available to each issue type in a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			types, err := jiraClient.ProjectStatuses(args[0])
			if err != nil {
				return err
			}
			var rows []projectStatusRow
			for _, t := range types {
				if len(t.Statuses) == 0 {
					rows = append(rows, projectStatusRow{IssueType: t.Name})
					continue
				}
				for _, s := range t.Statuses {
					cat := ""
					if s.StatusCategory != nil {
						cat = s.StatusCategory.Name
					}
					rows = append(rows, projectStatusRow{IssueType: t.Name, Status: s.Name, StatusCategory: cat})
				}
			}
			return render(rows, projectStatusTable())
		},
	}
}

func newProjectArchiveCmd() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:         "archive <idOrKey>",
		Short:       "Archive a project",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				if noInputFlag {
					return fmt.Errorf("refusing to archive project %s without --yes", args[0])
				}
				ans, _ := prompt(fmt.Sprintf("Archive project %s? [y/N]: ", args[0]))
				if !strings.EqualFold(strings.TrimSpace(ans), "y") {
					info("Aborted.")
					return nil
				}
			}
			if err := jiraClient.ArchiveProject(args[0]); err != nil {
				return err
			}
			info("Archived project %s", args[0])
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")
	return cmd
}

func newProjectRestoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "restore <idOrKey>",
		Short:       "Restore an archived or trashed project",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := jiraClient.RestoreProject(args[0]); err != nil {
				return err
			}
			info("Restored project %s", args[0])
			return nil
		},
	}
	return cmd
}

// projectRoleRef is a ROLE/URL pair for tabular rendering of the roles map.
type projectRoleRef struct {
	Role string `json:"role"`
	URL  string `json:"url"`
}

func newProjectRolesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "roles <idOrKey>",
		Short: "List a project's roles",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			roles, err := jiraClient.ProjectRoles(args[0])
			if err != nil {
				return err
			}
			refs := make([]projectRoleRef, 0, len(roles))
			for name, u := range roles {
				refs = append(refs, projectRoleRef{Role: name, URL: u})
			}
			sort.Slice(refs, func(i, j int) bool { return refs[i].Role < refs[j].Role })
			def := &output.TableDef{
				Headers: []string{"ROLE", "URL"},
				RowFunc: func(item interface{}) []string {
					r := item.(projectRoleRef)
					return []string{r.Role, r.URL}
				},
			}
			return render(refs, def)
		},
	}
}

func newProjectRoleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "role <idOrKey> <roleId>",
		Short: "Show a project role and its members",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			role, err := jiraClient.GetProjectRole(args[0], args[1])
			if err != nil {
				return err
			}
			def := &output.TableDef{
				Headers: []string{"NAME", "TYPE"},
				RowFunc: func(item interface{}) []string {
					a := item.(client.ProjectRoleActor)
					name := a.DisplayName
					if name == "" {
						name = a.Name
					}
					return []string{dash(name), dash(a.Type)}
				},
			}
			return render(role.Actors, def)
		},
	}
}

func projectCategoryTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"ID", "NAME", "DESCRIPTION"},
		RowFunc: func(item interface{}) []string {
			c := item.(client.ProjectCategory)
			return []string{c.ID, c.Name, dash(c.Description)}
		},
	}
}

func newProjectCategoriesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "categories",
		Short: "List project categories",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cats, err := jiraClient.ProjectCategories()
			if err != nil {
				return err
			}
			return render(cats, projectCategoryTable())
		},
	}
}

func newProjectCategoryCreateCmd() *cobra.Command {
	var name, description string
	cmd := &cobra.Command{
		Use:         "category-create",
		Short:       "Create a project category",
		Annotations: mutates,
		Args:        cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			body := map[string]any{"name": name}
			if description != "" {
				body["description"] = description
			}
			cat, err := jiraClient.CreateProjectCategory(body)
			if err != nil {
				return err
			}
			info("Created project category %s (%s)", cat.Name, cat.ID)
			return render(*cat, projectCategoryTable())
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Category name (required)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Category description")
	return cmd
}

func newProjectCategoryDeleteCmd() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:         "category-delete <id>",
		Short:       "Delete a project category",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := strconv.Atoi(args[0]); err != nil {
				return fmt.Errorf("category id must be a number: %w", err)
			}
			if !yes {
				if noInputFlag {
					return fmt.Errorf("refusing to delete category %s without --yes", args[0])
				}
				ans, _ := prompt(fmt.Sprintf("Delete project category %s? [y/N]: ", args[0]))
				if !strings.EqualFold(strings.TrimSpace(ans), "y") {
					info("Aborted.")
					return nil
				}
			}
			if err := jiraClient.DeleteProjectCategory(args[0]); err != nil {
				return err
			}
			info("Deleted project category %s", args[0])
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")
	return cmd
}
