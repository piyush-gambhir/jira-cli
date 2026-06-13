package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/piyush-gambhir/jira-cli/internal/output"
)

// stderr returns the stderr writer (indirection keeps prompts testable).
func stderr() io.Writer { return os.Stderr }

// nowUnix returns the current unix time in seconds.
func nowUnix() int64 { return time.Now().Unix() }

// render prints data in the active output format (table needs a TableDef).
func render(data any, def *output.TableDef) error {
	return output.Print(os.Stdout, outFormat, data, def)
}

// info prints an informational message to stderr unless --quiet (keeps stdout
// clean for JSON/YAML consumers).
func info(format string, args ...any) {
	if !quietFlag {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

// prompt reads a line from stdin after writing a label to stderr.
func prompt(label string) (string, error) {
	if noInputFlag {
		return "", fmt.Errorf("interactive input required but --no-input is set")
	}
	fmt.Fprint(os.Stderr, label)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// dash returns "-" for empty strings, for tidy table cells.
func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// truncate shortens s to n runes with an ellipsis.
func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}
