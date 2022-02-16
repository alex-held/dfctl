package cli

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/cli/cli/pkg/iostreams"
	"github.com/kr/text"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/alex-held/dfctl/pkg/cli/config"
	"github.com/alex-held/dfctl/pkg/cli/status"
	"github.com/alex-held/dfctl/pkg/cli/version"
	"github.com/alex-held/dfctl/pkg/cli/zsh/zsh"
	color2 "github.com/alex-held/dfctl/pkg/color"
	"github.com/alex-held/dfctl/pkg/factory"
	"github.com/alex-held/dfctl/pkg/globals"
)

func ConfigureLogger(levelString string) {
	level, err := zerolog.ParseLevel(levelString)
	if err != nil {
		level = zerolog.InfoLevel
	}
	w := zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.PartsExclude = []string{zerolog.CallerFieldName, zerolog.TimestampFieldName}
		w.PartsOrder = []string{zerolog.LevelFieldName, zerolog.MessageFieldName}
	})

	log.Logger = zerolog.New(w)
	zerolog.SetGlobalLevel(level)
}

func NewRootCommand(f *factory.Factory) (cmd *cobra.Command) {
	cmd = f.NewCommand("dfctl [flags] [command]",
		factory.WithHelp("dotfiles and development environment manager", ""),
		factory.WithGroupedSubcommands("module commands", zsh.NewZshCommand),
		factory.WithGroupedSubcommands("environment commands", config.NewConfigCommand),
		factory.WithGroupedSubcommands("status commands", status.NewStatusCommand, version.NewVersionCommand),
	)

	cmd.PersistentFlags().String("level", "info", "set the log level [ trace | debug | info | warn | error | fatal ]")
	cmd.PersistentPreRun = func(c *cobra.Command, _ []string) {
		initialize(c)
	}

	cmd.Aliases = []string{"dfctl [flags]", "dfctl [command]"}
	cmd.Example = `
ZSH:
	dfctl --level=debug zsh source
	dfctl zsh install --level info
	
ZSH Plugins:
	dfctl zsh plugins list --filters kind:gh,enabled
	dfctl zsh plugins enable golang
	dfctl zsh plugins disable brew
	`

	cmd.SetHelpFunc(rootHelpFunc)
	cmd.SetUsageFunc(rootUsageFunc)
	return cmd
}

func indent(spaces int, v string) string {
	pad := strings.Repeat(" ", spaces)
	return pad + strings.Replace(v, "\n", "\n"+pad, -1)
}

func colorize(colorStr string, v string) string {
	c := color2.ParseColorOrDefault(colorStr, color2.Green)
	colorized := color2.Colorize(v, c)
	return colorized
}

var funcMap = template.FuncMap{
	"indent":   indent,
	"colorize": colorize,
}

func getUsage(cmd *cobra.Command) string {
	usageBuilder := &strings.Builder{}

	// Use is the one-line usage message.
	// Recommended syntax is as follow:
	//   [ ] identifies an optional argument. Arguments that are not enclosed in brackets are required.
	//   ... indicates that you can specify multiple values for the previous argument.
	//   |   indicates mutually exclusive information. You can use the argument to the left of the separator or the
	//       argument to the right of the separator. You cannot use both arguments in a single use of the command.
	//   { } delimits a set of mutually exclusive arguments when one of the arguments is required. If the arguments are
	//       optional, they are enclosed in brackets ([ ]).
	// Example: add [-F file | -D dir]... [-f format] profile

	tmpl := `Usage:
  {{ .Use -}}
  {{- if .Aliases }}
  {{- range $alias := .Aliases }}
  {{ $alias }}
  {{- end -}}
  {{ end }}

{{- if .Flags }}

Flags:
{{ range .Flags -}}{{- tableRow  (truncate (colorize "yellow" .Name) 25)  (truncate (colorize "cyan" .DefValue) 15) (truncate (colorize "green" .Usage) 80) -}}{{- end -}}{{- endTable }}
{{ end -}}`

	type Flag struct {
		Name     string
		DefValue string
		Usage    string
	}

	data := struct {
		Use     string
		Aliases []string
		Flags   []Flag
	}{
		strings.TrimSuffix(cmd.Use, "\n"),
		cmd.Aliases,
		[]Flag{},
	}

	table := NewTable(usageBuilder)
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		name := fmt.Sprintf("--%s", f.Name)
		if f.Shorthand != "" {
			name = fmt.Sprintf("-%s, --%s", f.Shorthand, f.Name)
		}
		data.Flags = append(data.Flags, Flag{
			Name:     name,
			DefValue: f.DefValue,
			Usage:    f.Usage,
		})
	})

	tpl := template.Must(template.
		New("usage").
		Funcs(funcMap).
		Funcs(table.FuncMap()).
		Parse(tmpl),
	)

	if err := tpl.ExecuteTemplate(usageBuilder, "usage", data); err != nil {
		return cmd.Use
	}

	return usageBuilder.String()
}

