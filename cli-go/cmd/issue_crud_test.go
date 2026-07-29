package cmd

import (
	"testing"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/client"
)

func TestRecentCreateJQLIsNarrowAndEscaped(t *testing.T) {
	got := recentCreateJQL(`ABC"DEF`)
	want := `project = "ABC\"DEF" AND creator = currentUser() AND created >= -10m ORDER BY created DESC`
	if got != want {
		t.Fatalf("recentCreateJQL() = %q; want %q", got, want)
	}
}

func TestFindRecentDuplicateMatchesTypeSummaryAndParent(t *testing.T) {
	issues := []client.Issue{
		{
			Key: "ABC-1",
			Fields: client.IssueFields{
				Summary:   "Same title",
				IssueType: &client.Named{Name: "Bug"},
			},
		},
		{
			Key: "ABC-2",
			Fields: client.IssueFields{
				Summary:   "Same title",
				IssueType: &client.Named{Name: "Task"},
				Parent:    &client.Issue{Key: "ABC-10"},
			},
		},
	}

	got := findRecentDuplicate(issues, "task", "Same title", "abc-10")
	if got == nil || got.Key != "ABC-2" {
		t.Fatalf("findRecentDuplicate() = %#v; want ABC-2", got)
	}
}

func TestFindRecentDuplicateRejectsNearMatches(t *testing.T) {
	issues := []client.Issue{
		{
			Key: "ABC-1",
			Fields: client.IssueFields{
				Summary:   "Same title with suffix",
				IssueType: &client.Named{Name: "Task"},
			},
		},
		{
			Key: "ABC-2",
			Fields: client.IssueFields{
				Summary:   "Same title",
				IssueType: &client.Named{Name: "Story"},
			},
		},
	}

	if got := findRecentDuplicate(issues, "Task", "Same title", ""); got != nil {
		t.Fatalf("findRecentDuplicate() = %#v; want nil", got)
	}
}

func TestIssueCreateExposesAllowDuplicateEscapeHatch(t *testing.T) {
	cmd := newIssueCreateCmd()
	flag := cmd.Flags().Lookup("allow-duplicate")
	if flag == nil {
		t.Fatal("issue create is missing --allow-duplicate")
	}
	if flag.DefValue != "false" {
		t.Fatalf("--allow-duplicate default = %q; want false", flag.DefValue)
	}
}
