package factory

import (
	"fmt"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestWithFlag(t *testing.T) {

	tt := []struct {
		name   string
		flag   string
		option FlagsOption
		want   interface{}
	}{
		{
			name: "with %T flag %s",
			flag: "foo",
			option: func(f *pflag.FlagSet) interface{} {
				return f.StringP("foo", "f", "bar", "hello world")
			},
			want: "bar",
		},
	}
	for _, tt := range tt {
		t.Run(fmt.Sprintf(tt.name, tt.want, tt.flag), func(t *testing.T) {
			factory := BuildFactory()

			command := factory.NewCommand("test", WithFlags(tt.option))

			got, err := command.Flags().GetString(tt.flag)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
