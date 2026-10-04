package output

import (
	"encoding/json"
	"fmt"
	"io"

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

// blockStyle clears the flow and quoting styles the JSON source implies, so the
// encoder emits block YAML and quotes only where YAML needs it.
func blockStyle(n *yaml.Node) {
	n.Style = 0
	for _, child := range n.Content {
		blockStyle(child)
	}
}
