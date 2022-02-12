package cli

import (
	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl/pkg/cli/config"
	"github.com/alex-held/dfctl/pkg/cli/zsh/zsh"
	"github.com/alex-held/dfctl/pkg/factory"
)

func NewRootCommand(f factory.Factory) (cmd *cobra.Command) {
	cmd = f.NewCommand("dfctl")

	cmd.AddCommand(zsh.NewZshCommand(f))
	cmd.AddCommand(config.NewConfigCommand())

	return cmd
}
