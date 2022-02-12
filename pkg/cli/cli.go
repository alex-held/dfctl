package cli

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl/pkg/factory"
)

type CLI interface {
	Execute() (err error)
}

func ConfigureLogger(levelString string) {
	level, err := zerolog.ParseLevel(levelString)
	if err != nil {
		level = zerolog.InfoLevel
	}
	w := zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.PartsExclude = []string{zerolog.CallerFieldName, zerolog.TimestampFieldName}
		w.PartsOrder = []string{zerolog.LevelFieldName, zerolog.MessageFieldName}
	})

	log.Logger = zerolog.New(w)
	zerolog.SetGlobalLevel(level)
}

func New() CLI {
	f := factory.BuildFactory()

	root := NewRootCommand(f)
	levelF := root.PersistentFlags().String("level", "info", "--level = [ trace debug info warn error fatal ]")

	root.PersistentPreRun = func(_ *cobra.Command, _ []string) {
		ConfigureLogger(*levelF)
	}

	return root
}
