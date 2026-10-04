package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/output"
)

func TestUsageErrorsHonorJSONOutput(t *testing.T) {
	for _, args := range [][]string{
		{"issue", "get", "-o", "json"},                               // missing argument
		{"issue", "list", "--bogus", "x", "--output=json"},           // unknown flag before -o
		{"-o", "json", "issue", "bogus"},                             // unknown sub-command
		{"issue", "get", "-shttps://jira.example.com", "-o", "json"}, // value glued to -s
		{"-shttps://jira.example.com", "issue", "get", "-ojson"},     // glued values everywhere
		{"-kvojson", "issue", "get"},                                 // boolean cluster ending in -o
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
		{"issue", "list", "-pdocs"},          // the o in a glued value is not -o
		{"issue", "get", "--", "-o", "json"}, // after --, -o is an argument
	} {
		if got := ErrorFormat(args); got != output.FormatTable {
			t.Errorf("ErrorFormat(%v) = %q; want table", args, got)
		}
	}
}
