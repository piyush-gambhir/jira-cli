package update

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

// releaseServer serves releases/latest with tag (or status when non-zero) and
// counts the requests.
func releaseServer(t *testing.T, tag string, status int) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.URL.Path != "/repos/o/r/releases/latest" {
			http.NotFound(w, r)
			return
		}
		if status != 0 {
			w.WriteHeader(status)
			return
		}
		_, _ = w.Write([]byte(`{"tag_name":"` + tag + `"}`))
	}))
	t.Cleanup(srv.Close)
	return srv, &hits
}

type clock struct{ t time.Time }

func (c *clock) now() time.Time { return c.t }

func newChecker(t *testing.T, srv *httptest.Server, c *clock) *Checker {
	return &Checker{Repo: "o/r", CacheDir: t.TempDir(), APIURL: srv.URL, Client: srv.Client(), Now: c.now}
}

func TestLatestIsCachedForADayAndForceBypassesTheCache(t *testing.T) {
	srv, hits := releaseServer(t, "v0.1.11", 0)
	c := &clock{time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)}
	ch := newChecker(t, srv, c)

	for i := 0; i < 2; i++ {
		got, err := ch.Latest(context.Background(), false)
		if err != nil || got != "0.1.11" {
			t.Fatalf("Latest = %q, %v; want 0.1.11", got, err)
		}
	}
	if hits.Load() != 1 {
		t.Fatalf("hits = %d; want 1 (second call served from the cache)", hits.Load())
	}
	if _, err := ch.Latest(context.Background(), true); err != nil || hits.Load() != 2 {
		t.Fatalf("force: err=%v hits=%d; want a fresh request", err, hits.Load())
	}
	c.t = c.t.Add(25 * time.Hour)
	if _, err := ch.Latest(context.Background(), false); err != nil || hits.Load() != 3 {
		t.Fatalf("after 25h: err=%v hits=%d; want a fresh request", err, hits.Load())
	}
}

func TestFailedCheckIsCachedToo(t *testing.T) {
	srv, hits := releaseServer(t, "", http.StatusInternalServerError)
	ch := newChecker(t, srv, &clock{time.Now()})
	for i := 0; i < 3; i++ {
		if _, err := ch.Latest(context.Background(), false); err == nil {
			t.Fatal("want an error from a failing release API")
		}
	}
	if hits.Load() != 1 {
		t.Fatalf("hits = %d; want 1 (a failure is cached, no retry storm)", hits.Load())
	}
	if _, ok := ch.CachedLatest(); ok {
		t.Fatal("CachedLatest reported a version after a failed check")
	}
}

func TestCachedLatestNeverUsesTheNetworkAndExpires(t *testing.T) {
	srv, hits := releaseServer(t, "v0.2.0", 0)
	c := &clock{time.Now()}
	ch := newChecker(t, srv, c)
	if _, ok := ch.CachedLatest(); ok {
		t.Fatal("CachedLatest with no cache reported a version")
	}
	if hits.Load() != 0 {
		t.Fatal("CachedLatest used the network")
	}
	if _, err := ch.Latest(context.Background(), false); err != nil {
		t.Fatal(err)
	}
	if v, ok := ch.CachedLatest(); !ok || v != "0.2.0" {
		t.Fatalf("CachedLatest = %q, %v; want 0.2.0", v, ok)
	}
	c.t = c.t.Add(25 * time.Hour)
	if _, ok := ch.CachedLatest(); ok {
		t.Fatal("CachedLatest reported a check older than a day")
	}
}

func TestClaimNoticeOncePerVersionPerDay(t *testing.T) {
	srv, _ := releaseServer(t, "v0.1.11", 0)
	c := &clock{time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)}
	ch := newChecker(t, srv, c)
	if !ch.ClaimNotice("0.1.11") {
		t.Fatal("first notice for 0.1.11 was suppressed")
	}
	c.t = c.t.Add(time.Hour)
	if ch.ClaimNotice("v0.1.11") {
		t.Fatal("notice for 0.1.11 shown twice within a day")
	}
	if !ch.ClaimNotice("0.1.12") {
		t.Fatal("notice for a newer version 0.1.12 was suppressed")
	}
	c.t = c.t.Add(25 * time.Hour)
	if !ch.ClaimNotice("0.1.12") {
		t.Fatal("notice for 0.1.12 suppressed after a day")
	}
	if err := ch.ClearCache(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(ch.CacheDir, CacheFileName)); !os.IsNotExist(err) {
		t.Fatalf("cache still present after ClearCache: %v", err)
	}
}

