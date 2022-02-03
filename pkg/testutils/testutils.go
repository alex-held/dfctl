package testutils

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	//	"github.com/sudo-bcli/color"
)

const (
	colorBlack AnsiiCode = iota + 30
	colorRed
	colorGreen
	colorYellow
	colorBlue
	colorMagenta
	colorCyan
	colorWhite

	colorDarkGray AnsiiCode = 90
)

const (
	Normal       AnsiiCode = 0
	Bold                   = 1
	Underlined             = 4
	Blinking               = 5
	ReverseVideo           = 7
)

type AnsiiCode int

func (a AnsiiCode) Paint(i interface{}) string {
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", a, i)
}

// Colorize returns the string s wrapped in ANSI code c, unless disabled is true.
func Colorize(i interface{}, opts ...AnsiiCode) (colorized string) {
	colorized = fmt.Sprintf("%v", i)
	for _, opt := range opts {
		colorized = opt.Paint(colorized)
	}
	return colorized
}

func Logger(t *testing.T) (logger zerolog.Logger) {
	w := zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.FormatFieldName = func(i interface{}) string {
			fieldName := i.(string)
			switch fieldName {
			case "testcase":
				fieldName = Colorize(fieldName, colorRed, Bold)
			default:
				fieldName = Colorize(fieldName, Bold)
			}
			return fmt.Sprintf("%s=", fieldName)
		}
		w.PartsExclude = []string{zerolog.CallerFieldName, zerolog.TimestampFieldName}
		w.PartsOrder = []string{zerolog.LevelFieldName, zerolog.MessageFieldName}
	})

	logger = zerolog.New(w)
	logger = logger.With().Str("testcase", t.Name()).Logger()
	return logger
}

func TempDir(t *testing.T, path ...string) string {
	return filepath.Join(t.TempDir(), "dfctl-zsh", t.Name(), filepath.Join(path...))
}
