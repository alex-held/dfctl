package zsh

import (
	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl/pkg/cli/zsh/zsh/plugins"
	"github.com/alex-held/dfctl/pkg/factory"
)

func NewZshCommand(f factory.Factory) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "zsh",
	}

	cmd.AddCommand(newSourceCommand())
	cmd.AddCommand(newInstallCommand(f))
	cmd.AddCommand(plugins.NewPluginsCommand(f))
	return cmd
}
