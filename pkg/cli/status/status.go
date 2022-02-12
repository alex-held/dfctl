package status

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl/pkg/factory"
)

func NewStatusCommand(f factory.Factory) *cobra.Command {
	cmd := f.NewCommand("status",
		factory.WithHelp("shows dfctl status information", ""),
	)
	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		_, err = fmt.Fprintln(cmd.OutOrStdout(), "dfctl status")
		return err
	}
	return cmd
}
