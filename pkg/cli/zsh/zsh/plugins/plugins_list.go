package plugins

import (
	"fmt"

	"github.com/ahmetb/go-linq"
	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl/pkg/factory"
	"github.com/alex-held/dfctl/pkg/out"
	"github.com/alex-held/dfctl/pkg/zsh"
)

func newPluginsListCommand(f *factory.Factory) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "list",
	}

	filters := cmd.PersistentFlags().StringSliceP("filters", "f", []string{"all"}, "--filter [filter1,filter2,..]  (default: all) |  filters: all | enabled | disabled | kind:gh | kind:git | kind:omz |installed | uninstalled ]")
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

		installables = zsh.ListInstallables(zsh.KindFilterFn(zsh.PluginInstallableKind))
		linq.
			From(installables).
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

type pluginFormatter struct{}
type pluginListFormatter struct{}

func (p pluginListFormatter) Format(v interface{}) (values []string, options []out.FormatOption) {
	plugin := v.(Plugin)
	values = append(values, plugin.ID)
	return values, options
}

type Plugin struct {
	ID      string
	Enabled bool
	Kind    string
}

func (pluginFormatter) Format(v interface{}) (values []string, options []out.FormatOption) {
	plugin := v.(Plugin)

	// id
	switch plugin.Enabled {
	case true:
		options = append(options, out.ColorFormat(tablewriter.Colors{tablewriter.FgGreenColor, tablewriter.Bold}))
	case false:
		options = append(options, out.ColorFormat(tablewriter.Colors{tablewriter.FgHiWhiteColor, tablewriter.Normal}))
	}
	values = append(values, plugin.ID)

	// kind
	switch plugin.Kind {
	case "github":
		options = append(options, out.ColorFormat(tablewriter.Colors{tablewriter.FgYellowColor}))
	case "git":
		options = append(options, out.ColorFormat(tablewriter.Colors{tablewriter.FgHiYellowColor}))
	case "omz":
		options = append(options, out.ColorFormat(tablewriter.Colors{tablewriter.FgMagentaColor}))
	default:
		options = append(options, out.ColorFormat(tablewriter.Colors{tablewriter.FgHiRedColor}))
	}
	values = append(values, plugin.Kind)

	// enabled
	switch plugin.Enabled {
	case true:
		options = append(options, out.ColorFormat(tablewriter.Colors{tablewriter.FgGreenColor}))
	case false:
		options = append(options, out.ColorFormat(tablewriter.Colors{tablewriter.FgRedColor}))
	}
	values = append(values, fmt.Sprintf("%v", plugin.Enabled))

	return values, options
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
	var data []interface{}
	linq.
		From(installables).
		SelectT(func(i zsh.Installable) (ds interface{}) {
			var kind string
			switch it := i.(type) {
			case *zsh.Plugin:
				kind = string(it.Kind)
			case *zsh.OMZPlugin:
				kind = "omz"
			}
			return Plugin{
				ID:      i.Id(),
				Kind:    kind,
				Enabled: i.IsEnabled(),
			}
		}).
		ToSlice(&data)

	switch outputFormat {
	case "table":
		sink := out.NewTableSink(cmd.OutOrStdout(), pluginFormatter{}, func(t *tablewriter.Table) {
			t.SetHeader([]string{"Name", "Kind", "Enabled"})
		})
		if err := sink.WriteAndFlush(data); err != nil {
			log.Error().Err(err).Msgf("unable to format data %v", data)
			return
		}
	case "list":
		sink := out.NewListSink(cmd.OutOrStdout(), pluginListFormatter{})
		if err := sink.WriteAndFlush(data); err != nil {
			log.Error().Err(err).Msgf("unable to format data %v", data)
			return
		}
	default:
		log.Error().Err(ErrInvalidOutputFormat).Msgf("%s is not a supported output format", outputFormat)
	}
}

var ErrInvalidOutputFormat = fmt.Errorf("invalid output format")
