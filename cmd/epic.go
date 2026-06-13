package cmd

import (
	"github.com/spf13/cobra"
)

func newEpicCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "epic",
		Aliases: []string{"epics"},
		Short:   "Inspect epics and their issues (Agile)",
	}
	cmd.AddCommand(newEpicGetCmd(), newEpicIssuesCmd())
	cmd.AddCommand(newEpicUpdateCmd(), newEpicAddIssuesCmd(), newEpicRemoveIssuesCmd())
	return cmd
}

func newEpicGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <epicKey>",
		Short: "Get an epic",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			e, err := jiraClient.GetEpic(args[0])
			if err != nil {
				return err
			}
			// Epic shape varies; render as raw JSON/YAML (table falls back to JSON).
			return render(e, nil)
		},
	}
}

func newEpicIssuesCmd() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "issues <epicKey>",
		Short: "List issues belonging to an epic",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			issues, err := jiraClient.EpicIssues(args[0], limit)
			if err != nil {
				return err
			}
			return render(issues, issueListTable())
		},
	}
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Maximum issues to return")
	return cmd
}
