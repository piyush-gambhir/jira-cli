package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// JSONFormatter outputs data as pretty-printed JSON.
type JSONFormatter struct {
	Writer io.Writer
}

// Format outputs data as indented JSON.
func (f *JSONFormatter) Format(data interface{}) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	_, err = fmt.Fprintln(f.Writer, string(b))
	return err
}
