package out

import (
	"github.com/olekukonko/tablewriter"
)

type Formatter interface {
	Format(v interface{}) (values []string, options []FormatOption)
}

type sink struct {
	out       SinkWriter
	formatter Formatter
}

func (s *sink) WriteAndFlush(data []interface{}) (err error) {
	if err = s.Write(data); err != nil {
		return err
	}
	return s.Flush()
}

func (s *sink) Flush() (err error) {
	return s.out.Flush()
}

func (s *sink) Write(data []interface{}) (err error) {
	for _, ds := range data {
		values, options := s.formatter.Format(ds)
		if err = s.out.Write(values, options...); err != nil {
			return err
		}
	}
	return nil
}

type Format tablewriter.Colors

type FormatOption func() interface{}

type SinkWriter interface {
	Write(data []string, opts ...FormatOption) (err error)
	Flush() (err error)
}

type Sink interface {
	Write(data []interface{}) (err error)
	WriteAndFlush(data []interface{}) (err error)
	Flush() (err error)
}

type DataSet interface {
	Values() []string
	Formatting() []Format
}
