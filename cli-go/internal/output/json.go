package output

import (
	"encoding/json"
	"fmt"
	"io"
)

type JSONFormatter struct {
	Writer io.Writer
}

func (f *JSONFormatter) Format(data interface{}) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	_, err = fmt.Fprintln(f.Writer, string(b))
	return err
}
