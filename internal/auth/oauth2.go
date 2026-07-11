package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/piyush-gambhir/jira-cli/internal/config"
)

const (
	authorizeEndpoint = "https://auth.atlassian.com/authorize"
	tokenEndpoint     = "https://auth.atlassian.com/oauth/token"
	// DefaultScopes requested at login when the user doesn't specify their own
	// (the "write" preset: read + write issues/projects/users, plus offline_access).
	DefaultScopes = "read:jira-user read:jira-work write:jira-work offline_access"
	// DefaultCallbackPort is the loopback port for the OAuth redirect. The user
	// must register http://localhost:<port>/callback as the app's Callback URL.
	DefaultCallbackPort = 8765
)

// oauth2Auth implements Authenticator for the OAuth 2.0 (3LO) flow. It refreshes
// the access token on demand and persists the rotated refresh token.
type oauth2Auth struct {
	mu          sync.Mutex
	p           config.Profile
	profileName string
	persist     TokenPersister
	client      *http.Client
}

func newOAuth2(p config.Profile, profileName string, persist TokenPersister) (*oauth2Auth, error) {
	if p.CloudID == "" {
		return nil, fmt.Errorf("oauth2 auth requires cloud_id (run 'jira auth login --type oauth2')")
	}
	if p.AccessToken == "" && p.RefreshToken == "" {
		return nil, fmt.Errorf("oauth2 auth has no tokens (run 'jira auth login --type oauth2')")
	}
	return &oauth2Auth{p: p, profileName: profileName, persist: persist, client: httpClientFor(p.Insecure)}, nil
}

func (o *oauth2Auth) BaseURL() string  { return gatewayBaseURL(o.p.CloudID) }
func (o *oauth2Auth) Describe() string { return "OAuth 2.0 (3LO, Cloud)" }

func (o *oauth2Auth) Apply(req *http.Request) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Refresh if the access token is missing or within 60s of expiry.
	if o.p.AccessToken == "" || time.Now().Unix() >= o.p.TokenExpiry-60 {
		if err := o.refreshLocked(); err != nil {
			return err
		}
	}
	req.Header.Set("Authorization", "Bearer "+o.p.AccessToken)
	return nil
}

// refreshLocked exchanges the refresh token for a new access token. Caller holds o.mu.
func (o *oauth2Auth) refreshLocked() error {
	if o.p.RefreshToken == "" {
		return fmt.Errorf("OAuth access token expired and no refresh token available — run 'jira auth login --type oauth2'")
	}
	tok, err := tokenRequest(o.client, map[string]string{
		"grant_type":    "refresh_token",
		"client_id":     o.p.ClientID,
		"client_secret": o.p.ClientSecret,
		"refresh_token": o.p.RefreshToken,
	})
	if err != nil {
		return fmt.Errorf("refreshing OAuth token: %w", err)
	}
	o.p.AccessToken = tok.AccessToken
	if tok.RefreshToken != "" { // refresh tokens rotate — persist the new one
		o.p.RefreshToken = tok.RefreshToken
	}
	o.p.TokenExpiry = time.Now().Unix() + int64(tok.ExpiresIn)
	if o.persist != nil && o.profileName != "" {
		if err := o.persist(o.p); err != nil {
			return fmt.Errorf("persisting refreshed token: %w", err)
		}
	}
	return nil
}

// tokenResponse is the shape returned by the Atlassian token endpoint.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
}

// tokenRequest POSTs a JSON body to the token endpoint and parses the response.
func tokenRequest(client *http.Client, body map[string]string) (*tokenResponse, error) {
	payload, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, tokenEndpoint, strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, fmt.Errorf("decoding token response: %w", err)
	}
	if tr.Error != "" {
		return nil, fmt.Errorf("token endpoint error: %s (%s)", tr.Error, tr.ErrorDesc)
	}
	if tr.AccessToken == "" {
		return nil, fmt.Errorf("token endpoint returned no access_token (status %s)", resp.Status)
	}
	return &tr, nil
}

// OAuthResult is returned by OAuthLogin: the tokens plus the resolved scopes.
type OAuthResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	Scope        string
}

// OAuthLogin runs the interactive 3LO authorization-code flow: it starts a
// loopback HTTP server, opens the browser to the consent screen, captures the
// authorization code, and exchanges it for tokens. The app's registered
// Callback URL must be http://localhost:<port>/callback.
func OAuthLogin(clientID, clientSecret, scopes string, port int, insecure bool) (*OAuthResult, error) {
	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("oauth2 login requires --client-id and --client-secret (register an app at developer.atlassian.com)")
	}
	if scopes == "" {
		scopes = DefaultScopes
	}
	if port == 0 {
		port = DefaultCallbackPort
	}
	redirectURI := fmt.Sprintf("http://localhost:%d/callback", port)

	state, err := randomState()
	if err != nil {
		return nil, err
	}

	// Channels to receive the result from the callback handler.
	type cbResult struct {
		code string
		err  error
	}
	resultCh := make(chan cbResult, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		q := r.URL.Query()
		if e := q.Get("error"); e != "" {
			fmt.Fprintf(w, "Authorization failed: %s. You can close this tab.", html.EscapeString(e))
			resultCh <- cbResult{err: fmt.Errorf("authorization denied: %s (%s)", e, q.Get("error_description"))}
			return
		}
		if q.Get("state") != state {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			resultCh <- cbResult{err: fmt.Errorf("state mismatch (possible CSRF) — aborting")}
			return
		}
		code := q.Get("code")
		fmt.Fprint(w, "<html><body><h2>Authorization complete</h2>"+
			"<p>You can close this tab and return to the terminal.</p></body></html>")
		resultCh <- cbResult{code: code}
	})

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, fmt.Errorf("cannot listen on %s for the OAuth callback (is the port free?): %w", redirectURI, err)
	}
	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       15 * time.Second,
	}
	go func() { _ = srv.Serve(ln) }()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	authURL := authorizeEndpoint + "?" + url.Values{
		"audience":      {"api.atlassian.com"},
		"client_id":     {clientID},
		"scope":         {scopes},
		"redirect_uri":  {redirectURI},
		"state":         {state},
		"response_type": {"code"},
		"prompt":        {"consent"},
	}.Encode()

	fmt.Println("Opening your browser to authorize. If it doesn't open, visit:")
	fmt.Println("  " + authURL)
	_ = openBrowser(authURL)

	// Wait for the callback (with a generous timeout).
	var code string
	select {
	case res := <-resultCh:
		if res.err != nil {
			return nil, res.err
		}
		code = res.code
	case <-time.After(5 * time.Minute):
		return nil, fmt.Errorf("timed out waiting for browser authorization")
	}

	tok, err := tokenRequest(httpClientFor(insecure), map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     clientID,
		"client_secret": clientSecret,
		"code":          code,
		"redirect_uri":  redirectURI,
	})
	if err != nil {
		return nil, err
	}
	return &OAuthResult{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		ExpiresIn:    tok.ExpiresIn,
		Scope:        tok.Scope,
	}, nil
}

func randomState() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// openBrowser opens url in the default browser, best-effort.
func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}
