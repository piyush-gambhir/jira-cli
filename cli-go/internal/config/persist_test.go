package config

import "testing"

func TestPersistTokensKeepsStoredSettings(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	stored := Profile{AuthType: AuthOAuth2, Site: "https://acme.atlassian.net", CloudID: "c1", ClientID: "id", AccessToken: "old", RefreshToken: "r1", TokenExpiry: 1}
	if err := Save(&Config{CurrentProfile: "default", Profiles: map[string]Profile{"default": stored}}); err != nil {
		t.Fatal(err)
	}
	merged := stored
	merged.Insecure = true
	merged.APIVersion = "2"
	merged.ClientSecret = "from-env"
	merged.ReadOnly = true
	merged.AccessToken, merged.RefreshToken, merged.TokenExpiry = "new", "r2", 99
	if err := PersistTokens("default", merged); err != nil {
		t.Fatal(err)
	}
	if err := PersistTokens("env-only", merged); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	want := stored
	want.AccessToken, want.RefreshToken, want.TokenExpiry = "new", "r2", 99
	if got := cfg.Profiles["default"]; got != want {
		t.Fatalf("stored profile = %+v; want only the tokens updated: %+v", got, want)
	}
	if _, ok := cfg.Profiles["env-only"]; ok || len(cfg.Profiles) != 1 {
		t.Fatalf("profiles = %+v; want no profile created for an unsaved name", cfg.Profiles)
	}
}

func TestPersistTokensSkipsChangedClientID(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	stored := Profile{AuthType: AuthOAuth2, Site: "https://acme.atlassian.net", CloudID: "c1", ClientID: "A", AccessToken: "old", RefreshToken: "r1", TokenExpiry: 1}
	if err := Save(&Config{CurrentProfile: "default", Profiles: map[string]Profile{"default": stored}}); err != nil {
		t.Fatal(err)
	}
	other := stored
	other.ClientID = "B"
	other.AccessToken, other.RefreshToken, other.TokenExpiry = "new", "r2", 99
	if err := PersistTokens("default", other); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.Profiles["default"]; got != stored {
		t.Fatalf("stored profile = %+v; want it unchanged: %+v", got, stored)
	}
}
