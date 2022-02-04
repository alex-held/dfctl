package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl/pkg/dfpath"
	"github.com/alex-held/dfctl/pkg/zsh"
)

func newViewCommand() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "view",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := zsh.Load()
			if err != nil {
				return err
			}
			formatted, err := cfg.Format(func(f *zsh.ConfigFormatter) {
				f.ConfigFileType = filepath.Ext(dfpath.ConfigFile())
			})
			if err != nil {
				return err
			}
			_, err = os.Stdout.WriteString(formatted)
			return err
		},
	}

	return cmd
}
