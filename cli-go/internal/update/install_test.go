package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
)

type entry struct {
	name     string
	body     string
	typeflag byte // tar.TypeReg when zero
	mode     fs.FileMode
}

func tarGz(t *testing.T, entries ...entry) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, e := range entries {
		hdr := &tar.Header{Name: e.name, Mode: 0o755, Size: int64(len(e.body)), Typeflag: e.typeflag}
		if e.typeflag == 0 {
			hdr.Typeflag = tar.TypeReg
		}
		if hdr.Typeflag == tar.TypeSymlink {
			hdr.Linkname, hdr.Size = "/etc/passwd", 0
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if hdr.Size > 0 {
			if _, err := tw.Write([]byte(e.body)); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func zipArchive(t *testing.T, entries ...entry) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range entries {
		hdr := &zip.FileHeader{Name: e.name, Method: zip.Deflate}
		mode := e.mode
		if mode == 0 {
			mode = 0o755
		}
		hdr.SetMode(mode)
		w, err := zw.CreateHeader(hdr)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(e.body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func checksumLine(name string, data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]) + "  " + name + "\n"
}

// assetServer serves files under /o/r/releases/download/v0.1.11/.
func assetServer(t *testing.T, files map[string][]byte) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		name, ok := strings.CutPrefix(r.URL.Path, "/o/r/releases/download/v0.1.11/")
		data, found := files[name]
		if !ok || !found {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(data)
	}))
	t.Cleanup(srv.Close)
	return srv, &hits
}

