package output

import (
	"bytes"
	"fmt"
	"io"
	"reflect"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

type TableFormatter struct {
	Writer io.Writer
}

func (f *TableFormatter) Format(data interface{}) error {
	return (&JSONFormatter{Writer: f.Writer}).Format(data)
}

func (f *TableFormatter) FormatTable(data interface{}, def *TableDef) error {
	if def == nil {
		return f.Format(data)
	}
	var rendered bytes.Buffer
	table := tablewriter.NewTable(&rendered,
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Formatting: tw.CellFormatting{AutoFormat: tw.On, AutoWrap: tw.WrapNone},
				Alignment:  tw.CellAlignment{Global: tw.AlignLeft},
				Padding:    tw.CellPadding{Global: tw.Padding{Right: "  "}},
			},
			Row: tw.CellConfig{
				Formatting: tw.CellFormatting{AutoWrap: tw.WrapNone},
				Alignment:  tw.CellAlignment{Global: tw.AlignLeft},
				Padding:    tw.CellPadding{Global: tw.Padding{Right: "  "}},
			},
		}),
		tablewriter.WithRendition(tw.Rendition{
			Borders:  tw.BorderNone,
			Symbols:  tw.NewSymbols(tw.StyleNone),
			Settings: tw.Settings{Lines: tw.LinesNone, Separators: tw.SeparatorsNone},
		}),
	)
	table.Header(def.Headers)

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			row := def.RowFunc(v.Index(i).Interface())
			if err := table.Append(row); err != nil {
				return err
			}
		}
	} else {
		row := def.RowFunc(data)
		if err := table.Append(row); err != nil {
			return err
		}
	}

	if err := table.Render(); err != nil {
		return err
	}
	_, err := rendered.WriteTo(f.Writer)
	return err
}

func PrintMessage(w io.Writer, msg string) {
	fmt.Fprintln(w, msg)
}
