package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/output"
)

func TestUsageErrorsHonorJSONOutput(t *testing.T) {
	for _, args := range [][]string{
		{"issue", "get", "-o", "json"},
		{"issue", "list", "--bogus", "x", "--output=json"},
		{"-o", "json", "issue", "bogus"},
		{"issue", "get", "-shttps://jira.example.com", "-o", "json"},
		{"-shttps://jira.example.com", "issue", "get", "-ojson"},
		{"-kvojson", "issue", "get"},
		{"issue", "list", "-pPROJ", "-n5", "--output=json", "--bogus"},
	} {
		resetRootFlags(t)
		rootCmd.SetArgs(args)
		err := rootCmd.Execute()
		if err == nil {
			t.Fatalf("%v: expected a usage error", args)
		}
		if got := ErrorFormat(args); got != output.FormatJSON {
			t.Fatalf("%v: ErrorFormat = %q; want json", args, got)
		}
		var buf bytes.Buffer
		output.WriteError(&buf, ErrorFormat(args), err, 0)
		var payload output.ErrorResponse
		if jerr := json.Unmarshal(buf.Bytes(), &payload); jerr != nil || payload.Error == "" {
			t.Fatalf("%v: error output %q is not a JSON error object (%v)", args, buf.String(), jerr)
		}
	}
}

func TestErrorFormatDefaultsToTable(t *testing.T) {
	resetRootFlags(t)
	for _, args := range [][]string{
		nil, {"issue", "get"}, {"-o", "bogus"},
		{"issue", "list", "-pdocs"},
		{"issue", "get", "--", "-o", "json"},
	} {
		if got := ErrorFormat(args); got != output.FormatTable {
			t.Errorf("ErrorFormat(%v) = %q; want table", args, got)
		}
	}
}
