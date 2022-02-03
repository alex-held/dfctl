package config

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl/pkg/config"
	"github.com/alex-held/dfctl/pkg/dfpath"
)

func newEditCommand() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "edit",
		RunE: func(cmd *cobra.Command, args []string) error {
			file, err := os.CreateTemp("", "dfctl-zsh")
			data, err := os.ReadFile(dfpath.ConfigFile())
			if err != nil {
				return err
			}

			if _, err = file.Write(data); err != nil {
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

			patchedCfg, err := config.LoadFromPath(file.Name())
			if err != nil {
				return err
			}

			patchedToml, err := patchedCfg.Toml()
			if err != nil {
				return err
			}

			if err = os.WriteFile(dfpath.ConfigFile(), []byte(patchedToml), os.ModePerm); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
