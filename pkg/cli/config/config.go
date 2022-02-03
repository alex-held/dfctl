package config

import (
	"github.com/spf13/cobra"
)

func NewConfigCommand() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "config",
	}

	cmd.AddCommand(newViewCommand())
	cmd.AddCommand(newPathCommand())
	cmd.AddCommand(newEditCommand())
	return cmd
}
