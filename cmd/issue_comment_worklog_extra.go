package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/internal/adf"
	"github.com/piyush-gambhir/jira-cli/internal/client"
	"github.com/piyush-gambhir/jira-cli/internal/output"
)

// commentTable renders a single comment (id/author/created/body).
func commentTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"ID", "AUTHOR", "CREATED", "BODY"},
		RowFunc: func(item interface{}) []string {
			c := item.(client.Comment)
			return []string{c.ID, userName(c.Author), dash(c.Created), truncate(adf.ExtractText(c.Body), 60)}
		},
	}
}

// worklogTable renders a single worklog (id/author/time/started).
func worklogTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"ID", "AUTHOR", "TIME", "STARTED", "COMMENT"},
		RowFunc: func(item interface{}) []string {
			w := item.(client.Worklog)
			return []string{w.ID, userName(w.Author), dash(w.TimeSpent), dash(w.Started), truncate(adf.ExtractText(w.Comment), 40)}
		},
	}
}

func newIssueCommentGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment-get <key> <commentId>",
		Short: "Get a single comment on an issue",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := jiraClient.GetComment(args[0], args[1])
			if err != nil {
				return err
			}
			return render(*c, commentTable())
		},
	}
	return cmd
}

func newIssueCommentEditCmd() *cobra.Command {
	var body string
	var markdown bool
	cmd := &cobra.Command{
		Use:         "comment-edit <key> <commentId>",
		Short:       "Edit a comment on an issue",
		Annotations: mutates,
		Args:        cobra.ExactArgs(2),
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
			c, err := jiraClient.UpdateComment(args[0], args[1], doc)
			if err != nil {
				return err
			}
			info("Updated comment %s on %s", c.ID, args[0])
			return render(*c, commentTable())
		},
	}
	cmd.Flags().StringVarP(&body, "body", "b", "", "New comment body (required)")
	cmd.Flags().BoolVar(&markdown, "markdown", false, "Interpret --body as lightweight markdown")
	return cmd
}

func newIssueCommentDeleteCmd() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:         "comment-delete <key> <commentId>",
		Short:       "Delete a comment from an issue",
		Annotations: mutates,
		Args:        cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, id := args[0], args[1]
			if !yes {
				if noInputFlag {
					return fmt.Errorf("refusing to delete comment %s without --yes (or run interactively)", id)
				}
				ans, _ := prompt(fmt.Sprintf("Delete comment %s on %s? This cannot be undone. [y/N]: ", id, key))
				if !strings.EqualFold(strings.TrimSpace(ans), "y") {
					info("Aborted.")
					return nil
				}
			}
			if err := jiraClient.DeleteComment(key, id); err != nil {
				return err
			}
			info("Deleted comment %s on %s", id, key)
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")
	return cmd
}

func newIssueWorklogGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "worklog-get <key> <worklogId>",
		Short: "Get a single worklog entry on an issue",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			w, err := jiraClient.GetWorklog(args[0], args[1])
			if err != nil {
				return err
			}
			return render(*w, worklogTable())
		},
	}
	return cmd
}

func newIssueWorklogEditCmd() *cobra.Command {
	var timeSpent, comment, started string
	var markdown bool
	cmd := &cobra.Command{
		Use:         "worklog-edit <key> <worklogId>",
		Short:       "Edit a worklog entry on an issue",
		Annotations: mutates,
		Long: `Edit a worklog entry. At least one of --time/--comment/--started is required.

Examples:
  jira issue worklog-edit ABC-123 10001 --time "1h 30m"
  jira issue worklog-edit ABC-123 10001 --comment "revised" --markdown`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]any{}
			if timeSpent != "" {
				body["timeSpent"] = timeSpent
			}
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
			if len(body) == 0 {
				return fmt.Errorf("nothing to edit; pass --time/--comment/--started")
			}
			w, err := jiraClient.UpdateWorklog(args[0], args[1], body)
			if err != nil {
				return err
			}
			info("Updated worklog %s on %s", w.ID, args[0])
			return render(*w, worklogTable())
		},
	}
	cmd.Flags().StringVar(&timeSpent, "time", "", `New time spent, e.g. "3h 30m"`)
	cmd.Flags().StringVar(&comment, "comment", "", "New worklog comment")
	cmd.Flags().StringVar(&started, "started", "", "New start time (ISO 8601, e.g. 2026-06-13T09:00:00.000+0000)")
	cmd.Flags().BoolVar(&markdown, "markdown", false, "Interpret --comment as lightweight markdown")
	return cmd
}

func newIssueWorklogDeleteCmd() *cobra.Command {
	var adjust, newEstimate, increaseBy string
	var yes bool
	cmd := &cobra.Command{
		Use:         "worklog-delete <key> <worklogId>",
		Short:       "Delete a worklog entry from an issue",
		Annotations: mutates,
		Long: `Delete a worklog entry. Use --adjust to control the remaining estimate:
  auto    (default) reduce the estimate by the deleted worklog's time
  leave   leave the estimate unchanged
  new     set the estimate to --new-estimate
  manual  decrease the estimate by --increase-by`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, id := args[0], args[1]
			if !yes {
				if noInputFlag {
					return fmt.Errorf("refusing to delete worklog %s without --yes (or run interactively)", id)
				}
				ans, _ := prompt(fmt.Sprintf("Delete worklog %s on %s? This cannot be undone. [y/N]: ", id, key))
				if !strings.EqualFold(strings.TrimSpace(ans), "y") {
					info("Aborted.")
					return nil
				}
			}
			q := url.Values{}
			if adjust != "" {
				q.Set("adjustEstimate", adjust)
			}
			if newEstimate != "" {
				q.Set("newEstimate", newEstimate)
			}
			if increaseBy != "" {
				q.Set("increaseBy", increaseBy)
			}
			if err := jiraClient.DeleteWorklog(key, id, q); err != nil {
				return err
			}
			info("Deleted worklog %s on %s", id, key)
			return nil
		},
	}
	cmd.Flags().StringVar(&adjust, "adjust", "", "Estimate adjustment: auto, leave, new, manual")
	cmd.Flags().StringVar(&newEstimate, "new-estimate", "", "New remaining estimate (with --adjust new)")
	cmd.Flags().StringVar(&increaseBy, "increase-by", "", "Amount to decrease the estimate by (with --adjust manual)")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")
	return cmd
}
