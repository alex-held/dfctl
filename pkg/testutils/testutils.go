package testutils

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	//	"github.com/sudo-bcli/color"
)

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
