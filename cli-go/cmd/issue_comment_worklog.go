package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/adf"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/output"
)

func newIssueCommentCmd() *cobra.Command {
	var body string
	var markdown bool
	cmd := &cobra.Command{
		Use:         "comment <key>",
		Short:       "Add a comment to an issue",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if body == "" {
				return fmt.Errorf("--body is required")
			}
			var doc any
			if markdown {
				doc = adf.FromMarkdown(body)
			} else {
				doc = adf.FromPlainText(body)
			}
			c, err := jiraClient.AddComment(args[0], doc, nil)
			if err != nil {
				return err
			}
			info("Added comment %s to %s", c.ID, args[0])
			return nil
		},
	}
	cmd.Flags().StringVarP(&body, "body", "b", "", "Comment body (required)")
	cmd.Flags().BoolVar(&markdown, "markdown", false, "Interpret --body as lightweight markdown")
	return cmd
}

func newIssueCommentsCmd() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "comments <key>",
		Short: "List an issue's comments",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			comments, err := jiraClient.ListComments(args[0], limit)
			if err != nil {
				return err
			}
			def := &output.TableDef{
				Headers: []string{"ID", "AUTHOR", "CREATED", "BODY"},
				RowFunc: func(item interface{}) []string {
					c := item.(client.Comment)
					return []string{c.ID, userName(c.Author), c.Created, truncate(adf.ExtractText(c.Body), 60)}
				},
			}
			return render(comments, def)
		},
	}
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Maximum comments to return")
	return cmd
}

func newIssueWorklogCmd() *cobra.Command {
	var timeSpent, comment, started, adjust, newEstimate string
	var markdown bool
	cmd := &cobra.Command{
		Use:         "worklog <key>",
		Short:       "Log work on an issue",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		Long: `Log time against an issue. --time is required (e.g. "3h 30m" or "90m").

Examples:
  jira issue worklog ABC-123 --time "2h"
  jira issue worklog ABC-123 --time 45m --comment "code review" --started "2026-06-13T09:00:00.000+0000"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if timeSpent == "" {
				return fmt.Errorf("--time is required")
			}
			body := map[string]any{"timeSpent": timeSpent}
			if started != "" {
				body["started"] = started
			}
			if comment != "" {
				if markdown {
					body["comment"] = adf.FromMarkdown(comment)
				} else {
					body["comment"] = adf.FromPlainText(comment)
				}
			}
			w, err := jiraClient.AddWorklog(args[0], body, adjust, newEstimate)
			if err != nil {
				return err
			}
			info("Logged %s on %s (worklog %s)", timeSpent, args[0], w.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&timeSpent, "time", "", `Time spent, e.g. "3h 30m" (required)`)
	cmd.Flags().StringVar(&comment, "comment", "", "Worklog comment")
	cmd.Flags().StringVar(&started, "started", "", "Start time (ISO 8601, e.g. 2026-06-13T09:00:00.000+0000)")
	cmd.Flags().StringVar(&adjust, "adjust", "", "Estimate adjustment: auto, leave, new, manual")
	cmd.Flags().StringVar(&newEstimate, "new-estimate", "", "New remaining estimate (with --adjust new)")
	cmd.Flags().BoolVar(&markdown, "markdown", false, "Interpret --comment as lightweight markdown")
	return cmd
}

func newIssueWorklogsCmd() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "worklogs <key>",
		Short: "List an issue's worklog entries",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			worklogs, err := jiraClient.ListWorklogs(args[0], limit)
			if err != nil {
				return err
			}
			def := &output.TableDef{
				Headers: []string{"ID", "AUTHOR", "TIME", "STARTED"},
				RowFunc: func(item interface{}) []string {
					w := item.(client.Worklog)
					return []string{w.ID, userName(w.Author), dash(w.TimeSpent), dash(w.Started)}
				},
			}
			return render(worklogs, def)
		},
	}
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Maximum worklogs to return")
	return cmd
}
