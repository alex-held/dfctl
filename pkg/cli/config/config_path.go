package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl/pkg/dfpath"
	"github.com/alex-held/dfctl/pkg/factory"
)

func newPathCommand(f factory.Factory) (cmd *cobra.Command) {
	cmd = f.NewCommand("path",
		factory.WithHelp("view the current configuation file path", "displays a the full path of the $DFCTL_CONFIG file"),
	)
	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		path := dfpath.ConfigFile()
		_, err = fmt.Fprintln(cmd.OutOrStdout(), path)
		return err
	}
	return cmd
}
