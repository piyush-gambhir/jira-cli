package config

import "strings"

// Auth type identifiers stored in a profile's auth_type field.
const (
	AuthAPIToken    = "api_token"    // Cloud: Basic base64(email:token), base URL = site
	AuthScopedToken = "scoped_token" // Cloud: Basic base64(email:token), base URL = api.atlassian.com gateway
	AuthOAuth2      = "oauth2"       // Cloud: Bearer access token (3LO), base URL = gateway
	AuthPAT         = "pat"          // Server/DC: Bearer personal access token
	AuthBasic       = "basic"        // Server/DC: Basic base64(username:password)
)

// Config represents the top-level configuration file structure.
type Config struct {
	CurrentProfile string             `yaml:"current_profile"`
	Profiles       map[string]Profile `yaml:"profiles"`
	Defaults       Defaults           `yaml:"defaults"`
}

// Profile represents connection + authentication details for a Jira site.
//
// A single profile carries the union of fields across all auth methods; which
// fields are populated depends on AuthType. Secrets are stored here (file mode
// 0600). See docs/CREDENTIALS.md for the full auth reference.
type Profile struct {
	AuthType string `yaml:"auth_type"`          // one of the Auth* constants
	Site     string `yaml:"site"`               // base URL: https://x.atlassian.net (Cloud) or https://jira.host (DC)
	Email    string `yaml:"email,omitempty"`    // Cloud basic (api_token / scoped_token)
	Token    string `yaml:"token,omitempty"`    // API token (Cloud) or PAT (DC)
	Username string `yaml:"username,omitempty"` // DC basic
	Password string `yaml:"password,omitempty"` // DC basic
	CloudID  string `yaml:"cloud_id,omitempty"` // scoped_token / oauth2 gateway routing

	// OAuth 2.0 (3LO)
	ClientID     string `yaml:"client_id,omitempty"`
	ClientSecret string `yaml:"client_secret,omitempty"`
	AccessToken  string `yaml:"access_token,omitempty"`
	RefreshToken string `yaml:"refresh_token,omitempty"`
	TokenExpiry  int64  `yaml:"token_expiry,omitempty"` // unix seconds; access token expiry
	Scopes       string `yaml:"scopes,omitempty"`       // space-separated

	// Behaviour
	APIVersion string `yaml:"api_version,omitempty"` // platform API version: "3" (Cloud) or "2" (DC)
	Insecure   bool   `yaml:"insecure,omitempty"`    // skip TLS verification
	ReadOnly   bool   `yaml:"read_only,omitempty"`   // block mutating commands (agent safety)
}

// Defaults holds default settings.
type Defaults struct {
	Output string `yaml:"output,omitempty"`
}

// EffectiveAuthType returns the auth type, defaulting to api_token when unset
// (backwards-compatible with a minimally-configured profile).
func (p Profile) EffectiveAuthType() string {
	if p.AuthType == "" {
		return AuthAPIToken
	}
	return p.AuthType
}

// IsCloud reports whether this profile targets Jira Cloud (vs Server/DC).
// Derived from the auth type first, then the site host.
func (p Profile) IsCloud() bool {
	switch p.EffectiveAuthType() {
	case AuthPAT, AuthBasic:
		return false
	case AuthScopedToken, AuthOAuth2:
		return true
	default: // api_token
		return !looksLikeServer(p.Site)
	}
}

// EffectiveAPIVersion returns the platform REST API version: an explicit
// override if set, else "3" for Cloud and "2" for Server/DC.
func (p Profile) EffectiveAPIVersion() string {
	if p.APIVersion != "" {
		return p.APIVersion
	}
	if p.IsCloud() {
		return "3"
	}
	return "2"
}

func looksLikeServer(site string) bool {
	host := strings.ToLower(site)
	// Cloud sites are *.atlassian.net (or the api.atlassian.com gateway).
	if strings.Contains(host, "atlassian.net") || strings.Contains(host, "api.atlassian.com") {
		return false
	}
	return host != ""
}
