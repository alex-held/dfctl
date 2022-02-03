package cli

import (
	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl-zsh/pkg/cli/config"
	"github.com/alex-held/dfctl-zsh/pkg/cli/zsh"
)

func NewRootCommand() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "dfctl",
	}

	cmd.AddCommand(zsh.NewZshCommand())
	cmd.AddCommand(config.NewConfigCommand())

	return cmd
}
