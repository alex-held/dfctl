package factory

import (
	"fmt"
	"io"
	"reflect"

	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/alex-held/dfctl-kit/pkg/iostreams"

	"github.com/alex-held/dfctl/pkg/globals"
)

var Default = BuildFactory()

type factoryBuilder struct {
	*FactoryConfig
}

type Factory struct {
	Streams *iostreams.IOStreams
	Fs      afero.Fs
}

func (fb *factoryBuilder) Build() *Factory {
	return &Factory{
		Streams: fb.Streams,
		Fs:      fb.FS,
	}
}

type CommandConfig struct {
	Flag        *FlagsConfig
	Help        *helpConfig
	Name        string
	Subcommands CommandFactoryGroups
	Group       string
	Annotations map[string]string
}

type CommandFactoryFns []CommandFactoryFn
type CommandFactoryGroups map[string]CommandFactoryFns

func WithLocalFlags(opts ...FlagsOption) CommandOption {
	return func(c *CommandConfig) {
		for _, opt := range opts {
			name := getFlagName(opt)
			c.Flag.Local[name] = opt
		}
	}
}

func WithPersistentFlags(opts ...FlagsOption) CommandOption {
	return func(c *CommandConfig) {
		for _, opt := range opts {
			name := getFlagName(opt)
			c.Flag.Persistent[name] = opt
		}
	}
}

func WithFlags(opts ...FlagsOption) CommandOption {
	return func(c *CommandConfig) {
		for _, opt := range opts {
			name := getFlagName(opt)
			c.Flag.Flags[name] = opt
		}
	}
}

func getFlagName(opt FlagsOption) (name string) {
	fs := &pflag.FlagSet{}
	opt(fs)
	fs.VisitAll(func(flag *pflag.Flag) {
		name = flag.Name
	})
	return name
}

type FlagsOption func(f *pflag.FlagSet) interface{}

type FlagValueStore map[string]interface{}

func (store FlagValueStore) Has(name string) (ok bool) {
	_, ok = store[name]
	return ok
}

func (store FlagValueStore) SetValue(name string, value interface{}) (old interface{}, updated bool) {
	if old, updated = store[name]; updated {
		log.Debug().
			Str("flag_name", name).
			Interface("old_value", old).
			Interface("new_value", value).
			Msg("updated flag value in FlagStore")
		return old, true
	}
	store[name] = value
	log.Debug().
		Str("flag_name", name).
		Interface("value", value).
		Msg("set flag value in FlagStore")
	return nil, false
}

type FlagsConfig struct {
	Values     FlagValueStore
	Persistent map[string]func(*pflag.FlagSet) (value interface{})
	Flags      map[string]func(*pflag.FlagSet) (value interface{})
	Local      map[string]func(*pflag.FlagSet) (value interface{})
}

var ErrFlagNotFound = fmt.Errorf("flag value not found")
var ErrFlagInvalidType = fmt.Errorf("flag value has invalid type")

func (c *FlagsConfig) getFlag(name string, val interface{}) (err error) {
	if v, ok := c.Values[name]; ok {
		vElem := reflect.ValueOf(v).Elem()
		valElem := reflect.ValueOf(val).Elem()
		if vElem.Type().AssignableTo(valElem.Type()) {
			valElem.Set(vElem)
			return nil
		}

		return fmt.Errorf("flag value has type %s but should have %s; %w", vElem.Type().Name(), valElem.Type().Name(), ErrFlagInvalidType)
	}
	return ErrFlagNotFound
}

func WithAnnotationKeys(keys ...string) CommandOption {
	return func(c *CommandConfig) {
		if len(c.Annotations) == 0 {
			c.Annotations = map[string]string{}
		}
		for _, key := range keys {
			c.Annotations[key] = key
		}
	}
}

func WithHelp(short, long string) CommandOption {
	return func(c *CommandConfig) {
		c.Help = &helpConfig{
			Short: short,
			Long:  long,
		}
	}
}

type CommandOption func(c *CommandConfig)
type CommandFactoryFn func(f *Factory) *cobra.Command

func WithSubcommands(fns ...CommandFactoryFn) CommandOption {
	return WithGroupedSubcommands("core commands", fns...)
}

func WithGroup(group string) CommandOption {
	return func(c *CommandConfig) {
		c.Group = group
	}
}
func WithGroupedSubcommands(group string, fns ...CommandFactoryFn) CommandOption {
	return func(c *CommandConfig) {
		c.Subcommands[group] = fns
	}
}

func (c *CommandConfig) Build(f *Factory) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   c.Name,
		Short: c.Help.Short,
		Long:  c.Help.Long,
	}

	// Flag
	for name, opt := range c.Flag.Persistent {
		value := opt(cmd.PersistentFlags())
		c.Flag.Values.SetValue(name, value)
	}
	for name, opt := range c.Flag.Local {
		value := opt(cmd.LocalFlags())
		c.Flag.Values.SetValue(name, value)
	}
	for name, opt := range c.Flag.Flags {
		value := opt(cmd.Flags())
		c.Flag.Values.SetValue(name, value)
	}

	if len(c.Group) > 0 {
		if cmd.Annotations == nil {
			cmd.Annotations = map[string]string{}
		}
		cmd.Annotations[globals.COMAND_GROUP_ANNOTATION_KEY] = c.Group
	}

	for group, factoryFns := range c.Subcommands {
		for _, factoryFn := range factoryFns {
			subcommand := factoryFn(f)
			if subcommand.Annotations == nil {
				subcommand.Annotations = map[string]string{}
			}
			subcommand.Annotations[globals.COMAND_GROUP_ANNOTATION_KEY] = group
			cmd.AddCommand(subcommand)
		}
	}

	return cmd
}

type helpConfig struct {
	Short string
	Long  string
}

func (f *Factory) NewCommand(use string, opts ...CommandOption) (cmd *cobra.Command) {
	cfg := &CommandConfig{
		Name: use,
		Flag: &FlagsConfig{
			Persistent: map[string]func(*pflag.FlagSet) interface{}{},
			Local:      map[string]func(*pflag.FlagSet) interface{}{},
			Flags:      map[string]func(*pflag.FlagSet) (value interface{}){},
			Values:     FlagValueStore{},
		},
		Help:        &helpConfig{},
		Subcommands: CommandFactoryGroups{},
	}

	for _, opt := range opts {
		opt(cfg)
	}

	cmd = cfg.Build(f)
	return cmd
}

type FactoryConfig struct {
	FS      afero.Fs
	Streams *iostreams.IOStreams
}

type FactoryOption func(c *FactoryConfig)

func WithStreamWriter(stdout, stderr io.WriteCloser, stdin io.ReadCloser) FactoryOption {
	stream := iostreams.Default()
	stream.In = stdin
	stream.Out = stdout
	stream.Err = stderr

	return WithStreams(stream)
}

func WithStreams(streams *iostreams.IOStreams) FactoryOption {
	return func(c *FactoryConfig) {
		c.Streams = streams
	}
}

func WithFS(fs afero.Fs) FactoryOption {
	return func(c *FactoryConfig) {
		c.FS = fs
	}
}

var defaultFactoryOptions = []FactoryOption{
	WithFS(afero.NewOsFs()),
	WithStreams(iostreams.Default()),
}

func BuildFactory(opts ...FactoryOption) *Factory {
	cfg := &FactoryConfig{}
	for _, opt := range defaultFactoryOptions {
		opt(cfg)
	}
	for _, opt := range opts {
		opt(cfg)
	}
	fb := &factoryBuilder{cfg}
	return fb.Build()
}
