package output

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestTableMigration(t *testing.T) {
	type item struct{ Name, State string }
	def := &TableDef{
		Headers: []string{"name", "state"},
		RowFunc: func(v interface{}) []string { i := v.(item); return []string{i.Name, i.State} },
	}
	items := []item{{"build-東京", "SUCCESS"}, {strings.Repeat("long-job-", 20), "FAILURE"}}
	for _, data := range []interface{}{items, &items, items[0], []item{}} {
		var out bytes.Buffer
		if err := (&TableFormatter{Writer: &out}).FormatTable(data, def); err != nil {
			t.Fatal(err)
		}
		text := out.String()
		if !strings.Contains(text, "NAME") || !strings.Contains(text, "STATE") {
			t.Fatalf("missing headers: %q", text)
		}
		if strings.ContainsAny(text, "│┌┐└┘+|") {
			t.Fatalf("unexpected borders: %q", text)
		}
		if strings.Contains(text, "FAILURE") && !strings.Contains(text, items[1].Name) {
			t.Fatalf("wrapped or truncated value: %q", text)
		}
	}
	if def.Headers[0] != "name" {
		t.Fatal("formatter mutated caller headers")
	}
}

func TestTableMissingDefinitionUsesJSON(t *testing.T) {
	var out bytes.Buffer
	if err := (&TableFormatter{Writer: &out}).FormatTable(map[string]string{"name": "job"}, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"name": "job"`) {
		t.Fatalf("not JSON: %q", out.String())
	}
}

type failingTableWriter struct{ err error }

func (w failingTableWriter) Write([]byte) (int, error) { return 0, w.err }

func TestTablePropagatesWriteErrors(t *testing.T) {
	want := errors.New("output closed")
	def := &TableDef{Headers: []string{"name"}, RowFunc: func(v interface{}) []string { return []string{v.(string)} }}
	err := (&TableFormatter{Writer: failingTableWriter{want}}).FormatTable([]string{"job"}, def)
	if !errors.Is(err, want) {
		t.Fatalf("got %v, want %v", err, want)
	}
}