func writeExe(t *testing.T, dir, name, body string) string {
	t.Helper()
	exe := filepath.Join(dir, name)
	if err := os.WriteFile(exe, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return exe
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func installer(srv *httptest.Server, goos string) *Installer {
	return &Installer{Repo: "o/r", Project: "jira-cli", Binary: "jira", GOOS: goos, GOARCH: "arm64", DownloadURL: srv.URL, Client: srv.Client()}
}

func TestAssetName(t *testing.T) {
	for _, tc := range []struct{ goos, goarch, want string }{
		{"darwin", "arm64", "jira-cli_darwin_arm64.tar.gz"},
		{"linux", "amd64", "jira-cli_linux_amd64.tar.gz"},
		{"windows", "amd64", "jira-cli_windows_amd64.zip"},
	} {
		in := &Installer{Project: "jira-cli", Binary: "jira", GOOS: tc.goos, GOARCH: tc.goarch}
		if got := in.AssetName(); got != tc.want {
			t.Errorf("AssetName(%s/%s) = %q; want %q", tc.goos, tc.goarch, got, tc.want)
		}
	}
}

func TestInstallReplacesTheExecutable(t *testing.T) {
	archive := tarGz(t, entry{name: "LICENSE", body: "mit"}, entry{name: "jira", body: "new binary"})
	srv, _ := assetServer(t, map[string][]byte{
		"jira-cli_linux_arm64.tar.gz": archive,
		"checksums.txt":               []byte(checksumLine("jira-cli_linux_arm64.tar.gz", archive)),
	})
	exe := writeExe(t, t.TempDir(), "jira", "old binary")
	if err := os.Chmod(exe, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := installer(srv, "linux").Install(context.Background(), "0.1.11", exe); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, exe); got != "new binary" {
		t.Fatalf("executable = %q; want the new binary", got)
	}
	if runtime.GOOS != "windows" {
		if fi, _ := os.Stat(exe); fi.Mode().Perm() != 0o755 {
			t.Fatalf("mode = %v; want 0755", fi.Mode().Perm())
		}
	}
	if entries, _ := os.ReadDir(filepath.Dir(exe)); len(entries) != 1 {
		t.Fatalf("leftover files next to the executable: %v", entries)
	}
}

func TestInstallWindowsZip(t *testing.T) {
	archive := zipArchive(t, entry{name: "README.md", body: "readme"}, entry{name: "jira.exe", body: "new exe"})
	srv, _ := assetServer(t, map[string][]byte{
		"jira-cli_windows_arm64.zip": archive,
		"checksums.txt":              []byte(checksumLine("jira-cli_windows_arm64.zip", archive)),
	})
	exe := writeExe(t, t.TempDir(), "jira.exe", "old exe")
	if err := installer(srv, "windows").Install(context.Background(), "0.1.11", exe); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, exe); got != "new exe" {
		t.Fatalf("jira.exe = %q; want the new exe", got)
	}
	if got := readFile(t, exe+".old"); got != "old exe" {
		t.Fatalf("jira.exe.old = %q; want the old exe moved aside", got)
	}
	RemoveStaleOld(exe)
	if _, err := os.Stat(exe + ".old"); !os.IsNotExist(err) {
		t.Fatalf("jira.exe.old not removed on a later start: %v", err)
	}
}

func TestInstallRefusesBadChecksums(t *testing.T) {
	archive := tarGz(t, entry{name: "jira", body: "new binary"})
	for name, sums := range map[string]string{
		"mismatch":      checksumLine("jira-cli_linux_arm64.tar.gz", []byte("something else")),
		"missing entry": checksumLine("jira-cli_darwin_arm64.tar.gz", archive),
		"empty":         "",
	} {
		t.Run(name, func(t *testing.T) {
			srv, _ := assetServer(t, map[string][]byte{
				"jira-cli_linux_arm64.tar.gz": archive,
				"checksums.txt":               []byte(sums),
			})
			exe := writeExe(t, t.TempDir(), "jira", "old binary")
			err := installer(srv, "linux").Install(context.Background(), "0.1.11", exe)
			if err == nil || !strings.Contains(err.Error(), "refusing to install") {
				t.Fatalf("err = %v; want a refusal", err)
			}
			if got := readFile(t, exe); got != "old binary" {
				t.Fatalf("executable changed to %q", got)
			}
		})
	}
}

func TestExtractBinaryRefusesUnsafeArchives(t *testing.T) {
	const max = 1 << 20
	for name, tc := range map[string]struct {
		archive []byte
		asset   string
		binary  string
	}{
		"tar traversal":       {tarGz(t, entry{name: "../jira", body: "x"}), "a.tar.gz", "jira"},
		"tar nested escape":   {tarGz(t, entry{name: "bin/../../jira", body: "x"}), "a.tar.gz", "jira"},
		"tar absolute":        {tarGz(t, entry{name: "/usr/local/bin/jira", body: "x"}), "a.tar.gz", "jira"},
		"tar symlink":         {tarGz(t, entry{name: "jira", typeflag: tar.TypeSymlink}), "a.tar.gz", "jira"},
		"tar directory":       {tarGz(t, entry{name: "jira/", typeflag: tar.TypeDir}), "a.tar.gz", "jira"},
		"tar too large":       {tarGz(t, entry{name: "jira", body: strings.Repeat("x", max+1)}), "a.tar.gz", "jira"},
		"tar missing binary":  {tarGz(t, entry{name: "bin/jira", body: "x"}), "a.tar.gz", "jira"},
		"tar duplicate":       {tarGz(t, entry{name: "jira", body: "a"}, entry{name: "./jira", body: "b"}), "a.tar.gz", "jira"},
		"zip traversal":       {zipArchive(t, entry{name: "../jira.exe", body: "x"}), "a.zip", "jira.exe"},
		"zip backslash":       {zipArchive(t, entry{name: `..\jira.exe`, body: "x"}), "a.zip", "jira.exe"},
		"zip symlink":         {zipArchive(t, entry{name: "jira.exe", body: "target", mode: fs.ModeSymlink | 0o777}), "a.zip", "jira.exe"},
		"zip too large":       {zipArchive(t, entry{name: "jira.exe", body: strings.Repeat("x", max+1)}), "a.zip", "jira.exe"},
		"not an archive":      {[]byte("garbage"), "a.tar.gz", "jira"},
		"zip missing binary":  {zipArchive(t, entry{name: "jira", body: "x"}), "a.zip", "jira.exe"},
		"tar empty binary":    {tarGz(t, entry{name: "jira", body: ""}), "a.tar.gz", "jira"},
		"zip non-root binary": {zipArchive(t, entry{name: "dir/jira.exe", body: "x"}), "a.zip", "jira.exe"},
	} {
		t.Run(name, func(t *testing.T) {
			if got, err := ExtractBinary(tc.archive, tc.asset, tc.binary, max); err == nil {
				t.Fatalf("extracted %d bytes; want a refusal", len(got))
			}
		})
	}

	got, err := ExtractBinary(tarGz(t, entry{name: "./", typeflag: tar.TypeDir}, entry{name: "./jira", body: "ok"}), "a.tar.gz", "jira", max)
	if err != nil || string(got) != "ok" {
		t.Fatalf("ExtractBinary(./jira) = %q, %v; want ok", got, err)
	}
}

func TestReplaceExecutableWindowsRollsBackWhenTheMoveFails(t *testing.T) {
	exe := writeExe(t, t.TempDir(), "jira.exe", "old exe")
	orig := rename
	t.Cleanup(func() { rename = orig })
	rename = func(from, to string) error {
		if to == exe && !strings.HasSuffix(from, ".old") {
			return errors.New("injected: file in use")
		}
		return orig(from, to)
	}
	if err := ReplaceExecutable("windows", exe, []byte("new exe")); err == nil {
		t.Fatal("want an error when the new exe cannot be moved into place")
	}
	if got := readFile(t, exe); got != "old exe" {
		t.Fatalf("jira.exe = %q; want the old exe restored", got)
	}
	if entries, _ := os.ReadDir(filepath.Dir(exe)); len(entries) != 1 {
		t.Fatalf("leftover files after rollback: %v", entries)
	}
}

func TestReplaceExecutableReplacesAStaleOld(t *testing.T) {
	dir := t.TempDir()
	exe := writeExe(t, dir, "jira.exe", "v2")
	writeExe(t, dir, "jira.exe.old", "v1")
	if err := ReplaceExecutable("windows", exe, []byte("v3")); err != nil {
		t.Fatal(err)
	}
	if readFile(t, exe) != "v3" || readFile(t, exe+".old") != "v2" {
		t.Fatalf("jira.exe=%q jira.exe.old=%q; want v3 and v2", readFile(t, exe), readFile(t, exe+".old"))
	}
}

func TestInstallIntoUnwritableDirKeepsTheOldBinary(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("needs POSIX permissions and a non-root user")
	}
	srv, hits := assetServer(t, map[string][]byte{})
	dir := t.TempDir()
	exe := writeExe(t, dir, "jira", "old binary")
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	err := installer(srv, "linux").Install(context.Background(), "0.1.11", exe)
	var nw *NotWritableError
	if !errors.As(err, &nw) || nw.Dir != dir {
		t.Fatalf("err = %v; want a NotWritableError for %s", err, dir)
	}
	if got := readFile(t, exe); got != "old binary" {
		t.Fatalf("executable changed to %q", got)
	}
	if hits.Load() != 0 {
		t.Fatalf("downloaded %d files before finding the directory unwritable", hits.Load())
	}
	if err := ReplaceExecutable("linux", exe, []byte("new")); !errors.As(err, &nw) {
		t.Fatalf("ReplaceExecutable err = %v; want a NotWritableError", err)
	}
}

func TestRefuseHTTPSDowngrade(t *testing.T) {
	from, _ := http.NewRequest(http.MethodGet, "https://github.com/a", nil)
	to, _ := http.NewRequest(http.MethodGet, "http://example.com/a", nil)
	if err := refuseHTTPSDowngrade(to, []*http.Request{from}); err == nil {
		t.Fatal("HTTPS to HTTP redirect allowed")
	}
	ok, _ := http.NewRequest(http.MethodGet, "https://objects.githubusercontent.com/a", nil)
	if err := refuseHTTPSDowngrade(ok, []*http.Request{from}); err != nil {
		t.Fatal(err)
	}
}
