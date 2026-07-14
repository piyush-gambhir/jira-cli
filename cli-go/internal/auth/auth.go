// Package auth implements every authentication method supported by the Jira
// REST API. See docs/CREDENTIALS.md for the full reference.
//
// Each method is an Authenticator: it knows how to stamp an outgoing request
// with the right credential header and what base URL API calls should target.
package auth

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/config"
)

// Authenticator stamps requests with credentials and reports the API base URL.
type Authenticator interface {
	// Apply sets the authorization header on req (refreshing tokens if needed).
	Apply(req *http.Request) error
	// BaseURL is the scheme://host[/ex/jira/{cloudId}] that API paths hang off of.
	BaseURL() string
	// Describe returns a short human-readable summary for whoami/status.
	Describe() string
}

// TokenPersister saves an updated profile (used by OAuth to store rotated tokens).
type TokenPersister func(updated config.Profile) error

// New builds the Authenticator for a profile's auth type. profileName and
// persist are only used by the OAuth2 authenticator to persist refreshed tokens;
// pass an empty name and nil persister for the other methods.
func New(p config.Profile, profileName string, persist TokenPersister) (Authenticator, error) {
	switch p.EffectiveAuthType() {
	case config.AuthAPIToken:
		if p.Site == "" || p.Email == "" || p.Token == "" {
			return nil, fmt.Errorf("api_token auth requires site, email and token (run 'jira auth login')")
		}
		return &basicAuth{baseURL: strings.TrimRight(p.Site, "/"), user: p.Email, secret: p.Token, kind: "API token (Cloud)"}, nil

	case config.AuthScopedToken:
		if p.Email == "" || p.Token == "" {
			return nil, fmt.Errorf("scoped_token auth requires email and token (run 'jira auth login')")
		}
		if p.CloudID == "" {
			return nil, fmt.Errorf("scoped_token auth requires cloud_id; run 'jira auth login --type scoped' to resolve it")
		}
		return &basicAuth{baseURL: gatewayBaseURL(p.CloudID), user: p.Email, secret: p.Token, kind: "Scoped API token (Cloud gateway)"}, nil

	case config.AuthPAT:
		if p.Site == "" || p.Token == "" {
			return nil, fmt.Errorf("pat auth requires site and token (run 'jira auth login --type pat')")
		}
		return &bearerAuth{baseURL: strings.TrimRight(p.Site, "/"), token: p.Token, kind: "Personal access token (Server/DC)"}, nil

	case config.AuthBasic:
		if p.Site == "" || p.Username == "" || p.Password == "" {
			return nil, fmt.Errorf("basic auth requires site, username and password (run 'jira auth login --type basic')")
		}
		return &basicAuth{baseURL: strings.TrimRight(p.Site, "/"), user: p.Username, secret: p.Password, kind: "Username/password (Server/DC)"}, nil

	case config.AuthOAuth2:
		return newOAuth2(p, profileName, persist)

	default:
		return nil, fmt.Errorf("unknown auth_type %q", p.AuthType)
	}
}

// gatewayBaseURL is the api.atlassian.com gateway base for scoped tokens / OAuth.
func gatewayBaseURL(cloudID string) string {
	return "https://api.atlassian.com/ex/jira/" + cloudID
}

// --- Basic auth (Cloud api_token, Cloud scoped_token, DC username/password) ---

type basicAuth struct {
	baseURL string
	user    string
	secret  string
	kind    string
}

func (b *basicAuth) Apply(req *http.Request) error {
	req.Header.Set("Authorization", "Basic "+basicCredential(b.user, b.secret))
	return nil
}
func (b *basicAuth) BaseURL() string  { return b.baseURL }
func (b *basicAuth) Describe() string { return fmt.Sprintf("%s as %s", b.kind, b.user) }

// --- Bearer auth (DC personal access token) ---

type bearerAuth struct {
	baseURL string
	token   string
	kind    string
}

func (b *bearerAuth) Apply(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+b.token)
	return nil
}
func (b *bearerAuth) BaseURL() string  { return b.baseURL }
func (b *bearerAuth) Describe() string { return b.kind }

// basicCredential returns base64(user:secret) with no trailing newline.
func basicCredential(user, secret string) string {
	return base64.StdEncoding.EncodeToString([]byte(user + ":" + secret))
}

// httpClientFor builds a short-lived HTTP client honoring the insecure flag,
// used by helpers that talk to Atlassian directly (cloudId / token endpoints).
func httpClientFor(insecure bool) *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
		},
	}
}

// ResolveCloudID resolves a Cloud site's cloudId via the unauthenticated
// /_edge/tenant_info endpoint. Used at login time for scoped tokens.
func ResolveCloudID(site string, insecure bool) (string, error) {
	site = strings.TrimRight(site, "/")
	resp, err := httpClientFor(insecure).Get(site + "/_edge/tenant_info")
	if err != nil {
		return "", fmt.Errorf("resolving cloudId: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("resolving cloudId: %s returned %s", site+"/_edge/tenant_info", resp.Status)
	}
	var out struct {
		CloudID string `json:"cloudId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decoding tenant_info: %w", err)
	}
	if out.CloudID == "" {
		return "", fmt.Errorf("tenant_info did not return a cloudId for %s", site)
	}
	return out.CloudID, nil
}

// AccessibleResource is one site an OAuth token can reach.
type AccessibleResource struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	URL    string   `json:"url"`
	Scopes []string `json:"scopes"`
}

// AccessibleResources lists the sites an OAuth access token can access.
func AccessibleResources(accessToken string, insecure bool) ([]AccessibleResource, error) {
	req, _ := http.NewRequest(http.MethodGet, "https://api.atlassian.com/oauth/token/accessible-resources", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	resp, err := httpClientFor(insecure).Do(req)
	if err != nil {
		return nil, fmt.Errorf("listing accessible resources: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("accessible-resources returned %s", resp.Status)
	}
	var out []AccessibleResource
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decoding accessible-resources: %w", err)
	}
	return out, nil
}
