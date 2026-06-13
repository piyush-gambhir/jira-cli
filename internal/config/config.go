package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// persistMu serializes load-modify-save cycles within a process (notably OAuth
// token refresh, where a rotated refresh token must not be lost to a race).
var persistMu sync.Mutex

// ConfigDir returns the configuration directory path, respecting XDG_CONFIG_HOME.
func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "jira-cli")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".config", "jira-cli")
}

// ConfigPath returns the full path to the config file.
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

// Load reads the config file from disk. Returns a default config if file doesn't exist.
func Load() (*Config, error) {
	cfg := &Config{
		Profiles: make(map[string]Profile),
	}

	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}

	return cfg, nil
}

// Save writes the config to disk, creating directories as needed (mode 0600).
func Save(cfg *Config) error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	// Atomic write: write to a temp file then rename, so a crash mid-write
	// (or a concurrent reader) never sees a truncated config.
	tmp := ConfigPath() + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	if err := os.Rename(tmp, ConfigPath()); err != nil {
		return fmt.Errorf("replacing config: %w", err)
	}

	return nil
}

// SetProfile adds or updates a profile in the config.
func SetProfile(cfg *Config, name string, profile Profile) {
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}
	cfg.Profiles[name] = profile
}

// DeleteProfile removes a profile from the config.
func DeleteProfile(cfg *Config, name string) error {
	if _, ok := cfg.Profiles[name]; !ok {
		return fmt.Errorf("profile %q not found", name)
	}
	delete(cfg.Profiles, name)
	if cfg.CurrentProfile == name {
		cfg.CurrentProfile = ""
	}
	return nil
}

// SetCurrentProfile sets the active profile.
func SetCurrentProfile(cfg *Config, name string) error {
	if _, ok := cfg.Profiles[name]; !ok {
		return fmt.Errorf("profile %q not found", name)
	}
	cfg.CurrentProfile = name
	return nil
}

// GetCurrentProfile returns the currently active profile.
func GetCurrentProfile(cfg *Config) (Profile, error) {
	if cfg.CurrentProfile == "" {
		return Profile{}, fmt.Errorf("no current profile set")
	}
	p, ok := cfg.Profiles[cfg.CurrentProfile]
	if !ok {
		return Profile{}, fmt.Errorf("current profile %q not found in profiles", cfg.CurrentProfile)
	}
	return p, nil
}

// PersistProfile performs a serialized load-modify-save to update a single
// profile in place. Used by the OAuth authenticator to persist rotated tokens
// without clobbering other profiles or losing the new refresh token.
func PersistProfile(name string, profile Profile) error {
	persistMu.Lock()
	defer persistMu.Unlock()

	cfg, err := Load()
	if err != nil {
		return err
	}
	SetProfile(cfg, name, profile)
	return Save(cfg)
}
