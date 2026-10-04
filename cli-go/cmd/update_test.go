package cmd

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/spf13/pflag"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/update"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/version"
)

// updateEnv isolates a test from the real config dir, environment, terminal,
// GitHub, and binary: the releases/latest redirect and downloads go to an
// httptest server that reports latest as the newest release, and the "running
// binary" is a temp file outside any Go bin directory.
type updateEnv struct {
	t          *testing.T
	mu         sync.Mutex
	latest     string
	latestHits atomic.Int32
	dlHits     atomic.Int32
	files      map[string][]byte
	exe        string
	cfgDir     string
}

func newUpdateEnv(t *testing.T, latest string) *updateEnv {
	t.Helper()
	e := &updateEnv{t: t, latest: latest, files: map[string][]byte{}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		latest := e.getLatest()
		if r.URL.Path == "/"+repoSlug+"/releases/latest" {
			e.latestHits.Add(1)
			http.Redirect(w, r, "/"+repoSlug+"/releases/tag/v"+latest, http.StatusFound)
			return
		}
		e.dlHits.Add(1)
		name := strings.TrimPrefix(r.URL.Path, "/"+repoSlug+"/releases/download/v"+latest+"/")
		e.mu.Lock()
		data, ok := e.files[name]
		e.mu.Unlock()
		if ok {
			_, _ = w.Write(data)
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)

	xdg := t.TempDir()
	e.cfgDir = filepath.Join(xdg, "jira-cli")
	t.Setenv("XDG_CONFIG_HOME", xdg)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("GOBIN", "")
	t.Setenv("GOPATH", "")
	for _, k := range []string{"CI", "JIRA_NO_UPDATE_NOTIFIER", "NO_UPDATE_NOTIFIER", "JIRA_QUIET", "JIRA_NO_INPUT", "JIRA_READ_ONLY", "JIRA_PROFILE"} {
		t.Setenv(k, "")
	}
	e.exe = filepath.Join(t.TempDir(), "jira")
	if err := os.WriteFile(e.exe, []byte("old binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	saved := struct {
		base, ver  string
		exe        func() (string, error)
		stdin, tty func() bool
		wait       time.Duration
	}{updateBaseURL, version.Version, executablePath, stdinIsTerminal, stderrIsTerminal, updateNoticeWait}
	t.Cleanup(func() {
		updateBaseURL, version.Version = saved.base, saved.ver
		executablePath, stdinIsTerminal, stderrIsTerminal = saved.exe, saved.stdin, saved.tty
		updateNoticeWait = saved.wait
		updateResult, updateChecker = nil, nil
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
	})
	updateBaseURL = srv.URL
	version.Version = "0.1.10"
	executablePath = func() (string, error) { return e.exe, nil }
	stdinIsTerminal = func() bool { return false }
	stderrIsTerminal = func() bool { return true }
	updateNoticeWait = 5 * time.Second
	return e
}

func (e *updateEnv) getLatest() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.latest
}

func (e *updateEnv) setLatest(v string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.latest = v
}

// serveRelease publishes the release archive for this platform (a zip with
// jira.exe on Windows, else a tar.gz with jira) and its checksum.
func (e *updateEnv) serveRelease(binary string) {
	e.t.Helper()
	var buf bytes.Buffer
	if updateGOOS == "windows" {
		zw := zip.NewWriter(&buf)
		w, err := zw.Create("jira.exe")
		if err != nil {
			e.t.Fatal(err)
		}
		_, _ = w.Write([]byte(binary))
		_ = zw.Close()
	} else {
		gz := gzip.NewWriter(&buf)
		tw := tar.NewWriter(gz)
		if err := tw.WriteHeader(&tar.Header{Name: "jira", Mode: 0o755, Size: int64(len(binary)), Typeflag: tar.TypeReg}); err != nil {
			e.t.Fatal(err)
		}
		_, _ = tw.Write([]byte(binary))
		_ = tw.Close()
		_ = gz.Close()
	}
	asset := (&update.Installer{Project: releaseProject, Binary: "jira", GOOS: updateGOOS}).AssetName()
	sum := sha256.Sum256(buf.Bytes())
	e.mu.Lock()
	defer e.mu.Unlock()
	e.files[asset] = buf.Bytes()
	e.files["checksums.txt"] = []byte(hex.EncodeToString(sum[:]) + "  " + asset + "\n")
}

// run executes jira with args and returns stdout and stderr.
func (e *updateEnv) run(args ...string) (string, string, error) {
	e.t.Helper()
	resetRootFlags(e.t)
	if c, _, err := rootCmd.Find([]string{"update"}); err == nil {
		c.Flags().VisitAll(func(f *pflag.Flag) {
			_ = f.Value.Set(f.DefValue)
			f.Changed = false
		})
	}
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return stdout.String(), stderr.String(), err
}

const wantNotice = "\nA new version of jira is available: v0.1.10 -> v0.1.11\n" +
	"Update with: jira update\n" +
	"Release notes: https://github.com/piyush-gambhir/jira-cli/releases/tag/v0.1.11\n"

func TestUpdateNoticeShownOncePerVersionPerDay(t *testing.T) {
	e := newUpdateEnv(t, "0.1.11")
	_, stderr, err := e.run("auth", "list")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(stderr, wantNotice) {
		t.Fatalf("stderr = %q; want it to end with the notice %q", stderr, wantNotice)
	}
	if _, stderr, _ = e.run("auth", "list"); strings.Contains(stderr, "new version") {
		t.Fatalf("notice shown twice for the same version: %q", stderr)
	}
	if e.latestHits.Load() != 1 {
		t.Fatalf("release API hits = %d; want 1 (cached for 24h)", e.latestHits.Load())
	}

	// A newer release is announced right away, even within the same day.
	e.setLatest("0.1.12")
	if _, err := newUpdateChecker(time.Second).Latest(t.Context(), true); err != nil {
		t.Fatal(err)
	}
	if _, stderr, _ = e.run("auth", "list"); !strings.Contains(stderr, "v0.1.10 -> v0.1.12") {
		t.Fatalf("notice for 0.1.12 missing: %q", stderr)
	}
}

func TestUpdateNoticeFromAFreshCacheNeedsNoWait(t *testing.T) {
	e := newUpdateEnv(t, "0.1.11")
	if _, err := newUpdateChecker(time.Second).Latest(t.Context(), false); err != nil {
		t.Fatal(err)
	}
	updateNoticeWait = 0
	_, stderr, err := e.run("auth", "list")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(stderr, wantNotice) || e.latestHits.Load() != 1 {
		t.Fatalf("stderr = %q, API hits = %d; want the cached notice with no new request", stderr, e.latestHits.Load())
	}
}

func TestUpdateNoticeForGoBinInstallSuggestsASourceBuild(t *testing.T) {
	e := newUpdateEnv(t, "0.1.11")
	t.Setenv("GOBIN", filepath.Dir(e.exe))
	_, stderr, err := e.run("auth", "list")
	if err != nil {
		t.Fatal(err)
	}
	want := "Update with: git pull && make install (in your jira-cli/cli-go checkout)\n"
	if !strings.Contains(stderr, want) || strings.Contains(stderr, "Update with: jira update") {
		t.Fatalf("stderr = %q; want the update line %q", stderr, want)
	}
}

func TestUpdateNotifierSuppressed(t *testing.T) {
	for name, setup := range map[string]func(e *updateEnv) []string{
		"stderr not a terminal": func(e *updateEnv) []string { stderrIsTerminal = func() bool { return false }; return nil },
		"CI":                    func(e *updateEnv) []string { e.t.Setenv("CI", "true"); return nil },
		"JIRA_NO_UPDATE_NOTIFIER": func(e *updateEnv) []string {
			e.t.Setenv("JIRA_NO_UPDATE_NOTIFIER", "1")
			return nil
		},
		"NO_UPDATE_NOTIFIER": func(e *updateEnv) []string { e.t.Setenv("NO_UPDATE_NOTIFIER", "yes"); return nil },
		"--quiet":            func(e *updateEnv) []string { return []string{"--quiet"} },
		"JIRA_QUIET":         func(e *updateEnv) []string { e.t.Setenv("JIRA_QUIET", "1"); return nil },
		"dev build":          func(e *updateEnv) []string { version.Version = "dev"; return nil },
		"source build":       func(e *updateEnv) []string { version.Version = "v0.1.10-3-gabc1234-dirty"; return nil },
	} {
		t.Run(name, func(t *testing.T) {
			e := newUpdateEnv(t, "0.1.11")
			extra := setup(e)
			_, stderr, err := e.run(append([]string{"auth", "list"}, extra...)...)
			if err != nil {
				t.Fatal(err)
			}
			if updateResult != nil || e.latestHits.Load() != 0 || strings.Contains(stderr, "new version") {
				t.Fatalf("check started=%v hits=%d stderr=%q; want no check and no output", updateResult != nil, e.latestHits.Load(), stderr)
			}
		})
	}

	// Skipped commands never start the check. (update --check queries GitHub
	// itself, so the update command is checked without running it.)
	e := newUpdateEnv(t, "0.1.11")
	for _, args := range [][]string{
		{"version"}, {"completion", "bash"}, {"help"}, {"__complete", "issue", ""}, {"__completeNoDesc", "issue", ""},
	} {
		if _, _, err := e.run(args...); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		if updateResult != nil || e.latestHits.Load() != 0 {
			t.Fatalf("%v: background check started; want skipped", args)
		}
	}
	for args, want := range map[string]bool{"update": false, "auth list": true, "issue update": true, "project update": true} {
		c, _, err := rootCmd.Find(strings.Fields(args))
		if err != nil {
			t.Fatal(err)
		}
		if got := updateNotifierEnabled(c); got != want {
			t.Errorf("%s: background check enabled = %v; want %v", args, got, want)
		}
	}
}

// releaseAfter serves the releases/latest redirect to v<latest> after delay
// (or when release is closed, whichever comes first) and becomes the release
// source for the test.
func releaseAfter(t *testing.T, latest string, delay time.Duration, release chan struct{}) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(delay):
		case <-release:
		}
		http.Redirect(w, r, "/"+repoSlug+"/releases/tag/v"+latest, http.StatusFound)
	}))
	t.Cleanup(srv.Close)
	updateBaseURL = srv.URL
}

