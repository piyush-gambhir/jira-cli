package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/adf"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/output"
	"golang.org/x/term"
)

func stderr() io.Writer { return os.Stderr }

func nowUnix() int64 { return time.Now().Unix() }

func render(data any, def *output.TableDef) error {
	return output.Print(os.Stdout, outFormat, data, def)
}

func info(format string, args ...any) {
	if !quietFlag {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

var (
	stdinReader     *bufio.Reader
	stdinReaderFile *os.File
)

func stdinLines() *bufio.Reader {
	if stdinReader == nil || stdinReaderFile != os.Stdin {
		stdinReader, stdinReaderFile = bufio.NewReader(os.Stdin), os.Stdin
	}
	return stdinReader
}

func prompt(label string) (string, error) {
	if noInputFlag {
		return "", fmt.Errorf("interactive input required but --no-input is set")
	}
	fmt.Fprint(os.Stderr, label)
	line, err := stdinLines().ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func promptSecret(label string) (string, error) {
	if noInputFlag {
		return "", fmt.Errorf("interactive input required but --no-input is set")
	}
	fmt.Fprint(os.Stderr, label)
	if term.IsTerminal(int(os.Stdin.Fd())) {
		secret, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		return strings.TrimSpace(string(secret)), err
	}
	line, err := stdinLines().ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

func richText(text string, markdown bool) (any, error) {
	if jiraClient.APIVer() == "2" {
		if markdown {
			return nil, fmt.Errorf("--markdown needs REST API v3 (Jira Cloud); on API v2 (Server/Data Center) the text is sent as Jira wiki markup, so drop --markdown")
		}
		return text, nil
	}
	if markdown {
		return adf.FromMarkdown(text), nil
	}
	return adf.FromPlainText(text), nil
}
