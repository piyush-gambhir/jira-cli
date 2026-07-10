package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/piyush-gambhir/jira-cli/internal/auth"
	"github.com/piyush-gambhir/jira-cli/internal/config"
)

func testClient(t *testing.T, serverURL string) *Client {
	t.Helper()
	a, err := auth.New(config.Profile{
		Site:     serverURL,
		Email:    "user@example.com",
		Token:    "token",
		AuthType: config.AuthAPIToken,
	}, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	return NewClient(a, "3", false, false)
}

func TestMutatingPostIsNotRetried(t *testing.T) {
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		http.Error(w, "temporary failure", http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	c := testClient(t, ts.URL)
	err := c.PostJSON(c.api("issue"), nil, map[string]any{"fields": map[string]any{}}, nil)
	if err == nil {
		t.Fatal("expected POST failure")
	}
	if calls != 1 {
		t.Fatalf("mutating POST was attempted %d times; want 1", calls)
	}
}
