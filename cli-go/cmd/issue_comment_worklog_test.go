package cmd

import (
	"strings"
	"testing"
)

func TestWorklogDeleteManualAdjustIncreasesEstimate(t *testing.T) {
	cmd := newIssueWorklogDeleteCmd()
	if !strings.Contains(cmd.Long, "manual  increase the estimate by --increase-by") {
		t.Fatalf("worklog-delete help does not say manual increases the estimate:\n%s", cmd.Long)
	}
	if usage := cmd.Flags().Lookup("increase-by").Usage; !strings.HasPrefix(usage, "Amount to increase") {
		t.Fatalf("--increase-by usage = %q; want it to say increase", usage)
	}
}
