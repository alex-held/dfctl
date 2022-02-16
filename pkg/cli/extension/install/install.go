package install

import (
	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl/pkg/extensions"
	"github.com/alex-held/dfctl/pkg/factory"
	"github.com/alex-held/dfctl/pkg/git"
)

func NewInstallCommand(f *factory.Factory) *cobra.Command {
	cmd := f.NewCommand("install")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		repoUri := args[0]
		repo := git.NewRepoFromURL(repoUri)
		em := extensions.NewManager(f)
		err := em.Install(repo)
		return err
	}
	return cmd
}
