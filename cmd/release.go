package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/internal/client"
	"github.com/piyush-gambhir/jira-cli/internal/output"
)

func newReleaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "release",
		Aliases: []string{"releases", "version-mgmt"},
		Short:   "List, inspect, and manage project versions (releases)",
		Long: `Manage Jira project versions, which the Jira UI calls "Releases".

Create, update, publish, merge, and delete versions, and inspect the issue
counts tied to a version. (This is unrelated to the binary's own 'version'
command, which prints the CLI version.)

Examples:
  jira release list ABC
  jira release create --project ABC --name "1.0" --release-date 2026-07-01
  jira release publish 10001
  jira release count 10001`,
	}
	cmd.AddCommand(
		newReleaseListCmd(),
		newReleaseGetCmd(),
		newReleaseCreateCmd(),
		newReleaseUpdateCmd(),
		newReleasePublishCmd(),
		newReleaseDeleteCmd(),
		newReleaseMergeCmd(),
		newReleaseCountCmd(),
	)
	return cmd
}

func releaseTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"ID", "NAME", "STATUS", "RELEASE DATE", "START DATE"},
		RowFunc: func(item interface{}) []string {
			v := item.(client.Version)
			return []string{v.ID, dash(v.Name), releaseStatus(v), dash(v.ReleaseDate), dash(v.StartDate)}
		},
	}
}

// releaseStatus renders the released/archived booleans as a single column.
func releaseStatus(v client.Version) string {
	switch {
	case v.Archived:
		return "archived"
	case v.Released:
		return "released"
	default:
		return "unreleased"
	}
}

func newReleaseListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <projectIdOrKey>",
		Short: "List the versions of a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			versions, err := jiraClient.ListVersions(args[0])
			if err != nil {
				return err
			}
			return render(versions, releaseTable())
		},
	}
}

func newReleaseGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := jiraClient.GetVersion(args[0])
			if err != nil {
				return err
			}
			return render(*v, releaseTable())
		},
	}
}

func newReleaseCreateCmd() *cobra.Command {
	var project, name, description, releaseDate, startDate string
	cmd := &cobra.Command{
		Use:         "create",
		Short:       "Create a version in a project",
		Annotations: mutates,
		Args:        cobra.NoArgs,
		Long: `Create a project version (release). --project and --name are required.

The project key is resolved to its numeric projectId (required by the API).

Examples:
  jira release create --project ABC --name "1.0"
  jira release create --project ABC --name "1.1" --release-date 2026-08-01 \
    -d "Summer release"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if project == "" || name == "" {
				return fmt.Errorf("--project and --name are required")
			}
			pid, err := jiraClient.ResolveProjectID(project)
			if err != nil {
				return err
			}
			body := map[string]any{"name": name, "projectId": pid}
			if description != "" {
				body["description"] = description
			}
			if releaseDate != "" {
				body["releaseDate"] = releaseDate
			}
			if startDate != "" {
				body["startDate"] = startDate
			}
			v, err := jiraClient.CreateVersion(body)
			if err != nil {
				return err
			}
			info("Created version %s (id %s)", v.Name, v.ID)
			return render(*v, releaseTable())
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "Project id or key (required)")
	cmd.Flags().StringVar(&name, "name", "", "Version name (required)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Version description")
	cmd.Flags().StringVar(&releaseDate, "release-date", "", "Release date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&startDate, "start-date", "", "Start date (YYYY-MM-DD)")
	return cmd
}

func newReleaseUpdateCmd() *cobra.Command {
	var name, description, releaseDate, startDate string
	var archived bool
	cmd := &cobra.Command{
		Use:         "update <id>",
		Short:       "Update a version's fields",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]any{}
			if name != "" {
				body["name"] = name
			}
			if description != "" {
				body["description"] = description
			}
			if releaseDate != "" {
				body["releaseDate"] = releaseDate
			}
			if startDate != "" {
				body["startDate"] = startDate
			}
			if cmd.Flags().Changed("archived") {
				body["archived"] = archived
			}
			if len(body) == 0 {
				return fmt.Errorf("nothing to update; pass --name/--description/--release-date/--start-date/--archived")
			}
			v, err := jiraClient.UpdateVersion(args[0], body)
			if err != nil {
				return err
			}
			info("Updated version %s", args[0])
			return render(*v, releaseTable())
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "New version name")
	cmd.Flags().StringVarP(&description, "description", "d", "", "New description")
	cmd.Flags().StringVar(&releaseDate, "release-date", "", "New release date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&startDate, "start-date", "", "New start date (YYYY-MM-DD)")
	cmd.Flags().BoolVar(&archived, "archived", false, "Archive (true) or unarchive (false) the version")
	return cmd
}

func newReleasePublishCmd() *cobra.Command {
	var releaseDate string
	cmd := &cobra.Command{
		Use:         "publish <id>",
		Short:       "Mark a version as released",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		Long: `Mark a version as released (sets released=true). With --release-date the
version's release date is set at the same time.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]any{"released": true}
			if releaseDate != "" {
				body["releaseDate"] = releaseDate
			}
			v, err := jiraClient.UpdateVersion(args[0], body)
			if err != nil {
				return err
			}
			info("Published version %s", args[0])
			return render(*v, releaseTable())
		},
	}
	cmd.Flags().StringVar(&releaseDate, "release-date", "", "Release date to set (YYYY-MM-DD)")
	return cmd
}

