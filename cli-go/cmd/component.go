package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/output"
)

func newComponentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "component",
		Aliases: []string{"components"},
		Short:   "List, inspect, and manage project components",
		Long: `List, inspect, create, update, and delete project components.

Create/update/delete require project-admin permission for the component's project.`,
	}
	cmd.AddCommand(
		newComponentListCmd(),
		newComponentGetCmd(),
		newComponentCreateCmd(),
		newComponentUpdateCmd(),
		newComponentDeleteCmd(),
		newComponentCountCmd(),
	)
	return cmd
}

func componentTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"ID", "NAME", "LEAD", "ASSIGNEE TYPE", "PROJECT"},
		RowFunc: func(item interface{}) []string {
			c := item.(client.Component)
			return []string{
				dash(c.ID),
				dash(c.Name),
				userName(c.Lead),
				dash(c.AssigneeType),
				dash(c.Project),
			}
		},
	}
}

func newComponentListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <projectIdOrKey>",
		Short: "List a project's components",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			components, err := jiraClient.ListComponents(args[0])
			if err != nil {
				return err
			}
			return render(components, componentTable())
		},
	}
}

func newComponentGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a component",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := jiraClient.GetComponent(args[0])
			if err != nil {
				return err
			}
			return render(*c, componentTable())
		},
	}
}

func newComponentCreateCmd() *cobra.Command {
	var project, name, description, lead, assigneeType string
	cmd := &cobra.Command{
		Use:         "create --project <key> --name <name>",
		Short:       "Create a component",
		Annotations: mutates,
		Args:        cobra.NoArgs,
		Long: `Create a project component. --project and --name are required.

Examples:
  jira component create --project ABC --name Backend
  jira component create --project ABC --name API -d "Public API" \
    --lead @me --assignee-type COMPONENT_LEAD`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if project == "" || name == "" {
				return fmt.Errorf("--project and --name are required")
			}
			body := map[string]any{"project": project, "name": name}
			if description != "" {
				body["description"] = description
			}
			if assigneeType != "" {
				body["assigneeType"] = assigneeType
			}
			if lead != "" {
				acct, err := jiraClient.ResolveUser(lead)
				if err != nil {
					return err
				}
				body["leadAccountId"] = acct
			}
			c, err := jiraClient.CreateComponent(body)
			if err != nil {
				return err
			}
			info("Created component %s (%s)", c.Name, c.ID)
			return render(*c, componentTable())
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "Project key (required)")
	cmd.Flags().StringVar(&name, "name", "", "Component name (required)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Component description")
	cmd.Flags().StringVar(&lead, "lead", "", "Component lead (email, name, @me, id:<accountId>)")
	cmd.Flags().StringVar(&assigneeType, "assignee-type", "", "Default assignee: PROJECT_DEFAULT, COMPONENT_LEAD, PROJECT_LEAD, UNASSIGNED")
	return cmd
}

func newComponentUpdateCmd() *cobra.Command {
	var name, description, lead, assigneeType string
	cmd := &cobra.Command{
		Use:         "update <id>",
		Short:       "Update a component's fields",
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
			if assigneeType != "" {
				body["assigneeType"] = assigneeType
			}
			if lead != "" {
				acct, err := jiraClient.ResolveUser(lead)
				if err != nil {
					return err
				}
				body["leadAccountId"] = acct
			}
			if len(body) == 0 {
				return fmt.Errorf("nothing to update; pass --name/--description/--lead/--assignee-type")
			}
			c, err := jiraClient.UpdateComponent(args[0], body)
			if err != nil {
				return err
			}
			info("Updated component %s", args[0])
			return render(*c, componentTable())
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "New component name")
	cmd.Flags().StringVarP(&description, "description", "d", "", "New description")
	cmd.Flags().StringVar(&lead, "lead", "", "New component lead (email, name, @me, id:<accountId>)")
	cmd.Flags().StringVar(&assigneeType, "assignee-type", "", "Default assignee: PROJECT_DEFAULT, COMPONENT_LEAD, PROJECT_LEAD, UNASSIGNED")
	return cmd
}

func newComponentDeleteCmd() *cobra.Command {
	var moveIssuesTo string
	var yes bool
	cmd := &cobra.Command{
		Use:         "delete <id>",
		Short:       "Delete a component",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		Long: `Delete a component. Issues using it lose the component unless --move-issues-to
reassigns them to another component id.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				if noInputFlag {
					return fmt.Errorf("refusing to delete component %s without --yes", args[0])
				}
				ans, _ := prompt(fmt.Sprintf("Delete component %s? [y/N]: ", args[0]))
				if !strings.EqualFold(strings.TrimSpace(ans), "y") {
					info("Aborted.")
					return nil
				}
			}
			if err := jiraClient.DeleteComponent(args[0], moveIssuesTo); err != nil {
				return err
			}
			info("Deleted component %s", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&moveIssuesTo, "move-issues-to", "", "Reassign affected issues to this component id")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")
	return cmd
}

func newComponentCountCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "count <id>",
		Short: "Count issues using a component",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := jiraClient.ComponentIssueCount(args[0])
			if err != nil {
				return err
			}
			return render(map[string]int{"issueCount": n}, nil)
		},
	}
}
