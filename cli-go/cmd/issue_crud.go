package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/adf"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/output"
)

func newIssueGetCmd() *cobra.Command {
	var fields, expand string
	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get an issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			if outFormat != "table" {
				raw, err := jiraClient.GetIssueRaw(key, fields, expand)
				if err != nil {
					return err
				}
				return render(raw, nil)
			}
			issue, err := jiraClient.GetIssue(key, fields, expand)
			if err != nil {
				return err
			}
			printIssueDetail(issue)
			return nil
		},
	}
	cmd.Flags().StringVar(&fields, "fields", "", "Comma-separated fields to request (default: all)")
	cmd.Flags().StringVar(&expand, "expand", "", "Comma-separated expand options (e.g. renderedFields,changelog)")
	return cmd
}

func printIssueDetail(i *client.Issue) {
	f := i.Fields
	fmt.Printf("%s  %s\n", i.Key, f.Summary)
	fmt.Printf("  Type:     %s\n", namedName(f.IssueType))
	fmt.Printf("  Status:   %s\n", statusName(f.Status))
	fmt.Printf("  Priority: %s\n", namedName(f.Priority))
	fmt.Printf("  Assignee: %s\n", userName(f.Assignee))
	fmt.Printf("  Reporter: %s\n", userName(f.Reporter))
	if f.Resolution != nil {
		fmt.Printf("  Resolution: %s\n", namedName(f.Resolution))
	}
	if len(f.Labels) > 0 {
		fmt.Printf("  Labels:   %s\n", strings.Join(f.Labels, ", "))
	}
	if f.DueDate != "" {
		fmt.Printf("  Due:      %s\n", f.DueDate)
	}
	fmt.Printf("  Updated:  %s\n", f.Updated)
	if desc := adf.ExtractText(f.Description); desc != "" {
		fmt.Printf("\n%s\n", desc)
	}
}

func newIssueCreateCmd() *cobra.Command {
	var (
		project     string
		issueType   string
		summary     string
		description string
		priority    string
		assignee    string
		labels      []string
		parent      string
		markdown    bool
		extraFields []string
	)
	cmd := &cobra.Command{
		Use:         "create",
		Short:       "Create an issue",
		Annotations: mutates,
		Long: `Create an issue. --project, --type and --summary are required.

The description is sent as ADF; use --markdown to interpret it as lightweight
markdown. Set arbitrary fields with repeated --field name=value.

Examples:
  jira issue create -p ABC --type Task --summary "Title" -d "Some details"
  jira issue create -p ABC --type Bug --summary "Broken" -d "**bold** details" --markdown -a me@corp.com -l urgent`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if project == "" || issueType == "" || summary == "" {
				return fmt.Errorf("--project, --type and --summary are required")
			}
			fields := map[string]any{
				"project":   map[string]string{"key": project},
				"issuetype": map[string]string{"name": issueType},
				"summary":   summary,
			}
			if description != "" {
				if markdown {
					fields["description"] = adf.FromMarkdown(description)
				} else {
					fields["description"] = adf.FromPlainText(description)
				}
			}
			if priority != "" {
				fields["priority"] = map[string]string{"name": priority}
			}
			if len(labels) > 0 {
				fields["labels"] = labels
			}
			if parent != "" {
				fields["parent"] = map[string]string{"key": parent}
			}
			if assignee != "" {
				acct, err := jiraClient.ResolveUser(assignee)
				if err != nil {
					return err
				}
				fields["assignee"] = map[string]string{"accountId": acct}
			}
			for _, kv := range extraFields {
				k, v, ok := strings.Cut(kv, "=")
				if !ok {
					return fmt.Errorf("invalid --field %q (expected name=value)", kv)
				}
				fields[strings.TrimSpace(k)] = v
			}
			created, err := jiraClient.CreateIssue(client.IssueUpdate{Fields: fields})
			if err != nil {
				return err
			}
			info("Created %s", created.Key)
			return render(created, issueKeyTable())
		},
	}
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project key (required)")
	cmd.Flags().StringVar(&issueType, "type", "", "Issue type name, e.g. Task/Bug/Story (required)")
	cmd.Flags().StringVar(&summary, "summary", "", "Issue summary (required)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Description text")
	cmd.Flags().StringVar(&priority, "priority", "", "Priority name")
	cmd.Flags().StringVarP(&assignee, "assignee", "a", "", "Assignee (email, name, @me, or id:<accountId>)")
	cmd.Flags().StringArrayVarP(&labels, "label", "l", nil, "Label (repeatable)")
	cmd.Flags().StringVar(&parent, "parent", "", "Parent issue key (for subtasks)")
	cmd.Flags().BoolVar(&markdown, "markdown", false, "Interpret --description as lightweight markdown")
	cmd.Flags().StringArrayVar(&extraFields, "field", nil, "Extra field as name=value (repeatable)")
	return cmd
}

