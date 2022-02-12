package out

import (
	"bytes"
	"io"
)

func NewListSink(writer io.Writer, formatter Formatter) Sink {
	return &sink{out: &listWriter{writer: writer, buf: &bytes.Buffer{}}, formatter: formatter}
}

type listWriter struct {
	writer io.Writer
	buf    *bytes.Buffer
}

func (w *listWriter) Write(data []string, opts ...FormatOption) (err error) {
	for _, v := range data {
		if _, err = w.buf.WriteString(v + "\n"); err != nil {
			return err
		}
	}
	return nil
}

func (w *listWriter) Flush() (err error) {
	_, err = w.buf.WriteTo(w.writer)
	return err
}
