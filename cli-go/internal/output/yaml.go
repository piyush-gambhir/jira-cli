package output

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"

	"go.yaml.in/yaml/v3"
)

type YAMLFormatter struct {
	Writer io.Writer
}

func (f *YAMLFormatter) Format(data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling YAML: %w", err)
	}
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

var yaml11Ambiguous = regexp.MustCompile(`^(?:y|Y|yes|Yes|YES|n|N|no|No|NO|on|On|ON|off|Off|OFF|true|True|TRUE|false|False|FALSE|[-+]?[0-9][0-9_]*(?::[0-5]?[0-9])+(?:\.[0-9_]*)?)$`)

func blockStyle(n *yaml.Node) {
	n.Style = 0
	if n.Kind == yaml.ScalarNode && n.Tag == "!!str" && yaml11Ambiguous.MatchString(n.Value) {
		n.Style = yaml.DoubleQuotedStyle
	}
	for _, child := range n.Content {
		blockStyle(child)
	}
}
