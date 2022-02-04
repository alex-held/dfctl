package env

import (
	"context"
	"os"
	"reflect"
	"strings"

	"github.com/sethvargo/go-envconfig"
)

type envConfigOverrides struct {
	*EnvConfig
	Environment
}

var overrides = &envConfigOverrides{
	Environment: NewEnvironment(),
}

type Environment map[string]string
type EnvironmentOption func(cfg Environment)

func Except(excepts ...string) EnvironmentOption {
	return func(cfg Environment) {
		cfg = cfg.Without(excepts...)
	}
}

func With(environment Environment) EnvironmentOption {
	return func(cfg Environment) {
		cfg.With(environment)
	}
}

func WithOsEnviron() EnvironmentOption {
	return func(cfg Environment) {
		cfg.With(OsEnvironment())
	}
}

func OsEnvironment() (env Environment) {
	env = Environment{}
	for _, e := range os.Environ() {
		i := strings.Index(e, "=")
		envKey := e[:i]
		envVal := e[i+1:]
		env[envKey] = envVal
	}
	return env
}

func (env Environment) With(environment Environment) Environment {
	for key, val := range environment {
		env[key] = val
	}
	return env
}

func (env Environment) Without(excepts ...string) Environment {
	for _, except := range excepts {
		delete(env, except)
	}
	return env
}

func NewEnvironment(opts ...EnvironmentOption) (environment Environment) {
	environment = Environment{}
	for _, opt := range opts {
		opt(environment)
	}
	return environment
}

func ClearOverrides() {
	overrides.EnvConfig = nil
	overrides.Environment = map[string]string{}
}

func SetOverrides(cfg *EnvConfig) {
	overrides.EnvConfig = cfg
}

func SetEnvironmentOverrides(environment Environment) {
	overrides.Environment = environment
}

func GetEnvironmentOverrides() Environment {
	return overrides.Environment
}

func GetOverrides() (override *EnvConfig, ok bool) {
	override = overrides.EnvConfig
	return override, override != nil
}

type EnvConfig struct {
	Home string `env:"DFCTL_HOME,default=$HOME/.config/dfctl"`
	OMZ  string `env:"DFCTL_OMZ,default=$HOME/.config/dfctl/omz"`
	// Config         string `env:"DFCTL_CONFIG,default=$HOME/.config/dfctl/config.yaml"`
	ConfigFileType string `env:"DFCTL_CONFIGFILE_TYPE,default=.yaml"`
	Config         string `env:"DFCTL_CONFIG,default=$HOME/.config/dfctl/devctl.yaml"`
}

func (e *EnvConfig) overrideDefaults() {
	if override, ok := GetOverrides(); ok {
		oVal := reflect.ValueOf(override).Elem()
		numI := oVal.NumField()
		for i := 0; i < numI; i++ {
			field := oVal.Field(i)
			if !field.IsZero() {
				reflect.ValueOf(e).Elem().Field(i).Set(field)
			}
		}
	}
}

func MustLoad() *EnvConfig {
	env, _ := Load()
	return env
}

func Load() (*EnvConfig, error) {
	env := &EnvConfig{}

	environment, err := GetEnvironment()
	if err != nil {
		return env, err
	}
	lookuper := envconfig.MultiLookuper(
		envconfig.OsLookuper(),
		envconfig.MapLookuper(environment),
	)
	if err := envconfig.ProcessWith(context.Background(), env, lookuper); err != nil {
		return nil, err
	}

	env.overrideDefaults()

	return env, nil
}

func GetEnvironment() (environment Environment, err error) {
	environment = NewEnvironment(WithOsEnviron())
	environmentOverrides := GetEnvironmentOverrides()
	for key, val := range environmentOverrides {
		environment[key] = val
	}
	return environment, nil
}
