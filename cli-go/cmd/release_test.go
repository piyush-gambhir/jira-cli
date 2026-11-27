package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/auth"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/config"
)

func TestReleaseCountFailsWhenUnresolvedCountFails(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/rest/api/2/version/10001/relatedIssueCounts":
			_, _ = w.Write([]byte(`{"issuesFixedCount":3}`))
		case "/rest/api/2/version/10001/unresolvedIssueCount":
			http.Error(w, `{"errorMessages":["forbidden"]}`, http.StatusForbidden)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(ts.Close)
	a, err := auth.New(config.Profile{Site: ts.URL, Token: "pat", AuthType: config.AuthPAT}, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	old := jiraClient
	jiraClient = client.NewClient(a, "2", false, false)
	t.Cleanup(func() { jiraClient = old })

	cmd := newReleaseCountCmd()
	err = cmd.RunE(cmd, []string{"10001"})
	if err == nil {
		t.Fatal("release count succeeded although the unresolved count request failed")
	}
	if !strings.Contains(err.Error(), "forbidden") {
		t.Fatalf("error = %q; want it to contain forbidden", err)
	}
}
