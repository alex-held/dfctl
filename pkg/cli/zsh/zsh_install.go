package zsh

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl-zsh/pkg/zsh"
)

func newInstallCommand() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "install",
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		installables := zsh.ListInstallables()
		for _, installable := range installables {
			if !installable.IsInstalled() {
				log.Debug().Msgf("installing %v %s into %s", installable.GetKind(), installable.Id(), installable.Path())
				result := installable.Install()
				if result.Err != nil {
					log.Error().Err(result.Err).Msgf("failed to install %v %s into %s", installable.GetKind(), installable.Id(), installable.Path())
					return result.Err
				}
				log.Debug().Bool("was_installed", result.Installed).Msgf("%v %s", installable.GetKind(), installable.Id())
			}
		}
		return nil
	}

	return cmd
}
