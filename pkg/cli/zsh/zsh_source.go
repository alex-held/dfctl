package zsh

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl-zsh/pkg/zsh"
)

func newSourceCommand() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "source",
		RunE: func(cmd *cobra.Command, args []string) error {
			source, err := zsh.Source()
			if err != nil {
				return err
			}
			_, err = os.Stdout.WriteString(source)
			return err
		},
	}

	return cmd
}
