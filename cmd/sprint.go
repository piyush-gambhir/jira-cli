package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/internal/client"
	"github.com/piyush-gambhir/jira-cli/internal/output"
)

func newSprintCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sprint",
		Aliases: []string{"sprints"},
		Short:   "List, inspect, and create sprints (Agile)",
	}
	cmd.AddCommand(newSprintListCmd(), newSprintGetCmd(), newSprintCreateCmd(), newSprintIssuesCmd())
	cmd.AddCommand(newSprintStartCmd(), newSprintCloseCmd(), newSprintUpdateCmd(), newSprintDeleteCmd(), newSprintAddIssuesCmd())
	return cmd
}

func sprintTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"ID", "NAME", "STATE", "START", "END", "GOAL"},
		RowFunc: func(item interface{}) []string {
			s := item.(client.Sprint)
			return []string{strconv.Itoa(s.ID), s.Name, dash(s.State), dash(s.StartDate), dash(s.EndDate), truncate(s.Goal, 30)}
		},
	}
}

func newSprintListCmd() *cobra.Command {
	var board int
	var state string
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List sprints on a board",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if board == 0 {
				return fmt.Errorf("--board is required")
			}
			sprints, err := jiraClient.ListSprints(board, state, limit)
			if err != nil {
				return err
			}
			return render(sprints, sprintTable())
		},
	}
	cmd.Flags().IntVarP(&board, "board", "b", 0, "Board id (required)")
	cmd.Flags().StringVar(&state, "state", "", "Filter by state: future, active, closed (comma-separated)")
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Maximum sprints to return")
	return cmd
}

func newSprintGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <sprintId>",
		Short: "Get a sprint",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("sprint id must be a number: %w", err)
			}
			s, err := jiraClient.GetSprint(id)
			if err != nil {
				return err
			}
			return render(*s, sprintTable())
		},
	}
}

func newSprintCreateCmd() *cobra.Command {
	var board int
	var name, start, end, goal string
	cmd := &cobra.Command{
		Use:         "create",
		Short:       "Create a sprint on a board",
		Annotations: mutates,
		Args:        cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if board == 0 || name == "" {
				return fmt.Errorf("--board and --name are required")
			}
			s, err := jiraClient.CreateSprint(name, board, start, end, goal)
			if err != nil {
				return err
			}
			info("Created sprint %d (%s)", s.ID, s.Name)
			return render(*s, sprintTable())
		},
	}
	cmd.Flags().IntVarP(&board, "board", "b", 0, "Origin board id (required)")
	cmd.Flags().StringVar(&name, "name", "", "Sprint name (required)")
	cmd.Flags().StringVar(&start, "start", "", "Start date (ISO 8601)")
	cmd.Flags().StringVar(&end, "end", "", "End date (ISO 8601)")
	cmd.Flags().StringVar(&goal, "goal", "", "Sprint goal")
	return cmd
}

func newSprintIssuesCmd() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "issues <sprintId>",
		Short: "List issues in a sprint",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("sprint id must be a number: %w", err)
			}
			issues, err := jiraClient.SprintIssues(id, limit)
			if err != nil {
				return err
			}
			return render(issues, issueListTable())
		},
	}
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Maximum issues to return")
	return cmd
}
