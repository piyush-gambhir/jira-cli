package cmd

import (
	"testing"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/config"
)

// issue download only reads Jira, but writes a file, so read-only mode refuses it.
func TestAttachmentDownloadIsBlockedInReadOnlyMode(t *testing.T) {
	if err := checkReadOnly(newIssueDownloadCmd(), config.Profile{ReadOnly: true}); err == nil {
		t.Fatal("issue download allowed in read-only mode")
	}
	if err := checkReadOnly(newIssueDownloadCmd(), config.Profile{}); err != nil {
		t.Fatalf("issue download blocked outside read-only mode: %v", err)
	}
}
