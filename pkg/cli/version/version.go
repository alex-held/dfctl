package version

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl/pkg/factory"
	"github.com/alex-held/dfctl/pkg/globals"
)

func NewVersionCommand(f factory.Factory) *cobra.Command {
	cmd := f.NewCommand("version",
		factory.WithHelp("displays dfctl version information", ""),
	)
	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		sb := &strings.Builder{}
		sb.WriteString(fmt.Sprintf("go-version: %s\n", runtime.Version()))
		sb.WriteString(fmt.Sprintf("dfctl-version: %s\n", globals.Version))
		_, err = fmt.Fprint(cmd.OutOrStdout(), sb.String())
		return err
	}
	return cmd
}
