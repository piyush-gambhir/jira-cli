package config

import (
	"fmt"
	"strconv"
	"strings"
)

type EnvLookupFunc func(string) (string, bool)

type FlagValues struct {
	Site     string
	Email    string
	Token    string
	User     string
	Insecure bool

	SiteSet     bool
	EmailSet    bool
	TokenSet    bool
	UserSet     bool
	InsecureSet bool
}

func ResolveAuth(flags FlagValues, envLookup EnvLookupFunc, cfg *Config, profileName string) (Profile, error) {
	var base Profile

	pName := profileName
	if pName == "" {
		pName = cfg.CurrentProfile
	}
	if pName != "" {
		p, ok := cfg.Profiles[pName]
		if !ok {
			return Profile{}, fmt.Errorf("profile %q not found", pName)
		}
		base = p
	}

	if envLookup != nil {
		setStr := func(key string, dst *string) {
			if v, ok := envLookup(key); ok && v != "" {
				*dst = v
			}
		}
		setBool := func(key string, dst *bool) {
			if v, ok := envLookup(key); ok && v != "" {
				if b, err := strconv.ParseBool(v); err == nil {
					*dst = b
				}
			}
		}
		setStr("JIRA_SITE", &base.Site)
		setStr("JIRA_EMAIL", &base.Email)
		setStr("JIRA_TOKEN", &base.Token)
		setStr("JIRA_USER", &base.Username)
		setStr("JIRA_PASSWORD", &base.Password)
		setStr("JIRA_AUTH_TYPE", &base.AuthType)
		setStr("JIRA_CLOUD_ID", &base.CloudID)
		setStr("JIRA_API_VERSION", &base.APIVersion)
		setStr("JIRA_OAUTH_CLIENT_ID", &base.ClientID)
		setStr("JIRA_OAUTH_CLIENT_SECRET", &base.ClientSecret)
		setStr("JIRA_OAUTH_ACCESS_TOKEN", &base.AccessToken)
		setStr("JIRA_OAUTH_REFRESH_TOKEN", &base.RefreshToken)
		setStr("JIRA_SCOPES", &base.Scopes)
		setBool("JIRA_INSECURE", &base.Insecure)
		setBool("JIRA_READ_ONLY", &base.ReadOnly)
	}

	if flags.SiteSet && flags.Site != "" {
		base.Site = flags.Site
	}
	if flags.EmailSet && flags.Email != "" {
		base.Email = flags.Email
	}
	if flags.TokenSet && flags.Token != "" {
		base.Token = flags.Token
	}
	if flags.UserSet && flags.User != "" {
		base.Username = flags.User
	}
	if flags.InsecureSet {
		base.Insecure = flags.Insecure
	}

	base.Site = strings.TrimRight(base.Site, "/")
	return base, nil
}
