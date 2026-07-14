package output

import (
	"fmt"
	"io"
	"reflect"

	"github.com/olekukonko/tablewriter"
)

// TableFormatter outputs data as an ASCII table.
type TableFormatter struct {
	Writer io.Writer
}

// Format outputs data as a table (falls back to JSON without a TableDef).
func (f *TableFormatter) Format(data interface{}) error {
	return (&JSONFormatter{Writer: f.Writer}).Format(data)
}

// FormatTable outputs data as a table using the provided TableDef.
func (f *TableFormatter) FormatTable(data interface{}, def *TableDef) error {
	table := tablewriter.NewWriter(f.Writer)
	table.SetHeader(def.Headers)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(true)

	// Handle slice types
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			row := def.RowFunc(v.Index(i).Interface())
			table.Append(row)
		}
	} else {
		// Single item
		row := def.RowFunc(data)
		table.Append(row)
	}

	table.Render()
	return nil
}

// PrintMessage prints a simple message to the writer.
func PrintMessage(w io.Writer, msg string) {
	fmt.Fprintln(w, msg)
}
