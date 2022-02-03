package config

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl-zsh/pkg/config"
)

func newViewCommand() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "view",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			toml, err := cfg.Toml()
			if err != nil {
				return err
			}
			_, err = os.Stdout.WriteString(toml)
			return err
		},
	}

	return cmd
}
