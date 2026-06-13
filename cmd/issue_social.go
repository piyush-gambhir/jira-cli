package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/internal/client"
	"github.com/piyush-gambhir/jira-cli/internal/output"
)

func newIssueWatchCmd() *cobra.Command {
	var user string
	cmd := &cobra.Command{
		Use:         "watch <key>",
		Short:       "Watch an issue (or add another user as a watcher)",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			accountID := ""
			if user != "" {
				resolved, err := jiraClient.ResolveUser(user)
				if err != nil {
					return err
				}
				accountID = resolved
			}
			if err := jiraClient.AddWatcher(args[0], accountID); err != nil {
				return err
			}
			info("Watching %s", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&user, "user", "", "Add this user as a watcher (email, name, id:<accountId>); defaults to you")
	return cmd
}

func newIssueUnwatchCmd() *cobra.Command {
	var user string
	cmd := &cobra.Command{
		Use:         "unwatch <key>",
		Short:       "Stop watching an issue (or remove another watcher)",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref := user
			if ref == "" {
				ref = "@me"
			}
			accountID, err := jiraClient.ResolveUser(ref)
			if err != nil {
				return err
			}
			if err := jiraClient.RemoveWatcher(args[0], accountID); err != nil {
				return err
			}
			info("Stopped watching %s", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&user, "user", "", "Remove this watcher (defaults to you)")
	return cmd
}

func newIssueWatchersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "watchers <key>",
		Short: "List an issue's watchers",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			w, err := jiraClient.GetWatchers(args[0])
			if err != nil {
				return err
			}
			if outFormat != "table" {
				return render(w, nil)
			}
			info("%d watcher(s)", w.WatchCount)
			def := &output.TableDef{
				Headers: []string{"ACCOUNT ID", "NAME", "ACTIVE"},
				RowFunc: func(item interface{}) []string {
					u := item.(client.User)
					return []string{u.AccountID, u.DisplayName, fmt.Sprintf("%v", u.Active)}
				},
			}
			return render(w.Watchers, def)
		},
	}
}

func newIssueVoteCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "vote <key>",
		Short:       "Vote for an issue",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := jiraClient.AddVote(args[0]); err != nil {
				return err
			}
			info("Voted for %s", args[0])
			return nil
		},
	}
}

func newIssueUnvoteCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "unvote <key>",
		Short:       "Remove your vote from an issue",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := jiraClient.RemoveVote(args[0]); err != nil {
				return err
			}
			info("Removed vote from %s", args[0])
			return nil
		},
	}
}

func newIssueVotesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "votes <key>",
		Short: "Show vote count and voters for an issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := jiraClient.GetVotes(args[0])
			if err != nil {
				return err
			}
			if outFormat != "table" {
				return render(v, nil)
			}
			info("%d vote(s)", v.Votes)
			def := &output.TableDef{
				Headers: []string{"ACCOUNT ID", "NAME", "ACTIVE"},
				RowFunc: func(item interface{}) []string {
					u := item.(client.User)
					return []string{u.AccountID, u.DisplayName, fmt.Sprintf("%v", u.Active)}
				},
			}
			return render(v.Voters, def)
		},
	}
}
