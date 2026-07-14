// Package update implements a best-effort background check for newer releases
// on GitHub, cached for 24h so it never slows normal command execution.
package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// UpdateInfo describes the result of a release check.
type UpdateInfo struct {
	Available      bool
	CurrentVersion string
	LatestVersion  string
	URL            string
}

type cacheFile struct {
	CheckedAt     time.Time `json:"checked_at"`
	LatestVersion string    `json:"latest_version"`
	URL           string    `json:"url"`
}

const cacheTTL = 24 * time.Hour

// CheckForUpdate compares the current version against the latest GitHub release
// for repo (e.g. "piyush-gambhir/jira-cli"). Results are cached in cacheDir for
// 24h. Errors are returned but are safe to ignore (the check is best-effort).
func CheckForUpdate(current, repo, cacheDir string, force bool) (*UpdateInfo, error) {
	cachePath := filepath.Join(cacheDir, "update-check.json")

	if !force {
		if c, ok := readCache(cachePath); ok && time.Since(c.CheckedAt) < cacheTTL {
			return result(current, c.LatestVersion, c.URL), nil
		}
	}

	latest, url, err := fetchLatest(repo)
	if err != nil {
		return nil, err
	}
	_ = writeCache(cachePath, cacheFile{CheckedAt: time.Now(), LatestVersion: latest, URL: url})
	return result(current, latest, url), nil
}

func result(current, latest, url string) *UpdateInfo {
	return &UpdateInfo{
		Available:      latest != "" && current != "dev" && normalize(latest) != normalize(current),
		CurrentVersion: current,
		LatestVersion:  latest,
		URL:            url,
	}
}

func fetchLatest(repo string) (string, string, error) {
	req, _ := http.NewRequest(http.MethodGet, "https://api.github.com/repos/"+repo+"/releases/latest", nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("github releases returned %s", resp.Status)
	}
	body, _ := io.ReadAll(resp.Body)
	var rel struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.Unmarshal(body, &rel); err != nil {
		return "", "", err
	}
	return rel.TagName, rel.HTMLURL, nil
}

func readCache(path string) (cacheFile, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return cacheFile{}, false
	}
	var c cacheFile
	if json.Unmarshal(data, &c) != nil {
		return cacheFile{}, false
	}
	return c, true
}

func writeCache(path string, c cacheFile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, _ := json.Marshal(c)
	return os.WriteFile(path, data, 0o600)
}

func normalize(v string) string { return strings.TrimPrefix(strings.TrimSpace(v), "v") }

// PrintUpdateNotice prints a short upgrade hint to w.
func PrintUpdateNotice(w io.Writer, info *UpdateInfo) {
	if info == nil || !info.Available {
		return
	}
	fmt.Fprintf(w, "\nA new version of jira is available: %s -> %s\n", normalize(info.CurrentVersion), normalize(info.LatestVersion))
	if info.URL != "" {
		fmt.Fprintf(w, "  %s\n", info.URL)
	}
}
