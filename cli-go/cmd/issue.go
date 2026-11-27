package cmd

import (
	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/output"
)

var mutates = map[string]string{"mutates": "true"}

func newIssueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "issue",
		Aliases: []string{"issues", "i"},
		Short:   "Create, query, and manage issues",
		Long: `Work with Jira issues: search (JQL), get, create, edit, delete, assign,
transition, comment, log work, attach files, link, watch, and vote.

Examples:
  jira issue list --mine
  jira issue search --jql "project = ABC AND statusCategory != Done" -o json
  jira issue get ABC-123
  jira issue create -p ABC --type Task --summary "Title" -d "Details"
  jira issue transition ABC-123 "In Progress"`,
	}
	cmd.AddCommand(
		newIssueListCmd(),
		newIssueSearchCmd(),
		newIssueGetCmd(),
		newIssueCreateCmd(),
		newIssueEditCmd(),
		newIssueDeleteCmd(),
		newIssueAssignCmd(),
		newIssueTransitionsCmd(),
		newIssueTransitionCmd(),
		newIssueCommentCmd(),
		newIssueCommentsCmd(),
		newIssueWorklogCmd(),
		newIssueWorklogsCmd(),
		newIssueAttachCmd(),
		newIssueAttachmentsCmd(),
		newIssueDownloadCmd(),
		newIssueLinkCmd(),
		newIssueLinkTypesCmd(),
		newIssueWatchCmd(),
		newIssueUnwatchCmd(),
		newIssueWatchersCmd(),
		newIssueVoteCmd(),
		newIssueUnvoteCmd(),
		newIssueVotesCmd(),
		newIssueCreateMetaCmd(),
		newIssueEditMetaCmd(),
		newIssueChangelogCmd(),
		newIssueCommentGetCmd(),
		newIssueCommentEditCmd(),
		newIssueCommentDeleteCmd(),
		newIssueWorklogGetCmd(),
		newIssueWorklogEditCmd(),
		newIssueWorklogDeleteCmd(),
		newIssueLinkDeleteCmd(),
		newIssueAttachmentDeleteCmd(),
		newIssueAttachmentMetaCmd(),
		newIssueRemoteLinksCmd(),
		newIssueRemoteLinkAddCmd(),
		newIssueRemoteLinkDeleteCmd(),
		newIssuePropertyListCmd(),
		newIssuePropertyGetCmd(),
		newIssuePropertySetCmd(),
		newIssuePropertyDeleteCmd(),
		newIssueNotifyCmd(),
	)
	return cmd
}

func issueListTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"KEY", "TYPE", "STATUS", "PRIORITY", "ASSIGNEE", "SUMMARY"},
		RowFunc: func(item interface{}) []string {
			i := item.(client.Issue)
			f := i.Fields
			return []string{
				i.Key,
				namedName(f.IssueType),
				statusName(f.Status),
				namedName(f.Priority),
				assigneeName(f.Assignee),
				truncate(f.Summary, 60),
			}
		},
	}
}

func namedName(n *client.Named) string {
	if n == nil {
		return "-"
	}
	return dash(n.Name)
}

func statusName(s *client.Status) string {
	if s == nil {
		return "-"
	}
	return dash(s.Name)
}

func assigneeName(u *client.User) string {
	if u == nil {
		return "Unassigned"
	}
	return userName(u)
}

func userName(u *client.User) string {
	if u == nil {
		return "-"
	}
	if u.DisplayName != "" {
		return u.DisplayName
	}
	return userID(*u)
}

func userID(u client.User) string {
	if u.AccountID != "" {
		return u.AccountID
	}
	return dash(u.Name)
}
