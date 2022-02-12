package zsh

import (
	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl/pkg/cli/zsh/zsh/plugins"
	"github.com/alex-held/dfctl/pkg/factory"
)

func NewZshCommand(f factory.Factory) (cmd *cobra.Command) {
	cmd = f.NewCommand("zsh",
		factory.WithHelp("interacts with the zsh configuration", ""),
		factory.WithSubcommands(newSourceCommand),
		factory.WithGroupedSubcommands("plugins", plugins.NewPluginsCommand, newInstallCommand),
	)
	return cmd
}