func TestUpdateNoticeWaitsBrieflyForTheDaysCheck(t *testing.T) {
	e := newUpdateEnv(t, "0.1.11")
	updateNoticeWait = updateNoticeGrace
	releaseAfter(t, "0.1.11", 200*time.Millisecond, nil)
	_, stderr, err := e.run("auth", "list")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(stderr, wantNotice) {
		t.Fatalf("stderr = %q; want the notice on the run that started the check", stderr)
	}
}

func TestUpdateNoticeWaitsAtMostTheGrace(t *testing.T) {
	e := newUpdateEnv(t, "0.1.11")
	updateNoticeWait = updateNoticeGrace
	release := make(chan struct{})
	releaseAfter(t, "0.1.11", time.Minute, release)
	start := time.Now()
	_, stderr, err := e.run("auth", "list")
	elapsed, pending := time.Since(start), updateResult
	close(release)
	if pending != nil {
		<-pending // let the background check finish before the temp dirs go away
	}
	if err != nil {
		t.Fatal(err)
	}
	if pending == nil || elapsed > updateNoticeGrace+750*time.Millisecond || strings.Contains(stderr, "new version") {
		t.Fatalf("check started=%v, command took %v, stderr %q; want at most a ~1s wait and no notice", pending != nil, elapsed, stderr)
	}
}

