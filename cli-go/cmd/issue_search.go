package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newIssueListCmd() *cobra.Command {
	var (
		project  string
		status   string
		assignee string
		issType  string
		label    string
		mine     bool
		limit    int
		all      bool
		fields   string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issues using simple filters (builds a JQL query)",
		Long: `List issues with convenience filters. Combine any of --project, --status,
--assignee/--mine, --type, --label. For full control use 'jira issue search --jql'.

Examples:
  jira issue list --mine
  jira issue list -p ABC --status "In Progress"
  jira issue list -p ABC --type Bug --label regression -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var clauses []string
			if project != "" {
				clauses = append(clauses, "project = "+jqlQuote(project))
			}
			if status != "" {
				clauses = append(clauses, "status = "+jqlQuote(status))
			}
			if issType != "" {
				clauses = append(clauses, "issuetype = "+jqlQuote(issType))
			}
			if label != "" {
				clauses = append(clauses, "labels = "+jqlQuote(label))
			}
			if mine {
				clauses = append(clauses, "assignee = currentUser()")
			} else if assignee != "" {
				acct, err := jiraClient.ResolveUser(assignee)
				if err != nil {
					return err
				}
				clauses = append(clauses, "assignee = "+jqlQuote(acct))
			}
			if len(clauses) == 0 {
				// Bounded JQL is required; default to the caller's open work.
				clauses = append(clauses, "assignee = currentUser()")
			}
			jql := strings.Join(clauses, " AND ") + " ORDER BY updated DESC"
			info("JQL: %s", jql)
			issues, err := jiraClient.SearchIssues(jql, splitCSV(fields), limit, all)
			if err != nil {
				return err
			}
			return render(issues, issueListTable())
		},
	}
	cmd.Flags().StringVarP(&project, "project", "p", "", "Filter by project key")
	cmd.Flags().StringVar(&status, "status", "", "Filter by status name")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Filter by assignee (email, display name, @me, or id:<accountId>)")
	cmd.Flags().BoolVar(&mine, "mine", false, "Only issues assigned to me")
	cmd.Flags().StringVar(&issType, "type", "", "Filter by issue type")
	cmd.Flags().StringVar(&label, "label", "", "Filter by label")
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Maximum issues to return (ignored with --all)")
	cmd.Flags().BoolVar(&all, "all", false, "Fetch all matching issues (paginate)")
	cmd.Flags().StringVar(&fields, "fields", "", "Comma-separated fields to request")
	return cmd
}

func newIssueSearchCmd() *cobra.Command {
	var (
		jql    string
		limit  int
		all    bool
		fields string
		count  bool
	)
	cmd := &cobra.Command{
		Use:   "search [jql]",
		Short: "Search issues with a raw JQL query",
		Long: `Run a JQL query. On Cloud this uses the enhanced /search/jql endpoint
(cursor pagination); on Server/DC it uses the classic search.

JQL must be bounded (contain a real restriction). Pass the query as an argument
or via --jql.

Examples:
  jira issue search "project = ABC AND statusCategory != Done ORDER BY updated DESC"
  jira issue search --jql "assignee = currentUser() AND resolution IS EMPTY" -o json
  jira issue search "project = ABC" --count`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if jql == "" && len(args) == 1 {
				jql = args[0]
			}
			if jql == "" {
				return fmt.Errorf("a JQL query is required (positional arg or --jql)")
			}
			if count {
				n, err := jiraClient.ApproximateCount(jql)
				if err != nil {
					return err
				}
				return render(map[string]int{"count": n}, nil)
			}
			issues, err := jiraClient.SearchIssues(jql, splitCSV(fields), limit, all)
			if err != nil {
				return err
			}
			return render(issues, issueListTable())
		},
	}
	cmd.Flags().StringVarP(&jql, "jql", "j", "", "JQL query")
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Maximum issues to return (ignored with --all)")
	cmd.Flags().BoolVar(&all, "all", false, "Fetch all matching issues (paginate)")
	cmd.Flags().StringVar(&fields, "fields", "", "Comma-separated fields to request")
	cmd.Flags().BoolVar(&count, "count", false, "Return only the approximate match count")
	return cmd
}

// jqlQuote wraps a value in double quotes, escaping embedded quotes.
func jqlQuote(v string) string {
	return `"` + strings.ReplaceAll(v, `"`, `\"`) + `"`
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
