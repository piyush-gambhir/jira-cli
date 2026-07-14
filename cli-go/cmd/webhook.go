package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/output"
)

func newWebhookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "webhook",
		Aliases: []string{"webhooks"},
		Short:   "Manage dynamic (OAuth-app) webhooks",
		Long: `List, register, delete, and refresh dynamic webhooks.

Dynamic webhooks are scoped to the OAuth 2.0 (3LO) app you authenticated as, so
these commands require an OAuth login with the manage:jira-webhook scope (choose
the "admin" or "all" scope preset at 'jira auth login'). Webhook ids are integers.

Examples:
  jira webhook list
  jira webhook register --url https://example.com/hook \
    --jql "project = ABC" --events jira:issue_created,jira:issue_updated
  jira webhook refresh 10001 10002
  jira webhook delete 10001 --yes`,
	}
	cmd.AddCommand(
		newWebhookListCmd(),
		newWebhookRegisterCmd(),
		newWebhookDeleteCmd(),
		newWebhookRefreshCmd(),
	)
	return cmd
}

func webhookTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"ID", "EVENTS", "JQL", "EXPIRES"},
		RowFunc: func(item interface{}) []string {
			w := item.(client.Webhook)
			return []string{
				strconv.Itoa(w.ID),
				dash(strings.Join(w.Events, ",")),
				dash(truncate(w.JqlFilter, 50)),
				dash(w.ExpirationDate),
			}
		},
	}
}

func newWebhookListCmd() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List registered dynamic webhooks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			webhooks, err := jiraClient.ListWebhooks(limit)
			if err != nil {
				return err
			}
			return render(webhooks, webhookTable())
		},
	}
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Maximum webhooks to return")
	return cmd
}

func newWebhookRegisterCmd() *cobra.Command {
	var (
		callbackURL string
		jql         string
		events      string
		fieldIDs    string
	)
	cmd := &cobra.Command{
		Use:         "register",
		Short:       "Register a dynamic webhook",
		Annotations: mutates,
		Args:        cobra.NoArgs,
		Long: `Register a dynamic webhook for the authenticated OAuth app. --url, --jql,
and --events are required. The --events value is a comma-separated list, e.g.
jira:issue_created,jira:issue_updated,jira:issue_deleted,comment_created.

Examples:
  jira webhook register --url https://example.com/hook \
    --jql "project = ABC" --events jira:issue_created,jira:issue_updated
  jira webhook register --url https://example.com/hook --jql "project = ABC" \
    --events jira:issue_updated --field-ids summary,description`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if callbackURL == "" {
				return fmt.Errorf("--url is required")
			}
			if jql == "" {
				return fmt.Errorf("--jql is required")
			}
			evs := splitCSV(events)
			if len(evs) == 0 {
				return fmt.Errorf("--events is required (comma-separated, e.g. jira:issue_created,jira:issue_updated)")
			}
			reg := client.WebhookRegistration{
				JqlFilter:      jql,
				Events:         evs,
				FieldIDsFilter: splitCSV(fieldIDs),
			}
			results, err := jiraClient.RegisterWebhooks(callbackURL, []client.WebhookRegistration{reg})
			if err != nil {
				return err
			}
			// Surface per-webhook validation errors (the call can 200 with errors).
			for _, r := range results {
				if len(r.Errors) > 0 {
					return fmt.Errorf("webhook registration failed: %s", strings.Join(r.Errors, "; "))
				}
			}
			if len(results) > 0 && results[0].CreatedWebhookID != 0 {
				info("Registered webhook %d", results[0].CreatedWebhookID)
			}
			return render(results, webhookResultTable())
		},
	}
	cmd.Flags().StringVar(&callbackURL, "url", "", "Callback URL to deliver events to (required)")
	cmd.Flags().StringVar(&jql, "jql", "", "JQL filter selecting the issues to watch (required)")
	cmd.Flags().StringVar(&events, "events", "", "Comma-separated events, e.g. jira:issue_created,jira:issue_updated (required)")
	cmd.Flags().StringVar(&fieldIDs, "field-ids", "", "Comma-separated field ids to restrict jira:issue_updated events")
	return cmd
}

func webhookResultTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"CREATED ID", "ERRORS"},
		RowFunc: func(item interface{}) []string {
			r := item.(client.WebhookRegistrationResult)
			id := "-"
			if r.CreatedWebhookID != 0 {
				id = strconv.Itoa(r.CreatedWebhookID)
			}
			return []string{id, dash(strings.Join(r.Errors, "; "))}
		},
	}
}

func newWebhookDeleteCmd() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:         "delete <id>...",
		Short:       "Delete one or more dynamic webhooks",
		Annotations: mutates,
		Args:        cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := webhookParseIDs(args)
			if err != nil {
				return err
			}
			if !yes {
				if noInputFlag {
					return fmt.Errorf("refusing to delete %d webhook(s) without --yes", len(ids))
				}
				ans, _ := prompt(fmt.Sprintf("Delete %d webhook(s) %v? [y/N]: ", len(ids), ids))
				if !strings.EqualFold(strings.TrimSpace(ans), "y") {
					info("Aborted.")
					return nil
				}
			}
			if err := jiraClient.DeleteWebhooks(ids); err != nil {
				return err
			}
			info("Deleted %d webhook(s)", len(ids))
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")
	return cmd
}

func newWebhookRefreshCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "refresh <id>...",
		Short:       "Refresh (extend the expiry of) one or more dynamic webhooks",
		Annotations: mutates,
		Args:        cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := webhookParseIDs(args)
			if err != nil {
				return err
			}
			expires, err := jiraClient.RefreshWebhooks(ids)
			if err != nil {
				return err
			}
			info("Refreshed %d webhook(s)", len(ids))
			return render(map[string]string{"expirationDate": expires}, nil)
		},
	}
}

// webhookParseIDs converts positional args into integer webhook ids.
func webhookParseIDs(args []string) ([]int, error) {
	ids := make([]int, 0, len(args))
	for _, a := range args {
		n, err := strconv.Atoi(strings.TrimSpace(a))
		if err != nil {
			return nil, fmt.Errorf("invalid webhook id %q (must be an integer)", a)
		}
		ids = append(ids, n)
	}
	return ids, nil
}
