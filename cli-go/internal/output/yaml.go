package output

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// YAMLFormatter outputs data as YAML.
type YAMLFormatter struct {
	Writer io.Writer
}

// Format outputs data as YAML.
func (f *YAMLFormatter) Format(data interface{}) error {
	b, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling YAML: %w", err)
	}
	_, err = fmt.Fprint(f.Writer, string(b))
	return err
}
