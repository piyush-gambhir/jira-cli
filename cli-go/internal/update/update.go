// Package update checks GitHub for newer releases of the CLI and installs them.
//
// The release check is cached for 24h in update-check.json in the config dir
// (failures too, so an offline machine does not retry on every command), and
// the same file records which version the update notice was last shown for.
package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	// CacheFileName is the cache file inside the config dir.
	CacheFileName = "update-check.json"
	// CheckInterval is how long a release check (or a shown notice) is remembered.
	CheckInterval = 24 * time.Hour
	// DefaultBaseURL is the GitHub web origin. Releases are resolved and
	// downloaded from it, never from api.github.com, whose unauthenticated
	// limit (60 requests/hour per IP) breaks behind shared NAT, VPNs, and CI.
	DefaultBaseURL = "https://github.com"
)

// cacheFile is the JSON stored in CacheFileName.
type cacheFile struct {
	CheckedAt       time.Time `json:"checked_at"`
	LatestVersion   string    `json:"latest_version,omitempty"`
	Error           string    `json:"error,omitempty"`
	NotifiedVersion string    `json:"notified_version,omitempty"`
	NotifiedAt      time.Time `json:"notified_at"`
}

// Checker looks up the latest release of Repo and caches the answer in CacheDir.
type Checker struct {
	Repo     string // "owner/name"
	CacheDir string
	BaseURL  string           // defaults to DefaultBaseURL
	Client   *http.Client     // defaults to a client with a 3-second timeout (redirects are never followed)
	Now      func() time.Time // defaults to time.Now
}

func (c *Checker) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

func (c *Checker) cachePath() string { return filepath.Join(c.CacheDir, CacheFileName) }

// Latest returns the latest release version without a leading "v". Unless
// force is set, a check younger than CheckInterval is answered from the cache,
// including a cached failure. Every network check is written to the cache.
func (c *Checker) Latest(ctx context.Context, force bool) (string, error) {
	cache, _ := c.read()
	if !force && c.fresh(cache.CheckedAt) {
		if cache.Error != "" {
			return "", errors.New(cache.Error)
		}
		return Normalize(cache.LatestVersion), nil
	}
	latest, err := c.fetch(ctx)
	// Re-read: another process may have recorded a notice during the fetch.
	cache, _ = c.read()
	cache.CheckedAt = c.now()
	cache.LatestVersion = latest
	cache.Error = ""
	if err != nil {
		cache.Error = err.Error()
	}
	_ = c.write(cache)
	return latest, err
}

// Cached reports the result of a check younger than CheckInterval without
// using the network: the latest version ("" if that check failed), and whether
// such a check exists. A caller can then skip starting a background check.
func (c *Checker) Cached() (string, bool) {
	cache, ok := c.read()
	if !ok || !c.fresh(cache.CheckedAt) {
		return "", false
	}
	if cache.Error != "" || !IsRelease(cache.LatestVersion) {
		return "", true
	}
	return Normalize(cache.LatestVersion), true
}

// CachedLatest returns the latest version from a successful check younger than
// CheckInterval. It never uses the network.
func (c *Checker) CachedLatest() (string, bool) {
	latest, ok := c.Cached()
	return latest, ok && latest != ""
}

// ClaimNotice reports whether the update notice for latest should be shown now,
// and if so records it, so the notice appears at most once per version per
// CheckInterval. If the claim cannot be recorded, it reports false rather than
// repeating the notice on every command.
func (c *Checker) ClaimNotice(latest string) bool {
	cache, _ := c.read()
	latest = Normalize(latest)
	if Normalize(cache.NotifiedVersion) == latest && c.fresh(cache.NotifiedAt) {
		return false
	}
	cache.NotifiedVersion = latest
	cache.NotifiedAt = c.now()
	return c.write(cache) == nil
}

