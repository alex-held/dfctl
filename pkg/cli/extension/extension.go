package extension

import (
	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl/pkg/cli/extension/install"
	"github.com/alex-held/dfctl/pkg/cli/extension/run"
	"github.com/alex-held/dfctl/pkg/factory"
)

func NewExtensionCommand(f *factory.Factory) *cobra.Command {
	cmd := f.NewCommand("extension", factory.WithSubcommands(
		install.NewInstallCommand,
		run.NewRunCommand,
		newListCommand,
	))

	return cmd
}
