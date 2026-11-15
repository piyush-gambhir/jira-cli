package config

import "strings"

const (
	AuthAPIToken    = "api_token"
	AuthScopedToken = "scoped_token"
	AuthOAuth2      = "oauth2"
	AuthPAT         = "pat"
	AuthBasic       = "basic"
)

type Config struct {
	CurrentProfile string             `yaml:"current_profile"`
	Profiles       map[string]Profile `yaml:"profiles"`
	Defaults       Defaults           `yaml:"defaults"`
}

type Profile struct {
	AuthType string `yaml:"auth_type"`
	Site     string `yaml:"site"`
	Email    string `yaml:"email,omitempty"`
	Token    string `yaml:"token,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	CloudID  string `yaml:"cloud_id,omitempty"`

	ClientID     string `yaml:"client_id,omitempty"`
	ClientSecret string `yaml:"client_secret,omitempty"`
	AccessToken  string `yaml:"access_token,omitempty"`
	RefreshToken string `yaml:"refresh_token,omitempty"`
	TokenExpiry  int64  `yaml:"token_expiry,omitempty"`
	Scopes       string `yaml:"scopes,omitempty"`

	APIVersion string `yaml:"api_version,omitempty"`
	Insecure   bool   `yaml:"insecure,omitempty"`
	ReadOnly   bool   `yaml:"read_only,omitempty"`
}

type Defaults struct {
	Output string `yaml:"output,omitempty"`
}

func (p Profile) EffectiveAuthType() string {
	if p.AuthType == "" {
		return AuthAPIToken
	}
	return p.AuthType
}

func (p Profile) IsCloud() bool {
	switch p.EffectiveAuthType() {
	case AuthPAT, AuthBasic:
		return false
	case AuthScopedToken, AuthOAuth2:
		return true
	default:
		return !looksLikeServer(p.Site)
	}
}

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
	if strings.Contains(host, "atlassian.net") || strings.Contains(host, "api.atlassian.com") {
		return false
	}
	return host != ""
}
