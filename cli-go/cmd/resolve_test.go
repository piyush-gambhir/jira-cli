package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/output"
	"github.com/spf13/cobra"
)

func TestCommandsResolve(t *testing.T) {
	resetRootFlags(t)
	// Invalid configuration must never be read by resolution.
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	configDir := filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "jira-cli")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("[invalid YAML"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name    string
		args    []string
		command string
		effect  string
		flags   map[string]string
		wantArg []string
	}{
		{"read", []string{"issue", "get", "PROJ-123"}, "jira issue get", "read", map[string]string{}, []string{"PROJ-123"}},
		{"write", []string{"-o", "json", "issue", "delete", "PROJ-123", "--yes"}, "jira issue delete", "remote_write", map[string]string{"output": "json", "yes": "true"}, []string{"PROJ-123"}},
		{"alias", []string{"auth", "ls"}, "jira auth list", "read", map[string]string{}, []string{}},
		{"target output", []string{"-o", "yaml", "issue", "get", "PROJ-123", "--fields", "summary"}, "jira issue get", "read", map[string]string{"output": "yaml", "fields": "summary"}, []string{"PROJ-123"}},
		{"target separator", []string{"issue", "get", "--", "--yes"}, "jira issue get", "read", map[string]string{}, []string{"--yes"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			root := newRootCmd()
			// Flags on the running tree must not leak into the fresh probe.
			target, rest, err := root.Find([]string{"issue", "get", "PROJ-999", "--fields", "stale", "--site", "https://stale.example.com"})
			if err != nil {
				t.Fatal(err)
			}
			if err := target.ParseFlags(rest); err != nil {
				t.Fatal(err)
			}
			root.SetOut(&out)
			root.SetErr(&bytes.Buffer{})
			root.SetArgs(append([]string{"commands", "resolve", "--"}, tc.args...))
			if err := root.Execute(); err != nil {
				t.Fatalf("%v: %v", tc.args, err)
			}
			var got resolution
			if err := json.Unmarshal(out.Bytes(), &got); err != nil {
				t.Fatalf("invalid JSON %q: %v", out.String(), err)
			}
			want := resolution{Command: tc.command, Effect: tc.effect, Flags: tc.flags, Args: tc.wantArg}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %+v; want %+v", got, want)
			}
			if OutputFormat != "json" || outFormat != output.FormatJSON {
				t.Fatal("resolution did not force JSON output")
			}
		})
	}
	for _, args := range [][]string{{"nonsense"}, {}, {"issue"}, {"issue", "nonsense"}, {"commands"}, {"issue", "get", "--bogus"}} {
		root := newRootCmd()
		root.SetOut(&bytes.Buffer{})
		root.SetErr(&bytes.Buffer{})
		root.SetArgs(append([]string{"commands", "resolve", "--"}, args...))
		if err := root.Execute(); err == nil {
			t.Fatalf("%v: expected an error", args)
		}
	}
}

func TestResolveAnnotations(t *testing.T) {
	for _, tc := range []struct {
		name        string
		annotations map[string]string
		args        []string
		effect      string
		conditional bool
	}{
		{"remote", map[string]string{"mutates": "true", "writes-local": "true"}, nil, "remote_write", false},
		{"conditional remote", map[string]string{"mutates": "conditional"}, nil, "remote_write", true},
		{"remote flag absent", map[string]string{"mutates-when-flag": "apply", "writes-local": "true"}, nil, "read", false},
		{"remote flag set false", map[string]string{"mutates-when-flag": "apply"}, []string{"--apply=false"}, "remote_write", false},
		{"local", map[string]string{"writes-local": "true"}, nil, "local_write", false},
		{"conditional local", map[string]string{"writes-local": "conditional"}, nil, "local_write", true},
		{"self update", map[string]string{"self-update": "true"}, nil, "local_write", false},
		{"local flag absent", map[string]string{"writes-local-when-flag": "apply"}, nil, "read", false},
		{"local flag set", map[string]string{"writes-local-when-flag": "apply"}, []string{"--apply"}, "local_write", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := &cobra.Command{Use: "sample", Annotations: tc.annotations}
			target.Annotations["interactive"] = "true"
			target.Annotations["secret-output"] = "true"
			target.Flags().Bool("apply", false, "Apply changes")
			if err := target.ParseFlags(tc.args); err != nil {
				t.Fatal(err)
			}
			got := describeResolution(target)
			if got.Effect != tc.effect || got.Conditional != tc.conditional || !got.Interactive || !got.Secret {
				t.Fatalf("unexpected resolution: %+v", got)
			}
		})
	}
}
