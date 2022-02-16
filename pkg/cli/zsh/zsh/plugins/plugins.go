package plugins

import (
	"github.com/ahmetb/go-linq"
	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl/pkg/factory"
	"github.com/alex-held/dfctl/pkg/zsh"
)

func NewPluginsCommand(f *factory.Factory) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "plugins",
	}
	cmd.AddCommand(newPluginsListCommand(f))
	cmd.AddCommand(newPluginsInstallCommand(f))
	cmd.AddCommand(newPluginsEnableCommand(f))
	cmd.AddCommand(newPluginsDisableCommand(f))
	return cmd
}

type QueryFn func(query linq.Query) linq.Query

func GetPluginsByNames(names []string, fns ...QueryFn) (installables []zsh.Installable) {
	return GetInstallablesByNames(names,
		append([]QueryFn{func(q linq.Query) linq.Query {
			return q.WhereT(func(i zsh.Installable) bool {
				return i.GetKind() == zsh.PluginInstallableKind
			})
		}}, fns...)...)
}

func GetInstallablesByNames(names []string, fns ...QueryFn) (installables []zsh.Installable) {
	query := linq.
		From(names).
		JoinT(linq.
			From(zsh.ListInstallables()),
			func(id string) string { return id },
			func(i zsh.Installable) string { return i.Id() },
			func(id string, i zsh.Installable) zsh.Installable { return i },
		)

	for _, queryFn := range fns {
		query = queryFn(query)
	}
	query.
		DistinctByT(func(i zsh.Installable) string {
			return i.Id()
		}).
		ToSlice(&installables)
	return installables
}

type installablePredicate func(i zsh.Installable) bool
