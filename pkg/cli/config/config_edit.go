package config

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl/pkg/zsh"
)

func newEditCommand() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "edit",
		RunE: func(cmd *cobra.Command, args []string) error {
			file, err := os.CreateTemp("", "dfctl-config-*.yaml")
			defer file.Close()

			if err != nil {
				return err
			}
			cfg, err := zsh.Load()
			if err != nil {
				return err
			}

			formatted, err := cfg.Format()
			if err != nil {
				return err
			}
			_, err = file.WriteString(formatted)

			if err != nil {
				return err
			}

			vimCommand := exec.Command("vim", file.Name())
			vimCommand.Stdout = os.Stdout
			vimCommand.Stderr = os.Stderr
			vimCommand.Stdin = os.Stdin

			if err = vimCommand.Start(); err != nil {
				return err
			}
			if err = vimCommand.Wait(); err != nil {
				return err
			}

			patchedCfg, err := zsh.LoadFromPath(file.Name())
			if err != nil {
				return err
			}

			err = zsh.Save(patchedCfg)
			return err
		},
	}

	return cmd
}
