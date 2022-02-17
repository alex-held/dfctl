package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl-kit/pkg/env"

	"github.com/alex-held/dfctl/pkg/factory"
	"github.com/alex-held/dfctl/pkg/zsh"
)

func newViewCommand(f *factory.Factory) (cmd *cobra.Command) {
	cmd = f.NewCommand("view",
		factory.WithHelp("view the current configuation", "displays a formatted version of the $DFCTL_CONFIG file"),
	)
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := zsh.Load()
		if err != nil {
			return err
		}
		formatted, err := cfg.Format(func(f *zsh.ConfigFormatter) {
			f.ConfigFileType = filepath.Ext(env.ConfigFile())
		})
		if err != nil {
			return err
		}
		_, err = os.Stdout.WriteString(formatted)
		return err
	}

	return cmd
}