func TestUpdateCheckJSON(t *testing.T) {
	e := newUpdateEnv(t, "0.1.11")
	// A fresh cache saying 0.1.10 is latest must not hide the new release.
	if err := os.MkdirAll(e.cfgDir, 0o700); err != nil {
		t.Fatal(err)
	}
	cache := `{"checked_at":"` + time.Now().Format(time.RFC3339) + `","latest_version":"0.1.10"}`
	if err := os.WriteFile(filepath.Join(e.cfgDir, update.CacheFileName), []byte(cache), 0o600); err != nil {
		t.Fatal(err)
	}
	stdout, _, err := e.run("update", "--check", "-o", "json")
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("stdout %q is not JSON: %v", stdout, err)
	}
	want := map[string]any{
		"current_version":  "0.1.10",
		"latest_version":   "0.1.11",
		"update_available": true,
		"release_url":      "https://github.com/piyush-gambhir/jira-cli/releases/tag/v0.1.11",
		"install_method":   "self",
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %v; want %v", k, got[k], v)
		}
	}
	if len(got) != len(want) || e.latestHits.Load() != 1 || e.dlHits.Load() != 0 {
		t.Fatalf("got %v, api hits %d, downloads %d; want exactly the 5 fields, 1 API call, no download", got, e.latestHits.Load(), e.dlHits.Load())
	}

	// The notifier shares the answer instead of trusting its stale cache.
	if v, ok := newUpdateChecker(0).CachedLatest(); !ok || v != "0.1.11" {
		t.Fatalf("cached latest after --check = %q, %v; want 0.1.11", v, ok)
	}

	t.Setenv("GOBIN", filepath.Dir(e.exe))
	stdout, _, _ = e.run("update", "--check", "-o", "json")
	if !strings.Contains(stdout, `"install_method": "go"`) {
		t.Fatalf("Go bin install not reported: %s", stdout)
	}
}

