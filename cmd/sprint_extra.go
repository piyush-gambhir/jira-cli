package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func newSprintStartCmd() *cobra.Command {
	var start, end string
	cmd := &cobra.Command{
		Use:         "start <sprintId>",
		Short:       "Start a sprint (set state to active)",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		Long: `Start a sprint by moving it to the active state.

Jira requires a start and end date to activate a sprint; pass --start/--end
(ISO 8601). If omitted, only state:active is sent and Jira may reject the
request — the error is surfaced verbatim in that case.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("sprint id must be a number: %w", err)
			}
			body := map[string]any{"state": "active"}
			if start != "" {
				body["startDate"] = start
			}
			if end != "" {
				body["endDate"] = end
			}
			s, err := jiraClient.SprintPartialUpdate(id, body)
			if err != nil {
				return err
			}
			info("Started sprint %d (%s)", s.ID, s.Name)
			return render(*s, sprintTable())
		},
	}
	cmd.Flags().StringVar(&start, "start", "", "Start date (ISO 8601)")
	cmd.Flags().StringVar(&end, "end", "", "End date (ISO 8601)")
	return cmd
}

func newSprintCloseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "close <sprintId>",
		Aliases:     []string{"complete"},
		Short:       "Close a sprint (set state to closed)",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("sprint id must be a number: %w", err)
			}
			s, err := jiraClient.SprintPartialUpdate(id, map[string]any{"state": "closed"})
			if err != nil {
				return err
			}
			info("Closed sprint %d (%s)", s.ID, s.Name)
			return render(*s, sprintTable())
		},
	}
	return cmd
}

func newSprintUpdateCmd() *cobra.Command {
	var name, goal, start, end, state string
	cmd := &cobra.Command{
		Use:         "update <sprintId>",
		Short:       "Update a sprint's fields",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		Long: `Partially update a sprint. Only the flags you supply are changed.

Examples:
  jira sprint update 42 --name "Sprint 7" --goal "Ship v2"
  jira sprint update 42 --start 2026-06-14T09:00:00.000Z --end 2026-06-28T17:00:00.000Z
  jira sprint update 42 --state active`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("sprint id must be a number: %w", err)
			}
			body := map[string]any{}
			if name != "" {
				body["name"] = name
			}
			if goal != "" {
				body["goal"] = goal
			}
			if start != "" {
				body["startDate"] = start
			}
			if end != "" {
				body["endDate"] = end
			}
			if state != "" {
				body["state"] = state
			}
			if len(body) == 0 {
				return fmt.Errorf("nothing to update; pass --name/--goal/--start/--end/--state")
			}
			s, err := jiraClient.SprintPartialUpdate(id, body)
			if err != nil {
				return err
			}
			info("Updated sprint %d (%s)", s.ID, s.Name)
			return render(*s, sprintTable())
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "New sprint name")
	cmd.Flags().StringVar(&goal, "goal", "", "New sprint goal")
	cmd.Flags().StringVar(&start, "start", "", "New start date (ISO 8601)")
	cmd.Flags().StringVar(&end, "end", "", "New end date (ISO 8601)")
	cmd.Flags().StringVar(&state, "state", "", "New state: future, active, closed")
	return cmd
}

func newSprintDeleteCmd() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:         "delete <sprintId>",
		Short:       "Delete a sprint",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("sprint id must be a number: %w", err)
			}
			if !yes {
				if noInputFlag {
					return fmt.Errorf("refusing to delete sprint %d without --yes (or run interactively)", id)
				}
				ans, _ := prompt(fmt.Sprintf("Delete sprint %d? This cannot be undone. [y/N]: ", id))
				if !strings.EqualFold(strings.TrimSpace(ans), "y") {
					info("Aborted.")
					return nil
				}
			}
			if err := jiraClient.DeleteSprint(id); err != nil {
				return err
			}
			info("Deleted sprint %d", id)
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")
	return cmd
}

func newSprintAddIssuesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "add-issues <sprintId> <issueKey...>",
		Short:       "Move issues into a sprint (max 50)",
		Annotations: mutates,
		Args:        cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("sprint id must be a number: %w", err)
			}
			keys := args[1:]
			if len(keys) > 50 {
				return fmt.Errorf("at most 50 issues can be moved at once (got %d)", len(keys))
			}
			if err := jiraClient.MoveIssuesToSprint(id, keys); err != nil {
				return err
			}
			info("Moved %d issue(s) to sprint %d", len(keys), id)
			return nil
		},
	}
	return cmd
}
