package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/gofrs/flock"
	"go.yaml.in/yaml/v3"
)

var persistMu sync.Mutex

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

func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

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

func Save(cfg *Config) error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return fmt.Errorf("securing config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := atomicWriteFile(ConfigPath(), data, 0o600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	if err := os.Chmod(ConfigPath(), 0o600); err != nil {
		return fmt.Errorf("securing config: %w", err)
	}

	return nil
}

func SetProfile(cfg *Config, name string, profile Profile) {
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}
	cfg.Profiles[name] = profile
}

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

func SetCurrentProfile(cfg *Config, name string) error {
	if _, ok := cfg.Profiles[name]; !ok {
		return fmt.Errorf("profile %q not found", name)
	}
	cfg.CurrentProfile = name
	return nil
}

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

func Update(mutator func(*Config) error) error {
	persistMu.Lock()
	defer persistMu.Unlock()
	if err := os.MkdirAll(ConfigDir(), 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	lockPath := ConfigPath() + ".lock"
	fileLock := flock.New(lockPath)
	if err := fileLock.Lock(); err != nil {
		return fmt.Errorf("locking config: %w", err)
	}
	defer fileLock.Unlock()
	if err := os.Chmod(lockPath, 0o600); err != nil {
		return fmt.Errorf("securing config lock: %w", err)
	}

	cfg, err := Load()
	if err != nil {
		return err
	}
	if err := mutator(cfg); err != nil {
		return err
	}
	return Save(cfg)
}

func PersistProfile(name string, profile Profile) error {
	return Update(func(cfg *Config) error {
		SetProfile(cfg, name, profile)
		return nil
	})
}