func TestUpdateWhenAlreadyLatest(t *testing.T) {
	e := newUpdateEnv(t, "0.1.10")
	stdout, _, err := e.run("update", "--yes")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "jira v0.1.10 is already the latest version.") || e.dlHits.Load() != 0 {
		t.Fatalf("stdout = %q, downloads = %d", stdout, e.dlHits.Load())
	}
	stdout, _, err = e.run("update", "--check")
	if err != nil || !strings.Contains(stdout, "Update available: no") {
		t.Fatalf("--check: %q, %v", stdout, err)
	}
}

func TestUpdateInstallsTheRelease(t *testing.T) {
	e := newUpdateEnv(t, "0.1.11")
	e.serveRelease("new binary")
	if _, err := newUpdateChecker(time.Second).Latest(t.Context(), false); err != nil {
		t.Fatal(err)
	}
	stdout, _, err := e.run("update", "-y")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "Updated jira v0.1.10 -> v0.1.11\n") ||
		!strings.Contains(stdout, "Release notes: https://github.com/piyush-gambhir/jira-cli/releases/tag/v0.1.11\n") {
		t.Fatalf("stdout = %q", stdout)
	}
	if data, _ := os.ReadFile(e.exe); string(data) != "new binary" {
		t.Fatalf("binary = %q; want the new release", data)
	}
	if v, ok := newUpdateChecker(0).CachedLatest(); !ok || v != "0.1.11" {
		t.Fatalf("cached latest after update = %q, %v; want 0.1.11", v, ok)
	}
}

func TestUpdateNeedsYesWithoutAPrompt(t *testing.T) {
	e := newUpdateEnv(t, "0.1.11")
	e.serveRelease("new binary")
	for _, args := range [][]string{{"update", "--no-input"}, {"update"}} { // stdin is not a terminal
		_, _, err := e.run(args...)
		if err == nil || !strings.Contains(err.Error(), "pass --yes") {
			t.Fatalf("%v: err = %v; want a request to pass --yes", args, err)
		}
	}
	if data, _ := os.ReadFile(e.exe); string(data) != "old binary" || e.dlHits.Load() != 0 {
		t.Fatalf("binary = %q, downloads = %d; want nothing installed", data, e.dlHits.Load())
	}
}

func TestUpdateGoBinInstallIsNotReplaced(t *testing.T) {
	e := newUpdateEnv(t, "0.1.11")
	e.serveRelease("new binary")
	t.Setenv("GOBIN", filepath.Dir(e.exe))
	stdout, _, err := e.run("update", "--yes")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "Update with: git pull && make install") || e.dlHits.Load() != 0 {
		t.Fatalf("stdout = %q, downloads = %d", stdout, e.dlHits.Load())
	}
	if data, _ := os.ReadFile(e.exe); string(data) != "old binary" {
		t.Fatalf("binary replaced: %q", data)
	}
}