func isRootCmd(command *cobra.Command) bool {
	return command != nil && !command.HasParent()
}

// HasFailed signals that the main process should exit with non-zero status
func HasFailed() bool {
	return hasFailed
}

// Display helpful error message in case subcommand name was mistyped.
// This matches Cobra's behavior for root command, which Cobra
// confusingly doesn't apply to nested commands.
func nestedSuggestFunc(command *cobra.Command, arg string) {
	command.Printf("unknown command %q for %q\n", arg, command.CommandPath())

	var candidates []string
	if arg == "help" {
		candidates = []string{"--help"}
	} else {
		if command.SuggestionsMinimumDistance <= 0 {
			command.SuggestionsMinimumDistance = 2
		}
		candidates = command.SuggestionsFor(arg)
	}

	if len(candidates) > 0 {
		command.Print("\nDid you mean this?\n")
		for _, c := range candidates {
			command.Printf("\t%s\n", c)
		}
	}

	command.Print("\n")
	_ = rootUsageFunc(command)
}

func rootUsageFunc(command *cobra.Command) error {
	initialize(command)

	command.Printf("Usage:  %s", command.UseLine())

	subcommands := command.Commands()
	if len(subcommands) > 0 {
		command.Print("\n\nAvailable commands:\n")
		for _, c := range subcommands {
			if c.Hidden {
				continue
			}
			command.Printf("  %s\n", c.Name())
		}
		return nil
	}

	flagUsages := command.LocalFlags().FlagUsages()
	if flagUsages != "" {
		command.Println("\n\nFlags:")
		command.Print(text.Indent(dedent(flagUsages), "  "))
	}
	return nil
}

func dedent(s string) string {
	lines := strings.Split(s, "\n")
	minIndent := -1

	for _, l := range lines {
		if len(l) == 0 {
			continue
		}

		indent := len(l) - len(strings.TrimLeft(l, " "))
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return s
	}

	var buf bytes.Buffer
	for _, l := range lines {
		fmt.Fprintln(&buf, strings.TrimPrefix(l, strings.Repeat(" ", minIndent)))
	}
	return strings.TrimSuffix(buf.String(), "\n")
}

var hasFailed = false

func initialize(command *cobra.Command) {
	level, err := command.Flags().GetString("level")
	if err != nil {
		ConfigureLogger("info")
		return
	}
	ConfigureLogger(level)
}