// ClearCache removes the cache file (after an update the cached answer is stale).
func (c *Checker) ClearCache() error {
	err := os.Remove(c.cachePath())
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (c *Checker) fresh(t time.Time) bool {
	age := c.now().Sub(t)
	return !t.IsZero() && age >= 0 && age < CheckInterval
}

// fetch resolves the latest release from the redirect that
// <base>/<repo>/releases/latest answers with (302 to .../releases/tag/<tag>),
// without following it and without the rate-limited GitHub API.
func (c *Checker) fetch(ctx context.Context) (string, error) {
	base := c.BaseURL
	if base == "" {
		base = DefaultBaseURL
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(base, "/")+"/"+c.Repo+"/releases/latest", nil)
	if err != nil {
		return "", err
	}
	client := http.Client{Timeout: 3 * time.Second}
	if c.Client != nil {
		client = *c.Client
	}
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
	return tagFromRedirect(resp, baseURL, c.Repo)
}

// tagFromRedirect extracts the version from a releases/latest redirect. Only a
// 302 to <base host>/<repo>/releases/tag/v<semver> counts (exactly, no
// surrounding whitespace).
func tagFromRedirect(resp *http.Response, base *url.URL, repo string) (string, error) {
	if resp.StatusCode != http.StatusFound {
		return "", fmt.Errorf("GitHub answered %s for the latest release, not a 302 redirect", resp.Status)
	}
	loc, err := resp.Location()
	if err != nil {
		return "", fmt.Errorf("the latest-release redirect has no usable Location: %w", err)
	}
	if !strings.EqualFold(loc.Host, base.Host) || loc.Scheme != base.Scheme {
		return "", fmt.Errorf("the latest-release redirect points to %s, not %s", loc.Redacted(), base.Host)
	}
	prefix := "/" + repo + "/releases/tag/"
	tag := ""
	if len(loc.Path) > len(prefix) && strings.EqualFold(loc.Path[:len(prefix)], prefix) {
		tag = loc.Path[len(prefix):]
	}
	if tag != "v"+Normalize(tag) || !IsRelease(tag) {
		return "", fmt.Errorf("the latest-release redirect (%s) does not name a v<major>.<minor>.<patch> tag", loc.Redacted())
	}
	return Normalize(tag), nil
}

func (c *Checker) read() (cacheFile, bool) {
	var cache cacheFile
	data, err := os.ReadFile(c.cachePath())
	if err != nil || json.Unmarshal(data, &cache) != nil {
		return cacheFile{}, false
	}
	return cache, true
}

// write replaces the cache atomically so a process that exits mid-write (the
// check runs in the background) never leaves a truncated file behind.
func (c *Checker) write(cache cacheFile) error {
	if err := os.MkdirAll(c.CacheDir, 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(c.CacheDir, ".update-check-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), c.cachePath())
}

// Normalize strips surrounding space and a leading "v".
func Normalize(v string) string { return strings.TrimPrefix(strings.TrimSpace(v), "v") }

// parse splits a release version (MAJOR.MINOR.PATCH, optional leading "v").
func parse(v string) ([3]int, bool) {
	var out [3]int
	parts := strings.Split(Normalize(v), ".")
	if len(parts) != 3 {
		return out, false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 || p == "" || (len(p) > 1 && p[0] == '0') || strings.ContainsAny(p, "+-") {
			return out, false
		}
		out[i] = n
	}
	return out, true
}

// IsRelease reports whether v is a release version. "dev", an empty string,
// and source builds such as "0.1.10-3-gabc1234-dirty" are not.
func IsRelease(v string) bool {
	_, ok := parse(v)
	return ok
}

// Newer reports whether latest is a newer release than current.
func Newer(latest, current string) bool {
	l, okL := parse(latest)
	c, okC := parse(current)
	if !okL || !okC {
		return false
	}
	for i := range l {
		if l[i] != c[i] {
			return l[i] > c[i]
		}
	}
	return false
}

// ReleaseURL is the release notes page for version.
func ReleaseURL(repo, version string) string {
	return "https://github.com/" + repo + "/releases/tag/v" + Normalize(version)
}

// Notice is the update notice printed on stderr after a command's output.
// updateCommand is how to update ("jira update", or the source-build command).
func Notice(bin, repo, current, latest, updateCommand string) string {
	return fmt.Sprintf("\nA new version of %s is available: v%s -> v%s\nUpdate with: %s\nRelease notes: %s\n",
		bin, Normalize(current), Normalize(latest), updateCommand, ReleaseURL(repo, latest))
}

// NotifierDisabledByEnv reports whether the environment turns the update
// notifier off: CI, <prefix>_NO_UPDATE_NOTIFIER, or NO_UPDATE_NOTIFIER set to
// any non-empty value.
func NotifierDisabledByEnv(prefix string, getenv func(string) string) bool {
	for _, key := range []string{"CI", prefix + "_NO_UPDATE_NOTIFIER", "NO_UPDATE_NOTIFIER"} {
		if getenv(key) != "" {
			return true
		}
	}
	return false
}
