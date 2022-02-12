package factory

import (
	"fmt"
	"io"
	"os"
	"reflect"

	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func GetFS() afero.Fs {
	return fac.GetFS()
}
func GetSteams() *Streams {
	return fac.GetStreams()
}

var DefaultFactory = BuildFactory()
var fac = DefaultFactory

type factory struct {
	*FactoryConfig
}

func (f *factory) GetFS() afero.Fs {
	return f.FS
}

func (f *factory) GetStreams() *Streams {
	return f.Streams
}

type CommandConfig struct {
	Flag    *FlagsConfig
	Help    *helpConfig
	Name    string
	FS      afero.Fs
	Streams *Streams
	Factory *factory
}

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

func WithHelp(short, long string) CommandOption {
	return func(c *CommandConfig) {
		if long == "" {
			long = short
		}
		c.Help = &helpConfig{
			Short: short,
			Long:  long,
		}
	}
}

type CommandOption func(c *CommandConfig)

func (c *CommandConfig) CobraCommand() *cobra.Command {
	cmd := &cobra.Command{
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

	return cmd
}

func (c *CommandConfig) Build() (cmd *cobra.Command) {
	cmd = c.CobraCommand()
	return cmd
}

type helpConfig struct {
	Short string
	Long  string
}

func (f *factory) NewCommand(use string, opts ...CommandOption) (cmd *cobra.Command) {
	cfg := &CommandConfig{
		Name: use,
		Flag: &FlagsConfig{
			Persistent: map[string]func(*pflag.FlagSet) interface{}{},
			Local:      map[string]func(*pflag.FlagSet) interface{}{},
			Flags:      map[string]func(*pflag.FlagSet) (value interface{}){},
			Values:     FlagValueStore{},
		},
		FS:      f.FS,
		Streams: f.Streams,
		Help:    &helpConfig{},
		Factory: f,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	cmd = cfg.Build()
	return cmd
}

type Factory interface {
	GetFS() afero.Fs
	GetStreams() *Streams
	NewCommand(use string, opts ...CommandOption) *cobra.Command
}

type FactoryConfig struct {
	FS      afero.Fs
	Streams *Streams
}

type FactoryOption func(c *FactoryConfig)

func WithStreams(streams *Streams) FactoryOption {
	return func(c *FactoryConfig) {
		c.Streams = streams
	}
}

func WithFS(fs afero.Fs) FactoryOption {
	return func(c *FactoryConfig) {
		c.FS = fs
	}
}

func DefaultStreams() *Streams {
	return &Streams{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

var defaultFactoryOptions = []FactoryOption{
	WithFS(afero.NewOsFs()),
	WithStreams(DefaultStreams()),
}

func BuildFactory(opts ...FactoryOption) Factory {
	cfg := &FactoryConfig{}
	for _, opt := range defaultFactoryOptions {
		opt(cfg)
	}
	return &factory{cfg}
}

type Streams struct {
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader
}
