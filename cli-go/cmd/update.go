package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/config"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/output"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/update"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/version"
)

const (
	// releaseProject is the GoReleaser project_name (the archive name prefix).
	releaseProject = "jira-cli"
	// sourceUpdateCommand updates a binary built from source with `make install`
	// into a Go bin dir. `go install .../cli-go@latest` would install a binary
	// named cli-go, not jira, so it is not suggested.
	sourceUpdateCommand  = "git pull && make install (in your jira-cli/cli-go checkout)"
	installScriptCommand = "curl -sSfL https://raw.githubusercontent.com/" + repoSlug + "/main/install.sh | INSTALL_DIR=~/.local/bin sh"
)

// Test seams: tests point these at httptest servers and temp files.
var (
	updateBaseURL    = update.DefaultBaseURL
	updateGOOS       = runtime.GOOS
	executablePath   = currentExecutable
	stdinIsTerminal  = func() bool { return term.IsTerminal(int(os.Stdin.Fd())) }
	stderrIsTerminal = func() bool { return term.IsTerminal(int(os.Stderr.Fd())) }
)

func currentExecutable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}

func newUpdateChecker(timeout time.Duration) *update.Checker {
	return &update.Checker{
		Repo:     repoSlug,
		CacheDir: config.ConfigDir(),
		BaseURL:  updateBaseURL,
		Client:   &http.Client{Timeout: timeout},
	}
}

// installMethod is "go" when the running binary sits in a Go bin directory
// (built from source), else "self".
func installMethod() string {
	exe, err := executablePath()
	if err != nil {
		return "self"
	}
	home, _ := os.UserHomeDir()
	if update.InGoBin(exe, os.Getenv, home) {
		return "go"
	}
	return "self"
}

// updateHint is the command that updates this installation.
func updateHint(method string) string {
	if method == "go" {
		return sourceUpdateCommand
	}
	return "jira update"
}

type updateCheck struct {
	CurrentVersion  string `json:"current_version" yaml:"current_version"`
	LatestVersion   string `json:"latest_version" yaml:"latest_version"`
	UpdateAvailable bool   `json:"update_available" yaml:"update_available"`
	ReleaseURL      string `json:"release_url" yaml:"release_url"`
	InstallMethod   string `json:"install_method" yaml:"install_method"`
}

func newUpdateCmd() *cobra.Command {
	var checkOnly, yes bool
	cmd := &cobra.Command{
		Use:         "update",
		Short:       "Update jira to the latest release",
		Annotations: map[string]string{"mutates": "true"},
		Long: `Update jira to the latest GitHub release, on macOS, Linux, and Windows.

jira update downloads the release archive for this OS and architecture,
verifies its SHA-256 against the release's checksums.txt, and atomically
replaces the running binary (on Windows the old jira.exe is renamed to
jira.exe.old and deleted on a later run). If the binary's directory is not
writable, nothing changes: re-run with sudo, or reinstall with the install
script into a directory you can write to. A binary built from source into a Go
bin directory is not replaced; update the checkout instead:
  ` + sourceUpdateCommand + `

It asks "Update now? [Y/n]" when stdin is a terminal. --yes skips the prompt;
under --no-input, pass --yes. --read-only blocks installing but not --check.

--check only reports the current and latest versions (-o json for
current_version, latest_version, update_available, release_url, and
install_method) and exits 0 whether or not an update is available.

Update notice: in an interactive terminal, jira checks GitHub for a new release
at most once a day in the background and, when one exists, prints a short
notice on stderr after the command's output (once per version per day). Only
the command that starts the day's check waits for it, for at most one second.
It never runs when stderr is not a terminal, when CI is set, or with --quiet.
Turn it off with JIRA_NO_UPDATE_NOTIFIER=1 or NO_UPDATE_NOTIFIER=1.

Examples:
  jira update --check
  jira update --check -o json
  jira update
  jira update --yes`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if checkOnly {
				return runUpdateCheck(cmd)
			}
			return runUpdate(cmd, yes)
		},
	}
	cmd.Flags().BoolVar(&checkOnly, "check", false, "Only report whether an update is available; install nothing")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Install without the confirmation prompt")
	return cmd
}

func runUpdateCheck(cmd *cobra.Command) error {
	latest, err := newUpdateChecker(15*time.Second).Latest(cmd.Context(), true)
	if err != nil {
		return fmt.Errorf("checking for updates: %w", err)
	}
	res := updateCheck{
		CurrentVersion:  update.Normalize(version.Version),
		LatestVersion:   latest,
		UpdateAvailable: update.Newer(latest, version.Version),
		ReleaseURL:      update.ReleaseURL(repoSlug, latest),
		InstallMethod:   installMethod(),
	}
	if outFormat != output.FormatTable {
		return output.Print(cmd.OutOrStdout(), outFormat, res, nil)
	}
	w := cmd.OutOrStdout()
	fmt.Fprintf(w, "Current version: %s\n", displayVersion(res.CurrentVersion))
	fmt.Fprintf(w, "Latest version:  v%s\n", res.LatestVersion)
	switch {
	case res.UpdateAvailable:
		fmt.Fprintf(w, "Update available: yes (update with: %s)\n", updateHint(res.InstallMethod))
	case !update.IsRelease(version.Version):
		fmt.Fprintln(w, "Update available: unknown (development build)")
	default:
		fmt.Fprintln(w, "Update available: no")
	}
	fmt.Fprintf(w, "Release notes:   %s\n", res.ReleaseURL)
	return nil
}

