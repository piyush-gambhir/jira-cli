package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/output"
)

// linksMiscConfirm prompts for a destructive action unless --yes/--no-input is
// set, mirroring the issue-delete flow. It returns true when the action should
// proceed.
func linksMiscConfirm(yes bool, label string) bool {
	if yes {
		return true
	}
	if noInputFlag {
		return false
	}
	ans, _ := prompt(fmt.Sprintf("%s This cannot be undone. [y/N]: ", label))
	return strings.EqualFold(strings.TrimSpace(ans), "y")
}

// remoteLinkTable renders the id/title/url of a remote (web) link.
func remoteLinkTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"ID", "TITLE", "URL"},
		RowFunc: func(item interface{}) []string {
			r := item.(client.RemoteLink)
			return []string{fmt.Sprintf("%d", r.ID), dash(r.Object.Title), dash(r.Object.URL)}
		},
	}
}

func newIssueLinkDeleteCmd() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:         "link-delete <linkId>",
		Short:       "Delete an issue link by id",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			if !linksMiscConfirm(yes, fmt.Sprintf("Delete issue link %s?", id)) {
				if noInputFlag && !yes {
					return fmt.Errorf("refusing to delete link %s without --yes (or run interactively)", id)
				}
				info("Aborted.")
				return nil
			}
			if err := jiraClient.DeleteLink(id); err != nil {
				return err
			}
			info("Deleted issue link %s", id)
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")
	return cmd
}

func newIssueAttachmentDeleteCmd() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:         "attachment-delete <attachmentId>",
		Short:       "Delete an attachment by id",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			if !linksMiscConfirm(yes, fmt.Sprintf("Delete attachment %s?", id)) {
				if noInputFlag && !yes {
					return fmt.Errorf("refusing to delete attachment %s without --yes (or run interactively)", id)
				}
				info("Aborted.")
				return nil
			}
			if err := jiraClient.DeleteAttachment(id); err != nil {
				return err
			}
			info("Deleted attachment %s", id)
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")
	return cmd
}

func newIssueAttachmentMetaCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "attachment-meta <attachmentId>",
		Short: "Show metadata for a single attachment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := jiraClient.GetAttachmentMeta(args[0])
			if err != nil {
				return err
			}
			return render(*a, attachmentTable())
		},
	}
}

func newIssueRemoteLinksCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remote-links <key>",
		Short: "List an issue's remote (web) links",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			links, err := jiraClient.RemoteLinks(args[0])
			if err != nil {
				return err
			}
			return render(links, remoteLinkTable())
		},
	}
}

func newIssueRemoteLinkAddCmd() *cobra.Command {
	var rlURL, title, summary string
	cmd := &cobra.Command{
		Use:         "remote-link-add <key>",
		Short:       "Add a remote (web) link to an issue",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		Long: `Add a remote (web) link to an issue. --url and --title are required.

Examples:
  jira issue remote-link-add ABC-123 --url https://example.com --title "Design doc"
  jira issue remote-link-add ABC-123 --url https://example.com --title Spec --summary "the spec"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if rlURL == "" || title == "" {
				return fmt.Errorf("--url and --title are required")
			}
			object := map[string]any{"url": rlURL, "title": title}
			if summary != "" {
				object["summary"] = summary
			}
			body := map[string]any{"object": object}
			out, err := jiraClient.AddRemoteLink(args[0], body)
			if err != nil {
				return err
			}
			info("Added remote link to %s", args[0])
			return render(out, nil)
		},
	}
	cmd.Flags().StringVar(&rlURL, "url", "", "Link URL (required)")
	cmd.Flags().StringVar(&title, "title", "", "Link title (required)")
	cmd.Flags().StringVar(&summary, "summary", "", "Optional link summary")
	return cmd
}

func newIssueRemoteLinkDeleteCmd() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:         "remote-link-delete <key> <linkId>",
		Short:       "Delete a remote (web) link from an issue",
		Annotations: mutates,
		Args:        cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, id := args[0], args[1]
			if !linksMiscConfirm(yes, fmt.Sprintf("Delete remote link %s on %s?", id, key)) {
				if noInputFlag && !yes {
					return fmt.Errorf("refusing to delete remote link %s without --yes (or run interactively)", id)
				}
				info("Aborted.")
				return nil
			}
			if err := jiraClient.DeleteRemoteLink(key, id); err != nil {
				return err
			}
			info("Deleted remote link %s on %s", id, key)
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")
	return cmd
}

// issuePropertyRow is the {key,value} pair shown for a single property.
type issuePropertyRow struct {
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value"`
}

func newIssuePropertyListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "property-list <key>",
		Short: "List an issue's property keys",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			keys, err := jiraClient.ListIssueProperties(args[0])
			if err != nil {
				return err
			}
			def := &output.TableDef{
				Headers: []string{"PROPERTY KEY"},
				RowFunc: func(item interface{}) []string {
					return []string{item.(string)}
				},
			}
			return render(keys, def)
		},
	}
}

func newIssuePropertyGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "property-get <key> <propKey>",
		Short: "Get an issue property value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			value, err := jiraClient.GetIssueProperty(args[0], args[1])
			if err != nil {
				return err
			}
			row := issuePropertyRow{Key: args[1], Value: value}
			def := &output.TableDef{
				Headers: []string{"PROPERTY KEY", "VALUE"},
				RowFunc: func(item interface{}) []string {
					r := item.(issuePropertyRow)
					return []string{r.Key, truncate(string(r.Value), 80)}
				},
			}
			return render(row, def)
		},
	}
}

func newIssuePropertySetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "property-set <key> <propKey> <jsonValue>",
		Short:       "Set an issue property to a raw JSON value",
		Annotations: mutates,
		Args:        cobra.ExactArgs(3),
		Long: `Store a property on an issue. The third argument is parsed as raw JSON.

Examples:
  jira issue property-set ABC-123 my.flag true
  jira issue property-set ABC-123 my.meta '{"reviewed":true,"by":"qa"}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, propKey, raw := args[0], args[1], args[2]
			if !json.Valid([]byte(raw)) {
				return fmt.Errorf("jsonValue is not valid JSON: %q", raw)
			}
			if err := jiraClient.SetIssueProperty(key, propKey, json.RawMessage(raw)); err != nil {
				return err
			}
			info("Set property %s on %s", propKey, key)
			return nil
		},
	}
	return cmd
}

func newIssuePropertyDeleteCmd() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:         "property-delete <key> <propKey>",
		Short:       "Delete an issue property",
		Annotations: mutates,
		Args:        cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, propKey := args[0], args[1]
			if !linksMiscConfirm(yes, fmt.Sprintf("Delete property %s on %s?", propKey, key)) {
				if noInputFlag && !yes {
					return fmt.Errorf("refusing to delete property %s without --yes (or run interactively)", propKey)
				}
				info("Aborted.")
				return nil
			}
			if err := jiraClient.DeleteIssueProperty(key, propKey); err != nil {
				return err
			}
			info("Deleted property %s on %s", propKey, key)
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")
	return cmd
}

func newIssueNotifyCmd() *cobra.Command {
	var (
		subject  string
		bodyText string
		to       []string
		users    []string
		markdown bool
	)
	cmd := &cobra.Command{
		Use:         "notify <key>",
		Short:       "Email a notification about an issue",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		Long: `Send an email notification about an issue. --subject and --body are required.

Recipients are chosen with --to (any of reporter,assignee,watchers,voters) and/or
--users (a comma-separated list of accountIds). By default --body is sent as the
plain-text body; use --markdown to send it as the HTML body instead.

Examples:
  jira issue notify ABC-123 --subject "Heads up" --body "please review" --to assignee,watchers
  jira issue notify ABC-123 --subject Ping --body "**urgent**" --markdown --users 5b10ac...,5b10ad...`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if subject == "" || bodyText == "" {
				return fmt.Errorf("--subject and --body are required")
			}
			body := map[string]any{"subject": subject}
			if markdown {
				body["htmlBody"] = bodyText
			} else {
				body["textBody"] = bodyText
			}
			recipients := map[string]any{}
			for _, t := range to {
				switch strings.ToLower(strings.TrimSpace(t)) {
				case "reporter":
					recipients["reporter"] = true
				case "assignee":
					recipients["assignee"] = true
				case "watchers":
					recipients["watchers"] = true
				case "voters":
					recipients["voters"] = true
				case "":
					// ignore empty CSV entries
				default:
					return fmt.Errorf("invalid --to value %q (want reporter, assignee, watchers, or voters)", t)
				}
			}
			var userObjs []map[string]string
			for _, u := range users {
				u = strings.TrimSpace(u)
				if u == "" {
					continue
				}
				userObjs = append(userObjs, map[string]string{"accountId": u})
			}
			if len(userObjs) > 0 {
				recipients["users"] = userObjs
			}
			if len(recipients) == 0 {
				return fmt.Errorf("no recipients; pass --to and/or --users")
			}
			body["to"] = recipients
			if err := jiraClient.Notify(args[0], body); err != nil {
				return err
			}
			info("Sent notification for %s", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&subject, "subject", "", "Notification subject (required)")
	cmd.Flags().StringVar(&bodyText, "body", "", "Notification body (required)")
	cmd.Flags().StringSliceVar(&to, "to", nil, "Audience: reporter,assignee,watchers,voters (comma-separated)")
	cmd.Flags().StringSliceVar(&users, "users", nil, "Specific recipients by accountId (comma-separated)")
	cmd.Flags().BoolVar(&markdown, "markdown", false, "Send --body as the HTML body instead of plain text")
	return cmd
}
