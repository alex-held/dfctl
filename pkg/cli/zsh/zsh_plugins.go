package zsh

import (
	"fmt"

	"github.com/ahmetb/go-linq"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/alex-held/dfctl/pkg/zsh"
)

func newPluginsCommand() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "plugins",
	}
	cmd.AddCommand(newPluginsListCommand())
	cmd.AddCommand(newPluginsInstallCommand())
	return cmd
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
func newPluginsListCommand() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use: "list",
	}

	filter := cmd.PersistentFlags().StringP("filter", "f", "all", "--filter | -f [ all | installed | uninstalled ]")
	out := cmd.PersistentFlags().StringP("out", "o", "table", "--out | -o [ list | table ]")

	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		var installables []zsh.Installable
		switch *filter {
		case "all":
			linq.
				From(zsh.ListInstallables()).
				WhereT(zsh.KindFilterFn(zsh.PluginKind)).
				ToSlice(&installables)
		case "installed":
			linq.
				From(zsh.ListInstallables()).
				WhereT(zsh.KindFilterFn(zsh.PluginKind)).
				WhereT(zsh.InstalledFilterFn(true)).
				ToSlice(&installables)
		case "uninstalled":
			linq.
				From(zsh.ListInstallables()).
				WhereT(zsh.KindFilterFn(zsh.PluginKind)).
				WhereT(zsh.InstalledFilterFn(true)).
				ToSlice(&installables)
		}

		formatOutput(installables, *out, cmd)
		return nil
	}
	return cmd
}

func formatOutput(installables []zsh.Installable, outputFormat string, cmd *cobra.Command) {
	switch outputFormat {
	case "table":
		data := [][]string{}

		linq.From(installables).SelectT(func(i zsh.Installable) []string {
			return []string{i.Id(), fmt.Sprintf("%v", i.(*zsh.Plugin).Kind), fmt.Sprintf("%v", i.GetKind()), fmt.Sprintf("%v", i.IsInstalled())}
		}).ToSlice(&data)

		table := tablewriter.NewWriter(cmd.OutOrStderr())
		table.SetHeader([]string{"Name", "Kind", "Type", "Installed"})
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
