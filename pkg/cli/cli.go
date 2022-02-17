package cli

import (
	"os"

	"github.com/alex-held/dfctl-kit/pkg/dflog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/alex-held/dfctl/pkg/extensions"
	"github.com/alex-held/dfctl/pkg/factory"
)

type cli struct {
	RootCmd *cobra.Command
	factory *factory.Factory
}

func (c *cli) Execute() (err error) {

	if hasCommand(c.RootCmd, os.Args) {
		log.Info().Msgf("has command %v", os.Args[1])
	} else {

		em := extensions.NewManager(c.factory)
		log.Info().Msgf("does not have command %v", os.Args[1])

		for _, extension := range em.List(true) {
			if extension.Name() == os.Args[1] {
				log.Info().Msgf("extension found %v at path %v", extension.Name(), extension.Path())
				ok, err := em.Dispatch(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)

				if err != nil {
					log.Error().Err(err).Msgf("failed to dispatch to extension %v -- %v", extension.Name(), os.Args)
					return err
				}

				if !ok {
					log.Error().Msgf("dispatch to extension %v unsuccessful with args %v", extension.Name(), os.Args)
				}

				return nil
			}
		}

		return nil
	}

	return c.RootCmd.Execute()
}

func hasCommand(rootCmd *cobra.Command, args []string) bool {
	c, _, err := rootCmd.Traverse(args)
	return err == nil && c != rootCmd
}

type CLI interface {
	Execute() (err error)
}

func New() CLI {
	logging()

	c := &cli{
		factory: factory.BuildFactory(),
	}
	c.RootCmd = NewRootCommand(c.factory)
	return c
}

func logging() {
	flags := pflag.NewFlagSet("logging", pflag.ContinueOnError)
	level, err := flags.GetString("level")
	if err != nil {
		dflog.ConfigureWithLevel(zerolog.DebugLevel)
		return
	}
	dflog.ConfigureWithLevelString(level)
}
