package zsh

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl/pkg/zsh"
)

func newEnableCommand() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "enable [plugin]",
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		installables := zsh.ListInstallables()
		plugin := args[0]

		for _, installable := range installables {
			if installable.Id() == plugin {
				log.Debug().Msgf("plugin %s found; installed: %v", installable.IsInstalled())
				if installable.IsEnabled() {
					cmd.Printf("plugin %s already enabled", plugin)
					return nil
				}
				err := installable.Enable(true)
				return err
			}
		}

		return fmt.Errorf("plugin %s not found ")
	}

	return cmd
}
