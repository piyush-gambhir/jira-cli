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
	CacheFileName  = "update-check.json"
	CheckInterval  = 24 * time.Hour
	DefaultBaseURL = "https://github.com"
)

type cacheFile struct {
	CheckedAt       time.Time `json:"checked_at"`
	LatestVersion   string    `json:"latest_version,omitempty"`
	Error           string    `json:"error,omitempty"`
	NotifiedVersion string    `json:"notified_version,omitempty"`
	NotifiedAt      time.Time `json:"notified_at"`
}

type Checker struct {
	Repo     string
	CacheDir string
	BaseURL  string
	Client   *http.Client
	Now      func() time.Time
}

func (c *Checker) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

func (c *Checker) cachePath() string { return filepath.Join(c.CacheDir, CacheFileName) }

func (c *Checker) Latest(ctx context.Context, force bool) (string, error) {
	cache, _ := c.read()
	if !force && c.fresh(cache.CheckedAt) {
		if cache.Error != "" {
			return "", errors.New(cache.Error)
		}
		return Normalize(cache.LatestVersion), nil
	}
	latest, err := c.fetch(ctx)
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

func (c *Checker) CachedLatest() (string, bool) {
	latest, ok := c.Cached()
	return latest, ok && latest != ""
}

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

func (c *Checker) fresh(t time.Time) bool {
	age := c.now().Sub(t)
	return !t.IsZero() && age >= 0 && age < CheckInterval
}

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

func Normalize(v string) string { return strings.TrimPrefix(strings.TrimSpace(v), "v") }

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

func IsRelease(v string) bool {
	_, ok := parse(v)
	return ok
}

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

func ReleaseURL(repo, version string) string {
	return "https://github.com/" + repo + "/releases/tag/v" + Normalize(version)
}

func Notice(bin, repo, current, latest, updateCommand string) string {
	return fmt.Sprintf("\nA new version of %s is available: v%s -> v%s\nUpdate with: %s\nRelease notes: %s\n",
		bin, Normalize(current), Normalize(latest), updateCommand, ReleaseURL(repo, latest))
}

func NotifierDisabledByEnv(prefix string, getenv func(string) string) bool {
	for _, key := range []string{"CI", prefix + "_NO_UPDATE_NOTIFIER", "NO_UPDATE_NOTIFIER"} {
		if getenv(key) != "" {
			return true
		}
	}
	return false
}
