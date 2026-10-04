package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestYAMLQuotesYAML11AmbiguousStrings(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{"a": "yes", "b": "On", "c": "12:34", "d": "plain", "e": "no", "n": true}
	if err := (&YAMLFormatter{Writer: &buf}).Format(data); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{`a: "yes"`, `b: "On"`, `c: "12:34"`, "d: plain", `e: "no"`, `"n": true`} {
		if !strings.Contains(out, want) {
			t.Errorf("YAML output missing %q:\n%s", want, out)
		}
	}
}
