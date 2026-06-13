package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/internal/client"
	"github.com/piyush-gambhir/jira-cli/internal/output"
)

func newBoardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "board",
		Aliases: []string{"boards"},
		Short:   "List boards and their issues (Agile)",
	}
	cmd.AddCommand(newBoardListCmd(), newBoardGetCmd(), newBoardIssuesCmd(), newBoardBacklogCmd())
	return cmd
}

func boardTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"ID", "NAME", "TYPE"},
		RowFunc: func(item interface{}) []string {
			b := item.(client.Board)
			return []string{strconv.Itoa(b.ID), b.Name, dash(b.Type)}
		},
	}
}

func newBoardListCmd() *cobra.Command {
	var project, boardType string
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Agile boards",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			boards, err := jiraClient.ListBoards(project, boardType, limit)
			if err != nil {
				return err
			}
			return render(boards, boardTable())
		},
	}
	cmd.Flags().StringVarP(&project, "project", "p", "", "Filter by project key or id")
	cmd.Flags().StringVar(&boardType, "type", "", "Filter by board type: scrum, kanban, simple")
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Maximum boards to return")
	return cmd
}

func newBoardGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <boardId>",
		Short: "Get a board",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("board id must be a number: %w", err)
			}
			b, err := jiraClient.GetBoard(id)
			if err != nil {
				return err
			}
			return render(*b, boardTable())
		},
	}
}

func newBoardIssuesCmd() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "issues <boardId>",
		Short: "List issues on a board",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("board id must be a number: %w", err)
			}
			issues, err := jiraClient.BoardIssues(id, limit)
			if err != nil {
				return err
			}
			return render(issues, issueListTable())
		},
	}
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Maximum issues to return")
	return cmd
}

func newBoardBacklogCmd() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "backlog <boardId>",
		Short: "List backlog issues for a board",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("board id must be a number: %w", err)
			}
			issues, err := jiraClient.BacklogIssues(id, limit)
			if err != nil {
				return err
			}
			return render(issues, issueListTable())
		},
	}
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Maximum issues to return")
	return cmd
}
