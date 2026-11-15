package output

import (
	"fmt"
	"io"
	"os"
)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
)

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

type Formatter interface {
	Format(data interface{}) error
}

type TableDef struct {
	Headers []string
	RowFunc func(item interface{}) []string
}

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