func rootHelpFunc(command *cobra.Command, args []string) {
	initialize(command)

	streams := iostreams.System()
	cs := streams.ColorScheme()

	if isRootCmd(command.Parent()) && len(args) >= 2 && args[1] != "--help" && args[1] != "-h" {
		nestedSuggestFunc(command, args[1])
		hasFailed = true
		return
	}

	var coreCommands []string
	var actionsCommands []string
	var additionalCommands []string

	for _, c := range command.Commands() {
		if c.Short == "" {
			continue
		}
		if c.Hidden {
			continue
		}

		s := rpad(c.Name()+":", c.NamePadding()) + c.Short
		groups := map[string][]string{}
		if groupName, ok := c.Annotations[globals.COMAND_GROUP_ANNOTATION_KEY]; ok {
			groups[groupName] = append(groups[groupName], s)
		}

		if _, ok := c.Annotations["IsCore"]; ok {
			coreCommands = append(coreCommands, s)
		} else if _, ok := c.Annotations["IsActions"]; ok {
			actionsCommands = append(actionsCommands, s)
		} else {
			//	additionalCommands = append(additionalCommands, s)
		}
	}

	// If there are no core commands, assume everything is a core command
	if len(coreCommands) == 0 {
		coreCommands = additionalCommands
		additionalCommands = []string{}
	}

	type helpEntry struct {
		Title string
		Body  string
	}

	longText := command.Long
	if longText == "" {
		longText = command.Short
	}
	if longText != "" && command.LocalFlags().Lookup("jq") != nil {
		longText = strings.TrimRight(longText, "\n") +
			"\n\nFor more information about output formatting flags, see `gh help formatting`."
	}

	var helpEntries []helpEntry
	if longText != "" {
		helpEntries = append(helpEntries, helpEntry{"", longText})
	}
	helpEntries = append(helpEntries, helpEntry{"USAGE", command.UseLine()})

	// GROUPS
	groups := getCommandGroups(command)
	if len(coreCommands) > 0 {
		helpEntries = append(helpEntries, helpEntry{"CORE COMMANDS", strings.Join(coreCommands, "\n")})
	}

	for group, cmd := range groups {
		log.Debug().Msgf("rendering help for command group %s", group)
		var groupCommands []string
		for _, groupsCmd := range cmd {
			s := rpad(groupsCmd.c.Name()+":", groupsCmd.c.NamePadding()) + groupsCmd.c.Short
			log.Debug().Str("command_group", group).Msgf("rendering help for command %s", s)
			groupCommands = append(groupCommands, s)
		}
		helpEntries = append(helpEntries, helpEntry{strings.ToTitle(group), strings.Join(groupCommands, "\n")})
	}

	if len(actionsCommands) > 0 {
		helpEntries = append(helpEntries, helpEntry{"ACTIONS COMMANDS", strings.Join(actionsCommands, "\n")})
	}
	if len(additionalCommands) > 0 {
		helpEntries = append(helpEntries, helpEntry{"ADDITIONAL COMMANDS", strings.Join(additionalCommands, "\n")})
	}

	if isRootCmd(command) {
		// TODO: extensions
		// if exts := f.ExtensionManager.List(false); len(exts) > 0 {
		// 	var names []string
		// 	for _, ext := range exts {
		// 		names = append(names, ext.Name())
		// 	}
		// 	helpEntries = append(helpEntries, helpEntry{"EXTENSION COMMANDS", strings.Join(names, "\n")})
		// }
	}

	flagUsages := command.LocalFlags().FlagUsages()
	if flagUsages != "" {
		helpEntries = append(helpEntries, helpEntry{"FLAGS", dedent(flagUsages)})
	}
	inheritedFlagUsages := command.InheritedFlags().FlagUsages()
	if inheritedFlagUsages != "" {
		helpEntries = append(helpEntries, helpEntry{"INHERITED FLAGS", dedent(inheritedFlagUsages)})
	}
	if _, ok := command.Annotations["help:arguments"]; ok {
		helpEntries = append(helpEntries, helpEntry{"ARGUMENTS", command.Annotations["help:arguments"]})
	}
	if command.Example != "" {
		helpEntries = append(helpEntries, helpEntry{"EXAMPLES", command.Example})
	}
	if _, ok := command.Annotations["help:environment"]; ok {
		helpEntries = append(helpEntries, helpEntry{"ENVIRONMENT VARIABLES", command.Annotations["help:environment"]})
	}
	helpEntries = append(helpEntries, helpEntry{"LEARN MORE", `
Use 'dfctl <command> <subcommand> --help' for more information about a command.
Read the manual at https://dfctl.alexheld.io/manual`})

	if _, ok := command.Annotations["help:feedback"]; ok {
		helpEntries = append(helpEntries, helpEntry{"FEEDBACK", command.Annotations["help:feedback"]})
	}

	out := command.OutOrStdout()
	for _, e := range helpEntries {
		if e.Title != "" {
			// If there is a title, add indentation to each line in the body
			fmt.Fprintln(out, cs.Bold(e.Title))
			fmt.Fprintln(out, text.Indent(strings.Trim(e.Body, "\r\n"), "  "))
		} else {
			// If there is no title print the body as is
			fmt.Fprintln(out, e.Body)
		}
		fmt.Fprintln(out)
	}
}

