package zsh

import (
	"github.com/spf13/cobra"
)

func NewZshCommand() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "zsh",
	}

	cmd.AddCommand(newSourceCommand())
	cmd.AddCommand(newInstallCommand())
	cmd.AddCommand(newPluginsCommand())
	return cmd
}
