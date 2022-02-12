package out

import (
	"fmt"
	"io"

	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog/log"
)

func ColorFormat(colors tablewriter.Colors) FormatOption {
	return func() interface{} {
		return colors
	}
}

type TableOptions func(t *tablewriter.Table)

func DefaultTableOptions() (opt []TableOptions) {
	opt = append(opt, func(table *tablewriter.Table) {
		table.SetAutoWrapText(false)
		table.SetAutoFormatHeaders(true)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("")
		table.SetHeaderLine(false)
		table.SetBorder(false)
		table.SetTablePadding("\t") // pad with tabs
		table.SetNoWhiteSpace(true)
	})
	return opt
}

func NewTableSink(w io.Writer, formatter Formatter, opts ...TableOptions) Sink {
	table := tablewriter.NewWriter(w)

	for _, opt := range append(DefaultTableOptions(), opts...) {
		opt(table)
	}

	return &sink{out: &tableDataWriter{table: table}, formatter: formatter}
}

type tableDataWriter struct {
	table *tablewriter.Table
}

func (t *tableDataWriter) Write(data []string, opts ...FormatOption) (err error) {
	var colors []tablewriter.Colors
	if len(opts) > 0 {
		for i, option := range opts {
			opt := option()
			if color, ok := opt.(tablewriter.Colors); ok {
				colors = append(colors, color)
				continue
			}
			log.Debug().Str("value", data[i]).Msgf("unable to convert %T '%v' to %T", opt, opt, tablewriter.Colors{})
		}
	}
	values := data
	t.table.Rich(values, colors)
	return nil
}

func (t *tableDataWriter) Flush() (err error) {
	t.table.Render()
	return nil
}

type TableData struct {
	values []interface{}
	colors []tablewriter.Colors
}

func (t TableData) Values() (values []string) {
	for _, v := range t.values {
		values = append(values, fmt.Sprintf("%v", v))
	}
	return values
}

func (t TableData) Formatting() (formats []Format) {
	for _, c := range t.colors {
		formats = append(formats, Format(c))
	}
	return formats
}
