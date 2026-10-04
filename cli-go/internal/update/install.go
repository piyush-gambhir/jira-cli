package update

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	// DefaultDownloadURL is where release assets are downloaded from.
	DefaultDownloadURL = "https://github.com"

	maxArchiveSize   = 64 << 20  // release archive download limit
	maxChecksumsSize = 1 << 20   // checksums.txt download limit
	maxBinarySize    = 128 << 20 // extracted binary limit
)

// Installer downloads a GoReleaser release archive, verifies it against the
// release's checksums.txt, and replaces the running executable with the binary
// inside it.
type Installer struct {
	Repo        string // "owner/name"
	Project     string // GoReleaser project_name, the archive name prefix
	Binary      string // binary name without ".exe"
	GOOS        string // defaults to runtime.GOOS
	GOARCH      string // defaults to runtime.GOARCH
	DownloadURL string // defaults to DefaultDownloadURL
	Client      *http.Client
}

// NotWritableError means the executable's directory cannot be written, so the
// binary cannot be replaced. The old binary is left untouched.
type NotWritableError struct {
	Dir string
	Err error
}

func (e *NotWritableError) Error() string {
	return fmt.Sprintf("%s is not writable: %v", e.Dir, e.Err)
}

func (e *NotWritableError) Unwrap() error { return e.Err }

func (in *Installer) goos() string {
	if in.GOOS != "" {
		return in.GOOS
	}
	return runtime.GOOS
}

func (in *Installer) goarch() string {
	if in.GOARCH != "" {
		return in.GOARCH
	}
	return runtime.GOARCH
}

// AssetName is the release archive for the target platform, following the
// GoReleaser name_template "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}" (zip on
// Windows, tar.gz elsewhere).
func (in *Installer) AssetName() string {
	ext := ".tar.gz"
	if in.goos() == "windows" {
		ext = ".zip"
	}
	return in.Project + "_" + in.goos() + "_" + in.goarch() + ext
}

func (in *Installer) binaryName() string {
	if in.goos() == "windows" {
		return in.Binary + ".exe"
	}
	return in.Binary
}

// Install replaces exePath with the binary from release version. Any failure
// leaves the old binary in place and working.
func (in *Installer) Install(ctx context.Context, version, exePath string) error {
	if err := CheckWritable(filepath.Dir(exePath)); err != nil {
		return err
	}
	base := in.DownloadURL
	if base == "" {
		base = DefaultDownloadURL
	}
	releaseURL := strings.TrimRight(base, "/") + "/" + in.Repo + "/releases/download/v" + Normalize(version) + "/"
	asset := in.AssetName()

	archive, err := in.download(ctx, releaseURL+asset, maxArchiveSize)
	if err != nil {
		return fmt.Errorf("downloading %s: %w", asset, err)
	}
	checksums, err := in.download(ctx, releaseURL+"checksums.txt", maxChecksumsSize)
	if err != nil {
		return fmt.Errorf("downloading checksums.txt: %w", err)
	}
	if err := VerifyChecksum(checksums, asset, archive); err != nil {
		return err
	}
	bin, err := ExtractBinary(archive, asset, in.binaryName(), maxBinarySize)
	if err != nil {
		return err
	}
	return ReplaceExecutable(in.goos(), exePath, bin)
}

func (in *Installer) download(ctx context.Context, url string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	client := in.Client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Minute, CheckRedirect: refuseHTTPSDowngrade}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned %s", url, resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("%s is larger than %d bytes", url, limit)
	}
	return data, nil
}

// refuseHTTPSDowngrade keeps release downloads on HTTPS across redirects.
func refuseHTTPSDowngrade(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}
	if via[0].URL.Scheme == "https" && req.URL.Scheme != "https" {
		return fmt.Errorf("refusing redirect from HTTPS to %s", req.URL.Redacted())
	}
	return nil
}

// VerifyChecksum checks data against the SHA-256 that checksums.txt lists for
// asset. A missing entry is an error, not a pass.
func VerifyChecksum(checksums []byte, asset string, data []byte) error {
	want := ""
	sc := bufio.NewScanner(bytes.NewReader(checksums))
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) == 2 && strings.TrimPrefix(fields[1], "*") == asset {
			want = strings.ToLower(fields[0])
			break
		}
	}
	if err := sc.Err(); err != nil {
		return fmt.Errorf("reading checksums.txt: %w", err)
	}
	if len(want) != sha256.Size*2 {
		return fmt.Errorf("checksums.txt has no SHA-256 for %s; refusing to install", asset)
	}
	sum := sha256.Sum256(data)
	if got := hex.EncodeToString(sum[:]); got != want {
		return fmt.Errorf("checksum mismatch for %s (expected %s, got %s); refusing to install", asset, want, got)
	}
	return nil
}

// ExtractBinary returns the contents of the regular file named binary at the
// root of the archive (a .zip or .tar.gz, chosen by asset's extension). It
// refuses archives with absolute or parent-relative entry names, a binary entry
// that is not a regular file, and a binary larger than maxSize. Nothing is
// written to disk, so no entry name ever reaches the file system.
func ExtractBinary(archive []byte, asset, binary string, maxSize int64) ([]byte, error) {
	if strings.HasSuffix(asset, ".zip") {
		return extractZip(archive, binary, maxSize)
	}
	return extractTarGz(archive, binary, maxSize)
}

