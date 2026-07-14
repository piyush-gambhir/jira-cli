package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/adf"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/output"
)

func newIssueAssignCmd() *cobra.Command {
	var to string
	cmd := &cobra.Command{
		Use:         "assign <key>",
		Short:       "Assign an issue to a user",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		Long: `Assign an issue. --to accepts an email, display name, @me, id:<accountId>,
"default" (the project's default assignee), or "none"/"unassign" to clear it.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			var accountID any
			switch strings.ToLower(strings.TrimSpace(to)) {
			case "", "none", "unassign", "null":
				accountID = nil
			case "default", "-1":
				accountID = "-1"
			default:
				resolved, err := jiraClient.ResolveUser(to)
				if err != nil {
					return err
				}
				accountID = resolved
			}
			if err := jiraClient.AssignIssue(key, accountID); err != nil {
				return err
			}
			info("Assigned %s", key)
			return nil
		},
	}
	cmd.Flags().StringVar(&to, "to", "", "Assignee (email, name, @me, id:<accountId>, default, or none)")
	_ = cmd.MarkFlagRequired("to")
	return cmd
}

func newIssueTransitionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "transitions <key>",
		Short: "List the transitions available for an issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			transitions, err := jiraClient.GetTransitions(args[0])
			if err != nil {
				return err
			}
			def := &output.TableDef{
				Headers: []string{"ID", "NAME", "-> STATUS"},
				RowFunc: func(item interface{}) []string {
					t := item.(client.Transition)
					return []string{t.ID, t.Name, statusName(t.To)}
				},
			}
			return render(transitions, def)
		},
	}
}

func newIssueTransitionCmd() *cobra.Command {
	var resolution, comment string
	var markdown bool
	cmd := &cobra.Command{
		Use:         "transition <key> <status-or-transition>",
		Aliases:     []string{"move"},
		Short:       "Transition an issue to a new status",
		Annotations: mutates,
		Args:        cobra.ExactArgs(2),
		Long: `Transition an issue. The second argument matches either a target status name
(e.g. "In Progress", "Done") or a transition name, case-insensitively.

Examples:
  jira issue transition ABC-123 "In Progress"
  jira issue transition ABC-123 Done --resolution Fixed --comment "Shipped"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, target := args[0], args[1]
			transitions, err := jiraClient.GetTransitions(key)
			if err != nil {
				return err
			}
			var match *client.Transition
			for i := range transitions {
				t := transitions[i]
				if strings.EqualFold(t.Name, target) || (t.To != nil && strings.EqualFold(t.To.Name, target)) {
					match = &transitions[i]
					break
				}
			}
			if match == nil {
				var names []string
				for _, t := range transitions {
					names = append(names, fmt.Sprintf("%q -> %s", t.Name, statusName(t.To)))
				}
				return fmt.Errorf("no transition matching %q on %s. Available: %s", target, key, strings.Join(names, ", "))
			}

			var fields, update map[string]any
			if resolution != "" {
				fields = map[string]any{"resolution": map[string]string{"name": resolution}}
			}
			if comment != "" {
				var body any
				if markdown {
					body = adf.FromMarkdown(comment)
				} else {
					body = adf.FromPlainText(comment)
				}
				update = map[string]any{"comment": []map[string]any{{"add": map[string]any{"body": body}}}}
			}
			if err := jiraClient.DoTransition(key, match.ID, fields, update); err != nil {
				return err
			}
			info("Transitioned %s via %q -> %s", key, match.Name, statusName(match.To))
			return nil
		},
	}
	cmd.Flags().StringVar(&resolution, "resolution", "", "Set a resolution (e.g. Fixed) during the transition")
	cmd.Flags().StringVar(&comment, "comment", "", "Add a comment during the transition")
	cmd.Flags().BoolVar(&markdown, "markdown", false, "Interpret --comment as lightweight markdown")
	return cmd
}