func TestUpdateReadOnlyBlocksInstallNotCheck(t *testing.T) {
	e := newUpdateEnv(t, "0.1.11")
	e.serveRelease("new binary")
	for _, setup := range []func() []string{
		func() []string { return []string{"--read-only"} },
		func() []string { t.Setenv("JIRA_READ_ONLY", "true"); return nil },
	} {
		extra := setup()
		_, _, err := e.run(append([]string{"update", "--yes"}, extra...)...)
		if err == nil || !strings.Contains(err.Error(), "read-only") {
			t.Fatalf("err = %v; want a read-only refusal", err)
		}
		if _, _, err := e.run(append([]string{"update", "--check"}, extra...)...); err != nil {
			t.Fatalf("--check under read-only: %v", err)
		}
	}
	if data, _ := os.ReadFile(e.exe); string(data) != "old binary" || e.dlHits.Load() != 0 {
		t.Fatalf("binary = %q, downloads = %d; want nothing installed", data, e.dlHits.Load())
	}
}

func TestUpdateUnwritableDirKeepsTheOldBinary(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("needs POSIX permissions and a non-root user")
	}
	e := newUpdateEnv(t, "0.1.11")
	e.serveRelease("new binary")
	dir := filepath.Dir(e.exe)
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
	_, _, err := e.run("update", "--yes")
	if err == nil || !strings.Contains(err.Error(), "sudo") || !strings.Contains(err.Error(), "install.sh") {
		t.Fatalf("err = %v; want advice to use sudo or the install script", err)
	}
	if data, _ := os.ReadFile(e.exe); string(data) != "old binary" {
		t.Fatalf("binary = %q; want it untouched", data)
	}
}

func TestVersionShowsCachedLatestWithoutNetwork(t *testing.T) {
	e := newUpdateEnv(t, "0.1.11")
	stdout, _, err := e.run("version")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stdout, "latest") {
		t.Fatalf("version with no cache: %q", stdout)
	}
	if _, err := newUpdateChecker(time.Second).Latest(t.Context(), false); err != nil {
		t.Fatal(err)
	}
	hits := e.latestHits.Load()
	stdout, _, err = e.run("version")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(stdout, "jira version 0.1.10 (") || !strings.Contains(stdout, "\nlatest: 0.1.11\nupdate_available: true\n") {
		t.Fatalf("version = %q", stdout)
	}
	if e.latestHits.Load() != hits {
		t.Fatal("jira version contacted the release API")
	}
}

func TestUpdatePromptAnswers(t *testing.T) {
	for _, tc := range []struct {
		input   string
		install bool
	}{{"\n", true}, {"y\n", true}, {"YES\n", true}, {"n\n", false}, {"later\n", false}, {"", false}} {
		e := newUpdateEnv(t, "0.1.11")
		e.serveRelease("new binary")
		stdinIsTerminal = func() bool { return true }
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.WriteString(tc.input)
		w.Close()
		saved := os.Stdin
		os.Stdin = r
		stdout, stderr, err := e.run("update")
		os.Stdin = saved
		r.Close()
		if err != nil {
			t.Fatalf("%q: %v", tc.input, err)
		}
		data, _ := os.ReadFile(e.exe)
		if installed := string(data) == "new binary"; installed != tc.install || !strings.Contains(stderr, "Update now? [Y/n] ") {
			t.Fatalf("%q: installed=%v (stdout %q, stderr %q); want installed=%v after the prompt", tc.input, installed, stdout, stderr, tc.install)
		}
		if !tc.install && !strings.Contains(stdout, "Update cancelled.") {
			t.Fatalf("%q: stdout = %q; want a cancellation", tc.input, stdout)
		}
	}
}

func TestUpdateFailsClosedOnAnUnreadableConfig(t *testing.T) {
	e := newUpdateEnv(t, "0.1.11")
	e.serveRelease("new binary")
	if err := os.MkdirAll(e.cfgDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(e.cfgDir, "config.yaml"), []byte("profiles: [not: a map"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, _, err := e.run("update", "--yes")
	if err == nil || !strings.Contains(err.Error(), "read-only mode") {
		t.Fatalf("err = %v; want a refusal while read-only mode cannot be checked", err)
	}
	if data, _ := os.ReadFile(e.exe); string(data) != "old binary" || e.dlHits.Load() != 0 {
		t.Fatalf("binary = %q, downloads = %d; want nothing installed", data, e.dlHits.Load())
	}
}
