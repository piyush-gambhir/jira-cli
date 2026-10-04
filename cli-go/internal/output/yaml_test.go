package output

import (
	"bytes"
	"encoding/json"
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestYAMLDecodesRawJSONAndUsesJSONKeys(t *testing.T) {
	type row struct {
		Key       string          `json:"key"`
		Value     json.RawMessage `json:"value"`
		AccountID string          `json:"accountId,omitempty"`
		Empty     string          `json:"empty,omitempty"`
	}
	data := row{Key: "flag", Value: json.RawMessage(`{"enabled":true,"n":3,"s":"yes"}`), AccountID: "abc"}

	var buf bytes.Buffer
	if err := (&YAMLFormatter{Writer: &buf}).Format(data); err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := yaml.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not YAML: %v\n%s", err, buf.String())
	}
	value, ok := got["value"].(map[string]any)
	if !ok {
		t.Fatalf("value = %#v; want a decoded mapping\n%s", got["value"], buf.String())
	}
	if value["enabled"] != true || value["n"] != 3 || value["s"] != "yes" {
		t.Fatalf("value = %#v; want enabled=true n=3 s=\"yes\" (string)", value)
	}
	if got["accountId"] != "abc" {
		t.Fatalf("keys = %v; want the json tag name accountId", got)
	}
	if _, ok := got["empty"]; ok {
		t.Fatalf("omitempty field was emitted:\n%s", buf.String())
	}
}
