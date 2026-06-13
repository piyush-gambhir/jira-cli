package config

import "testing"

func TestIsCloudAndAPIVersion(t *testing.T) {
	cases := []struct {
		name       string
		profile    Profile
		wantCloud  bool
		wantAPIVer string
	}{
		{"api_token cloud", Profile{AuthType: AuthAPIToken, Site: "https://acme.atlassian.net"}, true, "3"},
		{"api_token server host", Profile{AuthType: AuthAPIToken, Site: "https://jira.acme.com"}, false, "2"},
		{"pat is server", Profile{AuthType: AuthPAT, Site: "https://jira.acme.com"}, false, "2"},
		{"scoped is cloud", Profile{AuthType: AuthScopedToken, Site: "https://acme.atlassian.net"}, true, "3"},
		{"oauth2 is cloud", Profile{AuthType: AuthOAuth2}, true, "3"},
		{"explicit api_version wins", Profile{AuthType: AuthAPIToken, Site: "https://acme.atlassian.net", APIVersion: "2"}, true, "2"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.profile.IsCloud(); got != c.wantCloud {
				t.Errorf("IsCloud() = %v, want %v", got, c.wantCloud)
			}
			if got := c.profile.EffectiveAPIVersion(); got != c.wantAPIVer {
				t.Errorf("EffectiveAPIVersion() = %q, want %q", got, c.wantAPIVer)
			}
		})
	}
}

func TestEffectiveAuthTypeDefault(t *testing.T) {
	if got := (Profile{}).EffectiveAuthType(); got != AuthAPIToken {
		t.Errorf("empty auth type should default to %q, got %q", AuthAPIToken, got)
	}
}

func TestResolveAuthPrecedence(t *testing.T) {
	cfg := &Config{
		CurrentProfile: "default",
		Profiles: map[string]Profile{
			"default": {AuthType: AuthAPIToken, Site: "https://config.atlassian.net", Email: "config@x.com", Token: "cfgtoken"},
		},
	}
	env := func(k string) (string, bool) {
		m := map[string]string{"JIRA_TOKEN": "envtoken"}
		v, ok := m[k]
		return v, ok
	}
	flags := FlagValues{Email: "flag@x.com", EmailSet: true}
	got, err := ResolveAuth(flags, env, cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Site != "https://config.atlassian.net" {
		t.Errorf("site from config not applied: %q", got.Site)
	}
	if got.Token != "envtoken" {
		t.Errorf("env should override config token: %q", got.Token)
	}
	if got.Email != "flag@x.com" {
		t.Errorf("flag should override config email: %q", got.Email)
	}
}
