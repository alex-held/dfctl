package zsh

import (
	"fmt"

	"github.com/ahmetb/go-linq"
	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl/pkg/zsh"
)

func newPluginsCommand() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "plugins",
	}
	cmd.AddCommand(newPluginsListCommand())
	cmd.AddCommand(newPluginsInstallCommand())
	cmd.AddCommand(newPluginsEnableCommand())
	cmd.AddCommand(newPluginsDisableCommand())
	return cmd
}

func newPluginsDisableCommand() (cmd *cobra.Command) {
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

func newPluginsEnableCommand() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "enable [plugin1 plugin2 plugin3]",
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		installables := GetInstallablesByNames(args, func(q linq.Query) linq.Query {
			return q.WhereT(installablePredicate(func(i zsh.Installable) bool {
				return !i.IsEnabled()
			}))
		})

		for _, i := range installables {
			log.Debug().Str("id", i.Id()).Str("kind", string(GetRepoKind(i))).Bool("enabled", i.IsEnabled()).Bool("installed", i.IsInstalled()).Msg("enabling...")

			if !i.IsInstalled() {
				log.Debug().Str("id", i.Id()).Str("kind", string(GetRepoKind(i))).Msg("installing...")
				if res := i.Install(); res.Err != nil {
					log.Error().Err(res.Err).Msgf("failed to install %s", i.Id())
				}
			}

			if err := i.SetEnabled(true); err != nil {
				log.Error().Err(err).Msgf("failed to enable %s", i.Id())
				return err
			}
		}

		return nil
	}

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

func newPluginsInstallCommand() (cmd *cobra.Command) {
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

type installablePredicate func(i zsh.Installable) bool

func newPluginsListCommand() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "list",
	}

	// filter := cmd.PersistentFlags().StringP("filter", "f", "all", "--filter | -f [ all | installed | uninstalled ]")
	filters := cmd.PersistentFlags().StringSliceP("filters", "f", []string{"all"}, "--filter [filter1,filter2,..]  (default: all) |  filters: all | enabled | disabled | kind:gh | kind:git | kind:omz |installed | uninstalled ]")
	//	enabled := cmd.PersistentFlags().String("enabled", "all", "--enabled [ true | false ] (default: true)")
	//	installed := cmd.PersistentFlags().String("installed", "all", "--installed [ true | false ] (default: true)")
	out := cmd.PersistentFlags().StringP("out", "o", "table", "--out | -o [ list | table ]")

	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		var installables []zsh.Installable
		var predicates []installablePredicate

	filters:
		for _, filter := range *filters {
			switch filter {
			case "all":
				predicates = []installablePredicate{}
				break filters
			case "enabled":
				predicates = append(predicates, func(i zsh.Installable) bool {
					return i.IsEnabled()
				})
			case "disabled":
				predicates = append(predicates, func(i zsh.Installable) bool {
					return !i.IsEnabled()
				})
			case "installed":
				predicates = append(predicates, func(i zsh.Installable) bool {
					return i.IsInstalled()
				})
			case "uninstalled":
				predicates = append(predicates, func(i zsh.Installable) bool {
					return !i.IsInstalled()
				})
			case "kind:gh":
				predicates = append(predicates, func(i zsh.Installable) bool {
					if it, ok := i.(*zsh.Plugin); ok {
						return it.Kind == zsh.PLUGIN_GITHUB
					}
					return false
				})
			case "kind:git":
				predicates = append(predicates, func(i zsh.Installable) bool {
					if it, ok := i.(*zsh.Plugin); ok {
						return it.Kind == zsh.PLUGIN_GIT
					}
					return false
				})
			case "kind:omz":
				predicates = append(predicates, func(i zsh.Installable) bool {
					_, ok := i.(*zsh.OMZPlugin)
					return ok
				})
			default:
				log.Error().Msgf("unsupported filter %s", filter)
			}
		}

		linq.
			From(zsh.ListInstallables(zsh.KindFilterFn(zsh.PluginInstallableKind))).
			WhereT(installablePredicate(func(i zsh.Installable) (isMatch bool) {
				for _, predicate := range predicates {
					if !predicate(i) {
						return false
					}
				}
				return true
			})).
			SortT(func(i, j zsh.Installable) bool {
				iKind := GetRepoKind(i)
				jKind := GetRepoKind(j)
				return iKind.Order() < jKind.Order()
			}).
			ToSlice(&installables)

		formatOutput(installables, *out, cmd)
		return nil
	}
	return cmd
}

func GetRepoKind(i zsh.Installable) zsh.RepoKind {
	switch it := i.(type) {
	case *zsh.Plugin:
		return it.Kind
	case *zsh.OMZPlugin:
		return zsh.PLUGIN_OMZ
	default:
		log.Fatal().Msgf("unable to get repo kind for %T %v", it, it)
		return "panic"
	}
}

func formatOutput(installables []zsh.Installable, outputFormat string, cmd *cobra.Command) {
	switch outputFormat {
	case "table":
		data := [][]string{}

		linq.From(installables).SelectT(func(i zsh.Installable) []string {
			var kind string
			switch it := i.(type) {
			case *zsh.Plugin:
				kind = string(it.Kind)
			case *zsh.OMZPlugin:
				kind = "omz"
			}
			return []string{i.Id(), fmt.Sprintf("%s", kind), fmt.Sprintf("%v", i.GetKind()), fmt.Sprintf("%v", i.IsEnabled())}
		}).ToSlice(&data)

		table := tablewriter.NewWriter(cmd.OutOrStderr())
		table.SetHeader([]string{"Name", "Kind", "Type", "Enabled"})
		table.SetHeaderColor(
			tablewriter.Colors{tablewriter.FgHiWhiteColor, tablewriter.Bold},
			tablewriter.Colors{tablewriter.FgHiMagentaColor, tablewriter.Bold},
			tablewriter.Colors{tablewriter.FgRedColor, tablewriter.Bold},
			tablewriter.Colors{tablewriter.FgCyanColor, tablewriter.Bold})

		table.SetAutoWrapText(false)
		table.SetAutoFormatHeaders(true)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)

		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("")
		table.SetHeaderLine(false)
		table.SetBorder(false)
		table.SetTablePadding("\t") // pad with tabs
		table.SetNoWhiteSpace(true)

		table.AppendBulk(data)
		table.Render()
		return

	case "list":
		fallthrough
	default:
		for _, installable := range installables {
			_, _ = cmd.OutOrStderr().Write([]byte(installable.Id() + "\n"))
		}
	}
}
