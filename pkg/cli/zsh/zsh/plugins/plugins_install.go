package plugins

import (
	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl/pkg/factory"
	"github.com/alex-held/dfctl/pkg/zsh"
)

func newPluginsInstallCommand(factory.Factory) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "install [repo]",
		Short: "installs a plugin locally",
		Long: `installs a plugin locally
		
			[repo] must have following format:
			[type]:[urn]
		
			[type] must be one of following:
			
				omz     oh-my-zsh bundled plugin
				gh      github repository
				git     git repository
		`,
	}

	nameFlag := cmd.Flags().StringP("name", "n", "", "--name | -n [name of the plugin]")
	idFlag := cmd.Flags().StringP("id", "i", "", "--id | -i [id of the plugin]")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		repo := args[0]
		plugin := zsh.NewPlugin(repo, idFlag, nameFlag)
		zsh.Install(plugin)
		return nil
	}

	return cmd
}
