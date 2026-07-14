package auth

import (
	"encoding/base64"
	"net/http"
	"strings"
	"testing"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/config"
)

func applyHeader(t *testing.T, a Authenticator) string {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, "https://example/x", nil)
	if err := a.Apply(req); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	return req.Header.Get("Authorization")
}

func TestAPITokenAuth(t *testing.T) {
	a, err := New(config.Profile{AuthType: config.AuthAPIToken, Site: "https://acme.atlassian.net/", Email: "me@acme.com", Token: "tok"}, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if a.BaseURL() != "https://acme.atlassian.net" {
		t.Errorf("base URL = %q", a.BaseURL())
	}
	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("me@acme.com:tok"))
	if got := applyHeader(t, a); got != want {
		t.Errorf("auth header = %q, want %q", got, want)
	}
}

func TestPATAuth(t *testing.T) {
	a, err := New(config.Profile{AuthType: config.AuthPAT, Site: "https://jira.acme.com", Token: "pat123"}, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := applyHeader(t, a); got != "Bearer pat123" {
		t.Errorf("auth header = %q", got)
	}
}

func TestScopedTokenUsesGateway(t *testing.T) {
	a, err := New(config.Profile{AuthType: config.AuthScopedToken, Email: "me@acme.com", Token: "tok", CloudID: "cid-123"}, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(a.BaseURL(), "api.atlassian.com/ex/jira/cid-123") {
		t.Errorf("scoped token should target the gateway, got %q", a.BaseURL())
	}
}

func TestMissingFieldsError(t *testing.T) {
	if _, err := New(config.Profile{AuthType: config.AuthAPIToken, Site: "https://acme.atlassian.net"}, "", nil); err == nil {
		t.Error("expected error for missing email/token")
	}
	if _, err := New(config.Profile{AuthType: config.AuthScopedToken, Email: "m", Token: "t"}, "", nil); err == nil {
		t.Error("expected error for scoped token without cloud_id")
	}
}
