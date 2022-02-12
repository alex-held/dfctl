package config

import (
	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl/pkg/factory"
)

func NewConfigCommand(f factory.Factory) (cmd *cobra.Command) {
	cmd = f.NewCommand("config [command]",
		factory.WithSubcommands(
			newViewCommand,
			newPathCommand,
			newEditCommand,
		),
		factory.WithHelp("dfctl config actions", "interact with the current dfctl config"),
	)

	cmd.AddCommand(newViewCommand(f))
	cmd.AddCommand(newPathCommand(f))
	cmd.AddCommand(newEditCommand(f))
	return cmd
}
