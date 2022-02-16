package zsh

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl/pkg/factory"
	"github.com/alex-held/dfctl/pkg/zsh"
)

func newSourceCommand(f *factory.Factory) (cmd *cobra.Command) {
	cmd = f.NewCommand("source",
		factory.WithHelp("outputs valid a generated .zshrc based on your configuration", ""),
		factory.WithAnnotationKeys("IsCore"),
	)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runSourceCommand(cmd, args)
	}
	return cmd
}

func runSourceCommand(cmd *cobra.Command, args []string) (err error) {
	source, err := zsh.Source()
	if err != nil {
		return err
	}
	_, err = os.Stdout.WriteString(source)
	return err
}
