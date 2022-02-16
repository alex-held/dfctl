package plugins

import (
	"github.com/ahmetb/go-linq"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl/pkg/factory"
	"github.com/alex-held/dfctl/pkg/zsh"
)

func newPluginsDisableCommand(*factory.Factory) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "disable [plugin1 plugin2 plugin3]",
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		log.Debug().Msgf("disabling %v", args)

		installables := GetPluginsByNames(args, func(q linq.Query) linq.Query {
			return q.WhereT(installablePredicate(func(i zsh.Installable) bool {
				return i.IsEnabled()
			}))
		})

		for _, i := range installables {
			log.Debug().Str("id", i.Id()).Str("kind", string(GetRepoKind(i))).Bool("enabled", i.IsEnabled()).Bool("installed", i.IsInstalled()).Msg("enabling...")
			if err := i.SetEnabled(false); err != nil {
				log.Error().Err(err).Msgf("failed to disable %s", i.Id())
				return err
			}
		}

		return nil
	}

	return cmd
}
