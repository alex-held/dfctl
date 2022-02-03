package config

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl/pkg/dfpath"
)

func newPathCommand() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "path",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			path := dfpath.ConfigFile()
			_, err = os.Stdout.WriteString(path)
			return err
		},
	}

	return cmd
}