func helpFn() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		// Short
		fmt.Println(cmd.Short)
		fmt.Println()

		// Usages
		usage := getUsage(cmd)
		fmt.Println(usage)

		// Examples
		example := cmd.Example
		fmt.Println(example)

		// Groups
		usage = helpMessageByGroups(cmd)
		fmt.Println(usage)
	}
}

// rpad adds padding to the right of a string.
func rpad(s string, padding int) string {
	template := fmt.Sprintf("%%-%ds ", padding)
	return fmt.Sprintf(template, s)
}

const cmdGroupOthers = "others"
const cmdGroupDelimiter = "-"

func helpMessageByGroups(cmd *cobra.Command) string {
	groups := map[string][]string{}

	for _, c := range cmd.Commands() {
		var groupName string
		v, ok := c.Annotations[globals.COMAND_GROUP_ANNOTATION_KEY]
		if !ok {
			groupName = cmdGroupOthers
		} else {
			groupName = v
		}

		groupCmds, ok := groups[groupName]
		groupCmds = append(groupCmds, fmt.Sprintf("\t%-24s%s", c.Name(), c.Short))
		sort.Strings(groupCmds)

		groups[groupName] = groupCmds
	}
	//
	// if len(groups[cmdGroupOthers]) != 0 {
	// 	groups[cmdGroupOthers] = append(groups[cmdGroupOthers], groups[cmdGroupCobra]...)
	// }
	// delete(groups, cmdGroupCobra)

	// sort by group name
	var groupNames []string
	for k, _ := range groups {
		groupNames = append(groupNames, k)
	}
	sort.Strings(groupNames)

	// Group by group and sort commands within the group
	buf := bytes.Buffer{}
	for _, groupName := range groupNames {
		commands, _ := groups[groupName]

		groupNameParts := strings.Split(groupName, cmdGroupDelimiter)
		if len(groupNameParts) > 1 {
			groupNameParts = append(groupNameParts[:0], groupNameParts[1:]...)
		}
		group := groupNameParts[0]
		buf.WriteString(fmt.Sprintf("%s\n", strings.ToTitle(group)))

		for _, cmd := range commands {
			buf.WriteString(fmt.Sprintf("%s\n", cmd))
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

type cmdGroupsCmd struct {
	name string
	c    *cobra.Command
}
type cmdGroupsCmds []cmdGroupsCmd
type cmdGroups map[string]cmdGroupsCmds

func (c cmdGroupsCmds) Len() int { return len(c) }

func (c cmdGroupsCmds) Less(i, j int) bool {
	return c[i].name < c[j].name
}

func (c cmdGroupsCmds) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func getCommandGroups(cmd *cobra.Command) (result cmdGroups) {
	groups := cmdGroups{}

	for _, c := range cmd.Commands() {
		var groupName string
		v, ok := c.Annotations[globals.COMAND_GROUP_ANNOTATION_KEY]
		if !ok {
			groupName = cmdGroupOthers
		} else {
			groupName = v
		}

		groupCmds, ok := groups[groupName]
		groupCmds = append(groupCmds, cmdGroupsCmd{
			fmt.Sprintf("\t%-24s%s", c.Name(), c.Short),
			c,
		})

		sort.Sort(groupCmds)

		groups[groupName] = groupCmds
	}

	// sort by group name
	var groupNames []string
	for k, _ := range groups {
		groupNames = append(groupNames, k)
	}
	sort.Strings(groupNames)

	// Group by group and sort commands within the group
	//	buf := bytes.Buffer{}
	result = cmdGroups{}

	for _, groupName := range groupNames {
		commands, _ := groups[groupName]

		groupNameParts := strings.Split(groupName, cmdGroupDelimiter)
		if len(groupNameParts) > 1 {
			groupNameParts = append(groupNameParts[:0], groupNameParts[1:]...)
		}
		group := groupNameParts[0]
		for _, cmd := range commands {
			result[group] = append(result[group], cmd)
		}
	}
	return result
}
