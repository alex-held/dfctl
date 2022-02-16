package run

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl/pkg/extensions"
	"github.com/alex-held/dfctl/pkg/factory"
)

func NewRunCommand(f *factory.Factory) *cobra.Command {
	cmd := f.NewCommand("run")
	cmd.Args = cobra.MinimumNArgs(1)
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		em := extensions.NewManager(f)
		ok, err := em.Dispatch(args, cmd.InOrStdin(), os.Stdout, os.Stderr)
		if !ok || err != nil {
			log.Error().Err(err).Msgf("failed to dispatch args '%v' to extension %s", args[0], args[1:])
		}
		return nil
	}
	return cmd
}
