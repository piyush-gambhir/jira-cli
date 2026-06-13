package cmd

import (
	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/internal/client"
	"github.com/piyush-gambhir/jira-cli/internal/output"
)

// ---------------------------------------------------------------------------
// issuetype
// ---------------------------------------------------------------------------

func newIssueTypeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "issuetype",
		Aliases: []string{"issuetypes"},
		Short:   "List and inspect issue types",
	}
	cmd.AddCommand(
		newIssueTypeListCmd(),
		newIssueTypeGetCmd(),
	)
	return cmd
}

func issueTypeTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"ID", "NAME", "SUBTASK", "DESCRIPTION"},
		RowFunc: func(item interface{}) []string {
			t := item.(client.IssueType)
			sub := "no"
			if t.Subtask {
				sub = "yes"
			}
			return []string{t.ID, t.Name, sub, dash(truncate(t.Description, 50))}
		},
	}
}

func newIssueTypeListCmd() *cobra.Command {
	var project string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issue types (optionally for a project)",
		Long: `List issue types. With --project (a numeric project id) only the issue
types associated with that project are returned.

Examples:
  jira issuetype list
  jira issuetype list --project 10001 -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				types []client.IssueType
				err   error
			)
			if project != "" {
				types, err = jiraClient.ListIssueTypesForProject(project)
			} else {
				types, err = jiraClient.ListIssueTypes()
			}
			if err != nil {
				return err
			}
			return render(types, issueTypeTable())
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "Filter by project id (numeric)")
	return cmd
}

func newIssueTypeGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get an issue type",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := jiraClient.GetIssueType(args[0])
			if err != nil {
				return err
			}
			return render(*t, issueTypeTable())
		},
	}
}

// ---------------------------------------------------------------------------
// priority
// ---------------------------------------------------------------------------

func newPriorityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "priority",
		Aliases: []string{"priorities"},
		Short:   "List and inspect issue priorities",
	}
	cmd.AddCommand(
		newPriorityListCmd(),
		newPriorityGetCmd(),
	)
	return cmd
}

func priorityTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"ID", "NAME", "DEFAULT", "DESCRIPTION"},
		RowFunc: func(item interface{}) []string {
			p := item.(client.Priority)
			def := "no"
			if p.IsDefault {
				def = "yes"
			}
			return []string{p.ID, p.Name, def, dash(truncate(p.Description, 50))}
		},
	}
}

func newPriorityListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issue priorities",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			priorities, err := jiraClient.ListPriorities()
			if err != nil {
				return err
			}
			return render(priorities, priorityTable())
		},
	}
	return cmd
}

func newPriorityGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a priority",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := jiraClient.GetPriority(args[0])
			if err != nil {
				return err
			}
			return render(*p, priorityTable())
		},
	}
}

// ---------------------------------------------------------------------------
// resolution
// ---------------------------------------------------------------------------

func newResolutionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "resolution",
		Aliases: []string{"resolutions"},
		Short:   "List and inspect issue resolutions",
	}
	cmd.AddCommand(
		newResolutionListCmd(),
		newResolutionGetCmd(),
	)
	return cmd
}

func resolutionTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"ID", "NAME", "DESCRIPTION"},
		RowFunc: func(item interface{}) []string {
			r := item.(client.Resolution)
			return []string{r.ID, r.Name, dash(truncate(r.Description, 60))}
		},
	}
}

func newResolutionListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issue resolutions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolutions, err := jiraClient.ListResolutions()
			if err != nil {
				return err
			}
			return render(resolutions, resolutionTable())
		},
	}
	return cmd
}

func newResolutionGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a resolution",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := jiraClient.GetResolution(args[0])
			if err != nil {
				return err
			}
			return render(*r, resolutionTable())
		},
	}
}

// ---------------------------------------------------------------------------
// label
// ---------------------------------------------------------------------------

func newLabelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "label",
		Aliases: []string{"labels"},
		Short:   "List available issue labels",
	}
	cmd.AddCommand(
		newLabelListCmd(),
	)
	return cmd
}

// labelRow wraps a single label string so it can carry through the TableDef
// RowFunc (which receives one slice element at a time).
type labelRow struct {
	Label string `json:"label"`
}

func labelTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"LABEL"},
		RowFunc: func(item interface{}) []string {
			l := item.(labelRow)
			return []string{l.Label}
		},
	}
}

func newLabelListCmd() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available issue labels",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			labels, err := jiraClient.ListLabels(limit)
			if err != nil {
				return err
			}
			rows := make([]labelRow, 0, len(labels))
			for _, l := range labels {
				rows = append(rows, labelRow{Label: l})
			}
			return render(rows, labelTable())
		},
	}
	cmd.Flags().IntVarP(&limit, "limit", "n", 100, "Maximum labels to return")
	return cmd
}