func newIssueEditCmd() *cobra.Command {
	var (
		summary     string
		description string
		priority    string
		assignee    string
		addLabels   []string
		rmLabels    []string
		markdown    bool
		noNotify    bool
	)
	cmd := &cobra.Command{
		Use:         "edit <key>",
		Short:       "Edit an issue's fields",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			update := client.IssueUpdate{Fields: map[string]any{}, Update: map[string]any{}}
			if summary != "" {
				update.Fields["summary"] = summary
			}
			if description != "" {
				if markdown {
					update.Fields["description"] = adf.FromMarkdown(description)
				} else {
					update.Fields["description"] = adf.FromPlainText(description)
				}
			}
			if priority != "" {
				update.Fields["priority"] = map[string]string{"name": priority}
			}
			if assignee != "" {
				acct, err := jiraClient.ResolveUser(assignee)
				if err != nil {
					return err
				}
				update.Fields["assignee"] = map[string]string{"accountId": acct}
			}
			var labelOps []map[string]string
			for _, l := range addLabels {
				labelOps = append(labelOps, map[string]string{"add": l})
			}
			for _, l := range rmLabels {
				labelOps = append(labelOps, map[string]string{"remove": l})
			}
			if len(labelOps) > 0 {
				update.Update["labels"] = labelOps
			}
			if len(update.Fields) == 0 && len(update.Update) == 0 {
				return fmt.Errorf("nothing to edit; pass --summary/--description/--priority/--assignee/--add-label/--remove-label")
			}
			if err := jiraClient.EditIssue(key, update, !noNotify); err != nil {
				return err
			}
			info("Updated %s", key)
			return nil
		},
	}
	cmd.Flags().StringVar(&summary, "summary", "", "New summary")
	cmd.Flags().StringVarP(&description, "description", "d", "", "New description")
	cmd.Flags().StringVar(&priority, "priority", "", "New priority name")
	cmd.Flags().StringVarP(&assignee, "assignee", "a", "", "New assignee (email, name, @me, id:<accountId>)")
	cmd.Flags().StringArrayVar(&addLabels, "add-label", nil, "Add a label (repeatable)")
	cmd.Flags().StringArrayVar(&rmLabels, "remove-label", nil, "Remove a label (repeatable)")
	cmd.Flags().BoolVar(&markdown, "markdown", false, "Interpret --description as lightweight markdown")
	cmd.Flags().BoolVar(&noNotify, "no-notify", false, "Don't email watchers about the change")
	return cmd
}

func newIssueDeleteCmd() *cobra.Command {
	var deleteSubtasks, yes bool
	cmd := &cobra.Command{
		Use:         "delete <key>",
		Short:       "Delete an issue",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			if !yes {
				if noInputFlag {
					return fmt.Errorf("refusing to delete %s without --yes (or run interactively)", key)
				}
				ans, _ := prompt(fmt.Sprintf("Delete %s? This cannot be undone. [y/N]: ", key))
				if !strings.EqualFold(strings.TrimSpace(ans), "y") {
					info("Aborted.")
					return nil
				}
			}
			if err := jiraClient.DeleteIssue(key, deleteSubtasks); err != nil {
				return err
			}
			info("Deleted %s", key)
			return nil
		},
	}
	cmd.Flags().BoolVar(&deleteSubtasks, "delete-subtasks", false, "Also delete the issue's subtasks")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")
	return cmd
}

// issueKeyTable renders the {id,key,self} returned by create.
func issueKeyTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"KEY", "ID", "URL"},
		RowFunc: func(item interface{}) []string {
			i := item.(*client.Issue)
			return []string{i.Key, i.ID, i.Self}
		},
	}
}