func newReleaseDeleteCmd() *cobra.Command {
	var moveFixTo, moveAffectedTo string
	var yes bool
	cmd := &cobra.Command{
		Use:         "delete <id>",
		Short:       "Delete a version",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		Long: `Delete a version. Issues referencing the version can be re-pointed at
another version with --move-fix-to / --move-affected-to (omit both to just
delete and clear the references).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				if noInputFlag {
					return fmt.Errorf("refusing to delete version %s without --yes", args[0])
				}
				ans, _ := prompt(fmt.Sprintf("Delete version %s? [y/N]: ", args[0]))
				if !strings.EqualFold(strings.TrimSpace(ans), "y") {
					info("Aborted.")
					return nil
				}
			}
			if err := jiraClient.DeleteVersionAndSwap(args[0], moveFixTo, moveAffectedTo); err != nil {
				return err
			}
			info("Deleted version %s", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&moveFixTo, "move-fix-to", "", "Version id to move 'fix version' references to")
	cmd.Flags().StringVar(&moveAffectedTo, "move-affected-to", "", "Version id to move 'affects version' references to")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")
	return cmd
}

func newReleaseMergeCmd() *cobra.Command {
	var into string
	var yes bool
	cmd := &cobra.Command{
		Use:         "merge <id> --into <targetId>",
		Short:       "Merge a version into another version",
		Annotations: mutates,
		Args:        cobra.ExactArgs(1),
		Long: `Merge version <id> into the --into target version: all issues referencing
<id> are re-pointed at the target, then <id> is deleted.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if into == "" {
				return fmt.Errorf("--into <targetId> is required")
			}
			if !yes {
				if noInputFlag {
					return fmt.Errorf("refusing to merge version %s into %s without --yes", args[0], into)
				}
				ans, _ := prompt(fmt.Sprintf("Merge version %s into %s? This deletes %s. [y/N]: ", args[0], into, args[0]))
				if !strings.EqualFold(strings.TrimSpace(ans), "y") {
					info("Aborted.")
					return nil
				}
			}
			if err := jiraClient.MergeVersion(args[0], into); err != nil {
				return err
			}
			info("Merged version %s into %s", args[0], into)
			return nil
		},
	}
	cmd.Flags().StringVar(&into, "into", "", "Target version id to merge into (required)")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip the confirmation prompt")
	return cmd
}

func newReleaseCountCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "count <id>",
		Short: "Show issue counts related to a version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			related, err := jiraClient.VersionRelatedCounts(args[0])
			if err != nil {
				return err
			}
			unresolved := 0
			if u, err := jiraClient.VersionUnresolvedIssues(args[0]); err == nil {
				unresolved = u.IssuesUnresolvedCount
			}
			out := releaseCounts{
				FixIssues:    related.IssuesFixedCount,
				AffectsIssue: related.IssuesAffectedCount,
				CustomFields: related.IssueCountWithCustomFieldsShowingVersion,
				Unresolved:   unresolved,
			}
			return render(out, releaseCountTable())
		},
	}
}

// releaseCounts is the flattened shape rendered by `release count`.
type releaseCounts struct {
	FixIssues    int `json:"fixIssues"`
	AffectsIssue int `json:"affectsIssues"`
	CustomFields int `json:"customFieldIssues"`
	Unresolved   int `json:"unresolved"`
}

func releaseCountTable() *output.TableDef {
	return &output.TableDef{
		Headers: []string{"FIX", "AFFECTS", "CUSTOM-FIELD", "UNRESOLVED"},
		RowFunc: func(item interface{}) []string {
			c := item.(releaseCounts)
			return []string{
				strconv.Itoa(c.FixIssues),
				strconv.Itoa(c.AffectsIssue),
				strconv.Itoa(c.CustomFields),
				strconv.Itoa(c.Unresolved),
			}
		},
	}
}
