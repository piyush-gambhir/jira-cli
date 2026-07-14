package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/output"
)

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project",
		Aliases: []string{"projects"},
		Short:   "List, inspect, and manage projects",
		Long: `List, inspect, create, update, and delete projects.

Create/update/delete require project-admin permission (and, with OAuth, the
manage:jira-project scope — choose the "admin" or "all" scope preset at login).`,
	}
	cmd.AddCommand(
		newProjectListCmd(),
		newProjectGetCmd(),
		newProjectCreateCmd(),
		newProjectUpdateCmd(),
		newProjectDeleteCmd(),
		newProjectStatusesCmd(),
		newProjectArchiveCmd(),
		newProjectRestoreCmd(),
		newProjectRolesCmd(),
		newProjectRoleCmd(),
		newProjectCategoriesCmd(),
		newProjectCategoryCreateCmd(),
		newProjectCategoryDeleteCmd(),
	)
	return cmd
}

func projectTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"KEY", "NAME", "TYPE", "LEAD"},
		RowFunc: func(item interface{}) []string {
			p := item.(client.Project)
			return []string{p.Key, p.Name, dash(p.ProjectTypeKey), userName(p.Lead)}
		},
	}
}

func newProjectListCmd() *cobra.Command {
	var query string
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			projects, err := jiraClient.ListProjects(query, limit)
			if err != nil {
				return err
			}
			return render(projects, projectTable())
		},
	}
	cmd.Flags().StringVar(&query, "query", "", "Filter by key or name")
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Maximum projects to return")
	return cmd
}

func newProjectGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <idOrKey>",
		Short: "Get a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := jiraClient.GetProject(args[0])
			if err != nil {
				return err
			}
			return render(*p, projectTable())
		},
	}
}

func newProjectCreateCmd() *cobra.Command {
	var key, name, ptype, lead, template, description string
	var extraFields []string
	cmd := &cobra.Command{
		Use:         "create",
		Short:       "Create a project",
		Annotations: mutates,
		Args:        cobra.NoArgs,
		Long: `Create a project. --key and --name are required. Most Jira Cloud sites also
require a --template (projectTemplateKey) and a project lead.

Examples:
  jira project create --key MOB --name "Mobile" --lead @me \
    --template com.pyxis.greenhopper.jira:gh-simplified-agility-kanban`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if key == "" || name == "" {
				return fmt.Errorf("--key and --name are required")
			}
			body := map[string]any{"key": key, "name": name, "projectTypeKey": ptype}
			if template != "" {
				body["projectTemplateKey"] = template
			}
			if description != "" {
				body["description"] = description
			}
			if lead != "" {
				acct, err := jiraClient.ResolveUser(lead)
				if err != nil {
					return err
				}
				body["leadAccountId"] = acct
			}
			for _, kv := range extraFields {
				k, v, ok := strings.Cut(kv, "=")
				if !ok {
					return fmt.Errorf("invalid --field %q (expected name=value)", kv)
				}
				body[strings.TrimSpace(k)] = v
			}
			out, err := jiraClient.CreateProject(body)
			if err != nil {
				return err
			}
			info("Created project %s", key)
			return render(out, nil)
		},
	}
	cmd.Flags().StringVar(&key, "key", "", "Project key, e.g. MOB (required)")
	cmd.Flags().StringVar(&name, "name", "", "Project name (required)")
	cmd.Flags().StringVar(&ptype, "type", "software", "Project type key: software, service_desk, business")
	cmd.Flags().StringVar(&lead, "lead", "@me", "Project lead (email, name, @me, id:<accountId>)")
	cmd.Flags().StringVar(&template, "template", "", "Project template key (projectTemplateKey)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Project description")
	cmd.Flags().StringArrayVar(&extraFields, "field", nil, "Extra field as name=value (repeatable)")
	return cmd
}

func newProjectUpdateCmd() *cobra.Command {
	var name, lead, description string
	var extraFields []string
	cmd := &cobra.Command{
		Use:         "update <idOrKey>",
		Short:       "Update a project's fields",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]any{}
			if name != "" {
				body["name"] = name
			}
			if description != "" {
				body["description"] = description
			}
			if lead != "" {
				acct, err := jiraClient.ResolveUser(lead)
				if err != nil {
					return err
				}
				body["leadAccountId"] = acct
			}
			for _, kv := range extraFields {
				k, v, ok := strings.Cut(kv, "=")
				if !ok {
					return fmt.Errorf("invalid --field %q (expected name=value)", kv)
				}
				body[strings.TrimSpace(k)] = v
			}
			if len(body) == 0 {
				return fmt.Errorf("nothing to update; pass --name/--description/--lead/--field")
			}
			p, err := jiraClient.UpdateProject(args[0], body)
			if err != nil {
				return err
			}
			info("Updated project %s", args[0])
			return render(*p, projectTable())
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "New project name")
	cmd.Flags().StringVarP(&description, "description", "d", "", "New description")
	cmd.Flags().StringVar(&lead, "lead", "", "New project lead (email, name, @me, id:<accountId>)")
	cmd.Flags().StringArrayVar(&extraFields, "field", nil, "Extra field as name=value (repeatable)")
	return cmd
}

func newProjectDeleteCmd() *cobra.Command {
	var enableUndo, yes bool
	cmd := &cobra.Command{
		Use:         "delete <idOrKey>",
		Short:       "Delete a project",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				if noInputFlag {
					return fmt.Errorf("refusing to delete project %s without --yes", args[0])
				}
				ans, _ := prompt(fmt.Sprintf("Delete project %s? This removes all its issues. [y/N]: ", args[0]))
				if !strings.EqualFold(strings.TrimSpace(ans), "y") {
					info("Aborted.")
					return nil
				}
			}
			if err := jiraClient.DeleteProject(args[0], enableUndo); err != nil {
				return err
			}
			info("Deleted project %s", args[0])
			return nil
		},
	}
	cmd.Flags().BoolVar(&enableUndo, "enable-undo", false, "Move to the recycle bin instead of permanent delete")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")
	return cmd
}