func runUpdate(cmd *cobra.Command, yes bool) error {
	if err := checkUpdateReadOnly(cmd); err != nil {
		return err
	}
	current := version.Version
	if !update.IsRelease(current) {
		return fmt.Errorf("this jira is a development build (version %q) and cannot update itself; update your source checkout with %s, or install a release with: %s", current, sourceUpdateCommand, installScriptCommand)
	}
	latest, err := newUpdateChecker(15*time.Second).Latest(cmd.Context(), true)
	if err != nil {
		return fmt.Errorf("checking for updates: %w", err)
	}
	out := cmd.OutOrStdout()
	if !update.Newer(latest, current) {
		fmt.Fprintf(out, "jira v%s is already the latest version.\n", update.Normalize(current))
		return nil
	}
	fmt.Fprintf(out, "Update available: v%s -> v%s\n", update.Normalize(current), latest)

	exe, err := executablePath()
	if err != nil {
		return fmt.Errorf("finding the running jira binary: %w", err)
	}
	if installMethod() == "go" {
		fmt.Fprintf(out, "jira was built from source into %s, so it is not replaced in place.\nUpdate with: %s\n", filepath.Dir(exe), sourceUpdateCommand)
		return nil
	}

	if !yes {
		if noInputFlag {
			return errors.New("jira update needs confirmation; pass --yes to update under --no-input")
		}
		if !stdinIsTerminal() {
			return errors.New("jira update needs confirmation and stdin is not a terminal; pass --yes to update without a prompt")
		}
		fmt.Fprint(cmd.ErrOrStderr(), "Update now? [Y/n] ")
		answer, err := stdinLines().ReadString('\n')
		answer = strings.ToLower(strings.TrimSpace(answer))
		// Enter means yes; end of input (Ctrl-D) without an answer means no.
		if (err != nil && answer == "") || (answer != "" && answer != "y" && answer != "yes") {
			fmt.Fprintln(out, "Update cancelled.")
			return nil
		}
	}

	in := &update.Installer{
		Repo:        repoSlug,
		Project:     releaseProject,
		Binary:      "jira",
		GOOS:        updateGOOS,
		DownloadURL: updateBaseURL,
	}
	if !quietFlag {
		fmt.Fprintf(cmd.ErrOrStderr(), "Downloading %s ...\n", in.AssetName())
	}
	if err := in.Install(cmd.Context(), latest, exe); err != nil {
		var nw *update.NotWritableError
		if errors.As(err, &nw) {
			return fmt.Errorf("cannot replace %s: %w. The current binary was left unchanged. %s", exe, err, notWritableHint())
		}
		return fmt.Errorf("update failed: %w", err)
	}
	fmt.Fprintf(out, "Updated jira v%s -> v%s\n", update.Normalize(current), latest)
	fmt.Fprintf(out, "Release notes: %s\n", update.ReleaseURL(repoSlug, latest))
	return nil
}

func notWritableHint() string {
	if updateGOOS == "windows" {
		return "Re-run jira update from an elevated (Administrator) terminal, or download the Windows .zip from https://github.com/" + repoSlug + "/releases/latest into a directory you can write to."
	}
	return "Re-run with sudo (sudo jira update), or reinstall with the install script into a directory you can write to: " + installScriptCommand
}

// checkUpdateReadOnly blocks installing under --read-only, JIRA_READ_ONLY, or a
// profile with read_only. update skips the client setup in PersistentPreRunE
// (where read-only is normally enforced), so it resolves the profile here.
func checkUpdateReadOnly(cmd *cobra.Command) error {
	// Fail closed: a config that cannot be read could hide a read_only profile.
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config to check read-only mode: %w", err)
	}
	name := profileFlag
	if name == "" {
		name = os.Getenv("JIRA_PROFILE")
	}
	profile, err := config.ResolveAuth(config.FlagValues{}, os.LookupEnv, cfg, name)
	if err != nil {
		return fmt.Errorf("resolving the profile to check read-only mode: %w", err)
	}
	if err := checkReadOnly(cmd, profile); err != nil {
		return fmt.Errorf("%w (jira update --check still works)", err)
	}
	return nil
}

func displayVersion(v string) string {
	if update.IsRelease(v) {
		return "v" + update.Normalize(v)
	}
	return v
}

// printUpdateNotice writes the update notice to w unless it was already shown
// for latest within the last day.
func printUpdateNotice(w io.Writer, checker *update.Checker, latest string) {
	if latest == "" || !update.Newer(latest, version.Version) || !checker.ClaimNotice(latest) {
		return
	}
	fmt.Fprint(w, update.Notice("jira", repoSlug, version.Version, latest, updateHint(installMethod())))
}
