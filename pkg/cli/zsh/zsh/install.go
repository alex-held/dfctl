package zsh

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl/pkg/factory"
	"github.com/alex-held/dfctl/pkg/zsh"
)

func newInstallCommand(f factory.Factory) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "install",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runInstallCommand(f)
	}
	return cmd
}

func runInstallCommand(f factory.Factory) (err error) {
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
		return nil
	}
	return nil
}
