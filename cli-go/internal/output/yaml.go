package output

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"

	"go.yaml.in/yaml/v3"
)

// YAMLFormatter outputs data as YAML.
type YAMLFormatter struct {
	Writer io.Writer
}

// Format outputs data as YAML. Data goes through JSON first so YAML matches
// -o json: keys follow the json tags, custom JSON marshalers apply, and
// json.RawMessage values print decoded instead of as lists of bytes.
func (f *YAMLFormatter) Format(data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling YAML: %w", err)
	}
	// Decoding into a yaml.Node keeps the JSON key order.
	var node yaml.Node
	if err := yaml.Unmarshal(b, &node); err != nil {
		return fmt.Errorf("marshaling YAML: %w", err)
	}
	blockStyle(&node)
	out, err := yaml.Marshal(&node)
	if err != nil {
		return fmt.Errorf("marshaling YAML: %w", err)
	}
	_, err = fmt.Fprint(f.Writer, string(out))
	return err
}

// yaml11Ambiguous matches strings that YAML 1.1 readers such as PyYAML resolve
// to booleans or base-60 numbers when unquoted; yaml.v3 (YAML 1.2) leaves them plain.
var yaml11Ambiguous = regexp.MustCompile(`^(?:y|Y|yes|Yes|YES|n|N|no|No|NO|on|On|ON|off|Off|OFF|true|True|TRUE|false|False|FALSE|[-+]?[0-9][0-9_]*(?::[0-5]?[0-9])+(?:\.[0-9_]*)?)$`)

// blockStyle clears the flow and quoting styles the JSON source implies, so the
// encoder emits block YAML and quotes only where YAML needs it.
func blockStyle(n *yaml.Node) {
	n.Style = 0
	if n.Kind == yaml.ScalarNode && n.Tag == "!!str" && yaml11Ambiguous.MatchString(n.Value) {
		n.Style = yaml.DoubleQuotedStyle
	}
	for _, child := range n.Content {
		blockStyle(child)
	}
}
