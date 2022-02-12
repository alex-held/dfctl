package cli

import (
	"io"
	"strings"
	"text/tabwriter"
	"text/template"
	"unicode/utf8"
)

type Table struct {
	w io.WriteCloser
}

func (t Table) FuncMap() template.FuncMap {
	return template.FuncMap{
		"tableRow": t.TableRow,
		"endTable": t.EndTable,
		"truncate": Truncate,
	}
}

type indentWriter struct {
	indent  string
	writer  io.Writer
	flushFn func() error
}

func (i *indentWriter) Close() error {
	return i.flushFn()
}

func (i *indentWriter) Write(p []byte) (n int, err error) {
	indented := append([]byte(i.indent), p...)
	_, err = i.writer.Write(indented)
	return len(p), err
}

func NewTable(output io.Writer) Table {
	tw := tabwriter.NewWriter(output, 30, 0, 4, ' ', 0)
	return Table{
		w: &indentWriter{
			indent:  "  ",
			writer:  tw,
			flushFn: tw.Flush,
		},
		// w: &indentWriter{
		// 	indent: "  ",
		// 	writer: tabwriter.NewWriter(output, 30, 0, 4, ' ', 0),
		// },
	}
}

func (t *Table) TableRow(fields ...string) string {
	row := strings.Join(fields, "\t")
	_, _ = t.w.Write([]byte(row + "\n"))
	return ""
}

func (t *Table) EndTable() string {
	_ = t.w.Close()
	return ""
}

func Truncate(text string, length int) string {
	const ellipsis string = "..."
	if utf8.RuneCountInString(text) > length {
		return text[:length-len(ellipsis)] + ellipsis
	}

	return text
}
