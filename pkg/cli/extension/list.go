package extension

import (
	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl/pkg/extensions"
	"github.com/alex-held/dfctl/pkg/factory"
)

func newListCommand(f *factory.Factory) *cobra.Command {
	cmd := f.NewCommand("list")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		em := extensions.NewManager(f)
		list := em.List(true)
		for _, extension := range list {
			name := extension.Name()
			_, _ = cmd.OutOrStdout().Write([]byte(name + "\n"))
		}
		return nil
	}

	return cmd
}
