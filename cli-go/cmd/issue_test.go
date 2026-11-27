package cmd

import (
	"testing"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/client"
)

func TestMissingUserIsUnassignedOnlyForAssignee(t *testing.T) {
	if got := assigneeName(nil); got != "Unassigned" {
		t.Fatalf("assigneeName(nil) = %q; want Unassigned", got)
	}
	if got := userName(nil); got != "-" {
		t.Fatalf("userName(nil) = %q; want -", got)
	}
	row := projectTable().RowFunc(client.Project{Key: "ABC", Name: "Alpha"})
	if lead := row[3]; lead != "-" {
		t.Fatalf("project LEAD without a lead = %q; want -", lead)
	}
}
