package output

import (
	"fmt"
	"io"
	"os"
)

// Format represents the output format type.
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
)

// ParseFormat parses a string into a Format.
func ParseFormat(s string) (Format, error) {
	switch s {
	case "table", "":
		return FormatTable, nil
	case "json":
		return FormatJSON, nil
	case "yaml":
		return FormatYAML, nil
	default:
		return "", fmt.Errorf("unsupported output format: %s (use table, json, or yaml)", s)
	}
}

// Formatter is the interface for output formatting.
type Formatter interface {
	Format(data interface{}) error
}

// TableDef defines a table layout.
type TableDef struct {
	Headers []string
	RowFunc func(item interface{}) []string
}

// NewFormatter creates a formatter for the given format.
func NewFormatter(format Format, w io.Writer) Formatter {
	if w == nil {
		w = os.Stdout
	}
	switch format {
	case FormatJSON:
		return &JSONFormatter{Writer: w}
	case FormatYAML:
		return &YAMLFormatter{Writer: w}
	default:
		return &TableFormatter{Writer: w}
	}
}

// Print formats and prints data according to the format and optional table definition.
func Print(w io.Writer, format Format, data interface{}, tableDef *TableDef) error {
	switch format {
	case FormatJSON:
		return (&JSONFormatter{Writer: w}).Format(data)
	case FormatYAML:
		return (&YAMLFormatter{Writer: w}).Format(data)
	default:
		if tableDef == nil {
			return (&JSONFormatter{Writer: w}).Format(data)
		}
		return (&TableFormatter{Writer: w}).FormatTable(data, tableDef)
	}
}
