package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/internal/client"
	"github.com/piyush-gambhir/jira-cli/internal/output"
)

// boardEpicTable renders epics listed under a board.
func boardEpicTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"ID", "KEY", "NAME", "DONE"},
		RowFunc: func(item interface{}) []string {
			e := item.(client.BoardEpic)
			return []string{strconv.Itoa(e.ID), dash(e.Key), dash(e.Name), fmt.Sprintf("%v", e.Done)}
		},
	}
}

// --- board subcommands (attach to parent "board") ---

func newBoardCreateCmd() *cobra.Command {
	var name, boardType, project string
	var filterID int
	cmd := &cobra.Command{
		Use:         "create",
		Short:       "Create an Agile board",
		Annotations: mutates,
		Args:        cobra.NoArgs,
		Long: `Create an Agile board backed by a saved filter. --name, --type and
--filter-id are required. Pass --project to locate the board in a project.

Examples:
  jira board create --name "Mobile Scrum" --type scrum --filter-id 10001
  jira board create --name "Ops" --type kanban --filter-id 10002 --project OPS`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" || boardType == "" || filterID == 0 {
				return fmt.Errorf("--name, --type and --filter-id are required")
			}
			body := map[string]any{
				"name":     name,
				"type":     boardType,
				"filterId": filterID,
			}
			if project != "" {
				body["location"] = map[string]any{"type": "project", "projectKeyOrId": project}
			}
			b, err := jiraClient.BoardCreate(body)
			if err != nil {
				return err
			}
			info("Created board %d (%s)", b.ID, b.Name)
			return render(*b, boardTable())
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Board name (required)")
	cmd.Flags().StringVar(&boardType, "type", "", "Board type: scrum, kanban, simple (required)")
	cmd.Flags().IntVar(&filterID, "filter-id", 0, "Saved filter id backing the board (required)")
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project key or id to locate the board in")
	return cmd
}

func newBoardDeleteCmd() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:         "delete <boardId>",
		Short:       "Delete a board",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("board id must be a number: %w", err)
			}
			if !yes {
				if noInputFlag {
					return fmt.Errorf("refusing to delete board %d without --yes", id)
				}
				ans, _ := prompt(fmt.Sprintf("Delete board %d? This cannot be undone. [y/N]: ", id))
				if !strings.EqualFold(strings.TrimSpace(ans), "y") {
					info("Aborted.")
					return nil
				}
			}
			if err := jiraClient.DeleteBoard(id); err != nil {
				return err
			}
			info("Deleted board %d", id)
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")
	return cmd
}

func newBoardConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config <boardId>",
		Short: "Show a board's configuration (columns, estimation, ranking)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("board id must be a number: %w", err)
			}
			cfg, err := jiraClient.BoardConfig(id)
			if err != nil {
				return err
			}
			// Configuration shape varies; render as raw JSON/YAML (table falls back to JSON).
			return render(cfg, nil)
		},
	}
}

func newBoardEpicsCmd() *cobra.Command {
	var done bool
	var limit int
	cmd := &cobra.Command{
		Use:   "epics <boardId>",
		Short: "List epics associated with a board",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("board id must be a number: %w", err)
			}
			epics, err := jiraClient.BoardEpics(id, done, limit)
			if err != nil {
				return err
			}
			return render(epics, boardEpicTable())
		},
	}
	cmd.Flags().BoolVar(&done, "done", false, "Only show completed epics")
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Maximum epics to return")
	return cmd
}

// --- epic subcommands (attach to parent "epic") ---

func newEpicUpdateCmd() *cobra.Command {
	var name, summary, color string
	var done, notDone bool
	cmd := &cobra.Command{
		Use:         "update <epicKey>",
		Short:       "Update an epic's name, summary, color, or done state",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]any{}
			if name != "" {
				body["name"] = name
			}
			if summary != "" {
				body["summary"] = summary
			}
			if color != "" {
				body["color"] = map[string]string{"key": color}
			}
			if done && notDone {
				return fmt.Errorf("--done and --not-done are mutually exclusive")
			}
			if done {
				body["done"] = true
			} else if notDone {
				body["done"] = false
			}
			if len(body) == 0 {
				return fmt.Errorf("nothing to update; pass --name/--summary/--color/--done/--not-done")
			}
			out, err := jiraClient.EpicUpdate(args[0], body)
			if err != nil {
				return err
			}
			info("Updated epic %s", args[0])
			return render(out, nil)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "New epic name")
	cmd.Flags().StringVar(&summary, "summary", "", "New epic summary")
	cmd.Flags().StringVar(&color, "color", "", "Epic color key, e.g. color_1 … color_9")
	cmd.Flags().BoolVar(&done, "done", false, "Mark the epic as done")
	cmd.Flags().BoolVar(&notDone, "not-done", false, "Mark the epic as not done")
	return cmd
}

func newEpicAddIssuesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "add-issues <epicKey> <issueKey...>",
		Short:       "Move issues into an epic",
		Annotations: mutates,
		Args:        cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			epicKey, issueKeys := args[0], args[1:]
			if err := jiraClient.EpicAddIssues(epicKey, issueKeys); err != nil {
				return err
			}
			info("Added %d issue(s) to epic %s", len(issueKeys), epicKey)
			return nil
		},
	}
	return cmd
}

func newEpicRemoveIssuesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "remove-issues <issueKey...>",
		Short:       "Remove issues from their epic",
		Annotations: mutates,
		Args:        cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := jiraClient.EpicRemoveIssues(args); err != nil {
				return err
			}
			info("Removed %d issue(s) from their epic", len(args))
			return nil
		},
	}
	return cmd
}

// --- backlog top-level group (attach to root) ---

func newBacklogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backlog",
		Short: "Manage the Agile backlog",
	}
	cmd.AddCommand(newBacklogMoveCmd())
	return cmd
}

func newBacklogMoveCmd() *cobra.Command {
	var board int
	cmd := &cobra.Command{
		Use:         "move <issueKey...>",
		Short:       "Move issues to the backlog",
		Annotations: mutates,
		Args:        cobra.MinimumNArgs(1),
		Long: `Move issues out of their sprint and into the backlog (max 50 per call).

Without --board this uses the board-agnostic endpoint (issues must already
belong to a scrum board). Pass --board to target a specific board's backlog.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			if board != 0 {
				err = jiraClient.MoveIssuesToBoardBacklog(board, args)
			} else {
				err = jiraClient.MoveIssuesToBacklog(args)
			}
			if err != nil {
				return err
			}
			info("Moved %d issue(s) to the backlog", len(args))
			return nil
		},
	}
	cmd.Flags().IntVarP(&board, "board", "b", 0, "Target a specific board's backlog")
	return cmd
}
