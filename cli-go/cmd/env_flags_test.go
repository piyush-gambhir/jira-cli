package cmd

import (
	"testing"

	"github.com/spf13/pflag"
)

func TestTruthy(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  bool
	}{{"true", true}, {"TRUE", true}, {"1", true}, {"false", false}, {"0", false}, {"anything", false}} {
		if got := truthy(tc.value); got != tc.want {
			t.Errorf("truthy(%q) = %v, want %v", tc.value, got, tc.want)
		}
	}
}

// resetRootFlags restores the global flags (values and Changed state) so tests
// that execute rootCmd do not leak into each other.
func resetRootFlags(t *testing.T) {
	t.Helper()
	reset := func() {
		rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
			_ = f.Value.Set(f.DefValue)
			f.Changed = false
		})
		OutputFormat = ""
		rootCmd.SetArgs(nil)
	}
	reset()
	t.Cleanup(reset)
}

func TestExplicitFalseFlagsOverrideEnv(t *testing.T) {
	t.Setenv("JIRA_QUIET", "1")
	t.Setenv("JIRA_NO_INPUT", "true")

	resetRootFlags(t)
	rootCmd.SetArgs([]string{"version"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !quietFlag || !noInputFlag {
		t.Fatalf("env fallback: quiet=%v no-input=%v; want both true", quietFlag, noInputFlag)
	}

	resetRootFlags(t)
	rootCmd.SetArgs([]string{"version", "--quiet=false", "--no-input=false"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if quietFlag || noInputFlag {
		t.Fatalf("explicit false flags: quiet=%v no-input=%v; want both false", quietFlag, noInputFlag)
	}
}