// entryName validates an archive entry name and returns it cleaned.
func entryName(name string) (string, error) {
	if strings.Contains(name, `\`) || path.IsAbs(name) || filepath.IsAbs(name) || filepath.VolumeName(name) != "" {
		return "", fmt.Errorf("archive entry %q has an unsafe path; refusing to install", name)
	}
	clean := path.Clean(name)
	if clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("archive entry %q escapes the archive; refusing to install", name)
	}
	return clean, nil
}

func readLimited(r io.Reader, maxSize int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxSize+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxSize {
		return nil, fmt.Errorf("binary is larger than %d bytes; refusing to install", maxSize)
	}
	return data, nil
}

func extractTarGz(archive []byte, binary string, maxSize int64) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, fmt.Errorf("opening the archive: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	var found []byte
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading the archive: %w", err)
		}
		name, err := entryName(hdr.Name)
		if err != nil {
			return nil, err
		}
		if name != binary {
			continue
		}
		if hdr.Typeflag != tar.TypeReg || found != nil {
			return nil, fmt.Errorf("archive entry %q is not a single regular file; refusing to install", hdr.Name)
		}
		if hdr.Size > maxSize {
			return nil, fmt.Errorf("binary is larger than %d bytes; refusing to install", maxSize)
		}
		if found, err = readLimited(tr, maxSize); err != nil {
			return nil, err
		}
	}
	return checkFound(found, binary)
}

func extractZip(archive []byte, binary string, maxSize int64) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, fmt.Errorf("opening the archive: %w", err)
	}
	var found []byte
	for _, f := range zr.File {
		name, err := entryName(f.Name)
		if err != nil {
			return nil, err
		}
		if name != binary {
			continue
		}
		if !f.Mode().IsRegular() || found != nil {
			return nil, fmt.Errorf("archive entry %q is not a single regular file; refusing to install", f.Name)
		}
		if f.UncompressedSize64 > uint64(maxSize) {
			return nil, fmt.Errorf("binary is larger than %d bytes; refusing to install", maxSize)
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("reading the archive: %w", err)
		}
		found, err = readLimited(rc, maxSize)
		rc.Close()
		if err != nil {
			return nil, err
		}
	}
	return checkFound(found, binary)
}

func checkFound(found []byte, binary string) ([]byte, error) {
	if len(found) == 0 {
		return nil, fmt.Errorf("%s not found in the archive (or empty); refusing to install", binary)
	}
	return found, nil
}

// CheckWritable reports a *NotWritableError when files cannot be created in dir.
func CheckWritable(dir string) error {
	f, err := os.CreateTemp(dir, ".update-probe-*")
	if err != nil {
		return &NotWritableError{Dir: dir, Err: err}
	}
	f.Close()
	_ = os.Remove(f.Name())
	return nil
}

// rename is a test seam for the Windows rollback path.
var rename = os.Rename

// ReplaceExecutable atomically replaces exePath with data. The new binary is
// written to a temp file in the same directory (so the final rename stays on
// one file system), made executable, and renamed over exePath. On Windows a
// running .exe cannot be overwritten but can be renamed, so exePath is first
// moved aside to exePath+".old" (removed on a later start by RemoveStaleOld)
// and restored if the new binary cannot be moved into place.
func ReplaceExecutable(goos, exePath string, data []byte) error {
	dir := filepath.Dir(exePath)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(exePath)+".new-*")
	if err != nil {
		return &NotWritableError{Dir: dir, Err: err}
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath) // no-op once renamed into place
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("writing the new binary: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("writing the new binary: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("writing the new binary: %w", err)
	}
	if err := os.Chmod(tmpPath, 0o755); err != nil {
		return fmt.Errorf("making the new binary executable: %w", err)
	}

	if goos != "windows" {
		if err := rename(tmpPath, exePath); err != nil {
			return replaceError(dir, err)
		}
		return nil
	}

	old := exePath + ".old"
	_ = os.Remove(old) // a leftover from an earlier update
	if err := rename(exePath, old); err != nil {
		return replaceError(dir, err)
	}
	if err := rename(tmpPath, exePath); err != nil {
		if rerr := rename(old, exePath); rerr != nil {
			return fmt.Errorf("moving the new binary into place: %w (restoring the old binary also failed: %v; it is at %s)", err, rerr, old)
		}
		return replaceError(dir, err)
	}
	return nil
}

func replaceError(dir string, err error) error {
	if errors.Is(err, fs.ErrPermission) {
		return &NotWritableError{Dir: dir, Err: err}
	}
	return fmt.Errorf("replacing the binary: %w", err)
}

// RemoveStaleOld deletes exePath+".old" left by a Windows update, best-effort.
func RemoveStaleOld(exePath string) {
	_ = os.Remove(exePath + ".old")
}

// InGoBin reports whether exePath is in a Go bin directory ($GOBIN,
// $GOPATH/bin for each GOPATH entry, or ~/go/bin), i.e. it was built from
// source and should be updated the same way rather than replaced in place.
func InGoBin(exePath string, getenv func(string) string, home string) bool {
	var dirs []string
	if gobin := getenv("GOBIN"); gobin != "" {
		dirs = append(dirs, gobin)
	}
	for _, p := range filepath.SplitList(getenv("GOPATH")) {
		if p != "" {
			dirs = append(dirs, filepath.Join(p, "bin"))
		}
	}
	if home != "" {
		dirs = append(dirs, filepath.Join(home, "go", "bin"))
	}
	exeDir := canonicalDir(filepath.Dir(exePath))
	for _, d := range dirs {
		if samePath(exeDir, canonicalDir(d)) {
			return true
		}
	}
	return false
}

func canonicalDir(dir string) string {
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	if real, err := filepath.EvalSymlinks(dir); err == nil {
		dir = real
	}
	return filepath.Clean(dir)
}

func samePath(a, b string) bool {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		return strings.EqualFold(a, b)
	}
	return a == b
}