func TestVersions(t *testing.T) {
	for _, v := range []string{"0.1.10", "v1.2.3", " 10.0.0 "} {
		if !IsRelease(v) {
			t.Errorf("IsRelease(%q) = false", v)
		}
	}
	for _, v := range []string{"", "dev", "0.1", "0.1.10-3-gabc1234-dirty", "v0.1.10+meta", "01.2.3", "a.b.c"} {
		if IsRelease(v) {
			t.Errorf("IsRelease(%q) = true", v)
		}
	}
	for _, tc := range []struct {
		latest, current string
		want            bool
	}{
		{"0.1.11", "0.1.10", true},
		{"v0.2.0", "0.1.99", true},
		{"1.0.0", "0.9.9", true},
		{"0.1.10", "0.1.10", false},
		{"0.1.9", "0.1.10", false}, // never "update" to an older release
		{"0.1.11", "dev", false},
		{"", "0.1.10", false},
	} {
		if got := Newer(tc.latest, tc.current); got != tc.want {
			t.Errorf("Newer(%q, %q) = %v; want %v", tc.latest, tc.current, got, tc.want)
		}
	}
}

func TestNoticeFormat(t *testing.T) {
	got := Notice("jira", "piyush-gambhir/jira-cli", "0.1.10", "v0.1.11", "jira update")
	want := "\nA new version of jira is available: v0.1.10 -> v0.1.11\n" +
		"Update with: jira update\n" +
		"Release notes: https://github.com/piyush-gambhir/jira-cli/releases/tag/v0.1.11\n"
	if got != want {
		t.Fatalf("Notice =\n%q\nwant\n%q", got, want)
	}
}

func TestNotifierDisabledByEnv(t *testing.T) {
	env := map[string]string{}
	getenv := func(k string) string { return env[k] }
	if NotifierDisabledByEnv("JIRA", getenv) {
		t.Fatal("disabled with an empty environment")
	}
	for _, key := range []string{"CI", "JIRA_NO_UPDATE_NOTIFIER", "NO_UPDATE_NOTIFIER"} {
		env = map[string]string{key: "anything"}
		if !NotifierDisabledByEnv("JIRA", getenv) {
			t.Errorf("%s set: notifier still enabled", key)
		}
	}
}

func TestInGoBin(t *testing.T) {
	home := t.TempDir()
	gobin := t.TempDir()
	gopath1, gopath2 := t.TempDir(), t.TempDir()
	other := t.TempDir()
	for _, d := range []string{filepath.Join(home, "go", "bin"), filepath.Join(gopath2, "bin")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	env := map[string]string{"GOBIN": gobin, "GOPATH": gopath1 + string(filepath.ListSeparator) + gopath2}
	getenv := func(k string) string { return env[k] }
	for _, tc := range []struct {
		exe  string
		want bool
	}{
		{filepath.Join(gobin, "jira"), true},
		{filepath.Join(gopath2, "bin", "jira"), true},
		{filepath.Join(home, "go", "bin", "jira"), true},
		{filepath.Join(other, "jira"), false},
	} {
		if got := InGoBin(tc.exe, getenv, home); got != tc.want {
			t.Errorf("InGoBin(%s) = %v; want %v", tc.exe, got, tc.want)
		}
	}
}

func TestClaimNoticeFailsWhenItCannotBeRecorded(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("needs POSIX permissions and a non-root user")
	}
	srv, _ := releaseServer(t, "v0.1.11", 0)
	ch := newChecker(t, srv, &clock{time.Now()})
	if _, err := ch.Latest(context.Background(), false); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(ch.CacheDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(ch.CacheDir, 0o755) })
	if ch.ClaimNotice("0.1.11") {
		t.Fatal("ClaimNotice reported true without recording the claim, so the notice would repeat on every command")
	}
}
