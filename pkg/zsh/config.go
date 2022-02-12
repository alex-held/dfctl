package zsh

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"

	"github.com/alex-held/dfctl/pkg/dfpath"
	"github.com/alex-held/dfctl/pkg/factory"
)

type PluginSpec struct {
	ID      string   `toml:"id,required"`
	Name    string   `toml:"name,omitempty"`
	Repo    string   `toml:"repo,omitempty"`
	Kind    RepoKind `toml:"kind,omitempty"`
	Enabled bool     `toml:"enabled,required"`
}

type RepoKind string

func (k *RepoKind) Order() int {
	switch *k {
	case PLUGIN_GITHUB:
		return 0
	case PLUGIN_GIT:
		return 1
	case PLUGIN_OMZ:
		return 2
	default:
		return 1000
	}
}

const (
	PLUGIN_GITHUB RepoKind = "github"
	PLUGIN_GIT    RepoKind = "git"
	PLUGIN_OMZ    RepoKind = "omz"
)

type PluginsList []PluginSpec
type OMZPluginsList []OMZPlugin

func (o *OMZPluginsList) UnmarshalYAML(value *yaml.Node) error {
	var plugins []string
	if err := value.Decode(&plugins); err != nil {
		return err
	}

	for _, plugin := range plugins {
		*o = append(*o, OMZPlugin{plugin})
	}
	return nil
}

func (o OMZPluginsList) MarshalYAML() (interface{}, error) {
	var plugins []string
	for _, plugin := range o {
		plugins = append(plugins, plugin.ID)
	}
	return plugins, nil
}

func (o *OMZPluginsList) Enable(id string, enable bool) {
	switch enable {
	case true: // enable plugin
		found := false
		for _, plugin := range *o {
			found = plugin.ID == id
		}
		if !found {
			*o = append(*o, OMZPlugin{ID: id})
		}
	case false: // enable plugin
		for i, plugin := range *o {
			if plugin.ID == id {
				*o = append((*o)[:i], (*o)[i+1:]...)
				return
			}
		}
		return
	}
}

func OMZPluginList(omzs ...string) (plugins OMZPluginsList) {
	for _, omz := range omzs {
		plugins = append(plugins, OMZPlugin{ID: omz})
	}
	return plugins
}

type PluginsSpec struct {
	OMZ    OMZPluginsList `json:"omz,omitempty"`
	Custom PluginsList    `json:"custom,omitempty"`
}

func (omzs OMZPluginsList) PluginIDs() (plugins []string) {
	for _, plugin := range omzs {
		plugins = append(plugins, plugin.ID)
	}
	return plugins
}

func (s PluginsSpec) GetOMZ(id string) (omz *OMZPlugin, ok bool) {
	for _, omzPlugin := range s.OMZ {
		if omzPlugin.ID == id {
			return &omzPlugin, true
		}
	}
	return nil, false
}

func ParsePluginKind(kindStr string) RepoKind {
	var kind RepoKind
	switch strings.ToLower(kindStr) {
	case "omz":
		kind = PLUGIN_OMZ
	case "git":
		kind = PLUGIN_GIT
	case "gh":
		kind = PLUGIN_GITHUB
	default:
		err := fmt.Errorf("kind %s is not supported", kindStr)
		log.Error().Err(err).Msgf("failed to parse RepoKind")
		panic(err)
	}
	return kind
}

type PluginURN string

func (p PluginURN) GetScheme() RepoKind {
	urn := string(p)
	i := strings.Index(urn, ":")
	scheme := urn[:i]
	return ParsePluginKind(scheme)
}

func (p PluginURN) GetURI() string {
	urn := string(p)
	i := strings.Index(urn, ":")
	uri := urn[i+1:]
	if strings.Count(uri, "/") == 2 {
		return "https://github.com/" + uri
	}
	return uri
}

func (p PluginURN) GetID() string {
	uri := p.GetURI()
	return path.Base(uri)
}

func (s *PluginsSpec) ContainsOMZ(id string) bool {
	for _, plugin := range s.OMZ {
		if plugin.ID == id {
			return true
		}
	}
	return false
}

func (s *PluginsSpec) GetByIdOMZ(id string) (omz *OMZPlugin, ok bool) {
	for _, plugin := range s.OMZ {
		if plugin.ID == id {
			return &plugin, true
		}
	}
	return nil, false
}

func (s *PluginsSpec) ContainsWithRepo(repo string, kind RepoKind) bool {
	for _, pluginSpec := range s.Custom {
		if pluginSpec.Kind == kind && pluginSpec.Repo == repo {
			return true
		}
	}
	return false
}

type SourceSpec struct {
	Pre  []string `toml:"pre,omitempty"`
	Post []string `toml:"post,omitempty"`
}

type ConfigsSpec struct {
	Paths      []string          `toml:"path,omitempty"`
	User       map[string]string `toml:"user,omitempty"`
	OMZ        map[string]string `toml:"omz,omitempty"`
	ZshOptions map[string]bool   `toml:"zsh_options,omitempty"`
}

type ThemesSpec []ThemeSpec
type ThemeSpec struct {
	ID   string   `yaml:"id"`
	Name string   `yaml:"name,omitempty"`
	Repo string   `yaml:"repo,omitempty"`
	Kind RepoKind `yaml:"kind,omitempty"`
}

type ConfigSpec struct {
	Theme   string            `yaml:"theme,omitempty"`
	Plugins PluginsSpec       `yaml:"plugins,omitempty"`
	Themes  ThemesSpec        `yaml:"themes,omitempty"`
	Exports map[string]string `yaml:"exports,omitempty"`
	Configs ConfigsSpec       `yaml:"configs,omitempty"`
	Source  SourceSpec        `yaml:"source,omitempty"`
	Aliases map[string]string `yaml:"aliases,omitempty"`
}

type ConfigFormatter struct {
	ConfigFileType string
}
type ConfigFormatterOption func(formatter *ConfigFormatter)

func NewConfigFormatter(opts ...ConfigFormatterOption) (f *ConfigFormatter) {
	f = &ConfigFormatter{
		ConfigFileType: ".yaml",
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

func (cfg *ConfigSpec) Format(opts ...ConfigFormatterOption) (formatted string, err error) {
	formatter := NewConfigFormatter(opts...)

	switch formatter.ConfigFileType {
	case ".yaml", ".yml":
		if data, err := yaml.Marshal(cfg); err == nil {
			return string(data), nil
		}
	case ".toml":
		buf := &bytes.Buffer{}
		if err = toml.NewEncoder(buf).Encode(cfg); err == nil {
			return buf.String(), nil
		}
	}
	return "", err
}

func SaveToPath(cfg *ConfigSpec, path string) (err error) {
	err = factory.GetFS().MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return err
	}
	switch filepath.Ext(path) {
	case ".yaml", ".yml":
		if data, err := yaml.Marshal(cfg); err == nil {
			err = afero.WriteFile(factory.GetFS(), path, data, os.ModePerm)
			return err
		}
		return err
	case ".toml":
		b := &bytes.Buffer{}
		if err = toml.NewEncoder(b).Encode(cfg); err == nil {
			err = afero.WriteFile(factory.GetFS(), path, b.Bytes(), os.ModePerm)
			return err
		}
		return err
	default:
		err = fmt.Errorf("unsupported configFile extention %s", filepath.Ext(path))
		log.Error().Err(err).Msgf("unable to save config to path %s", path)
		panic(err)
	}
}

func Save(cfg *ConfigSpec) (err error) {
	return SaveToPath(cfg, dfpath.ConfigFile())
}

func LoadFromPath(path string) (cfg *ConfigSpec, err error) {
	data, err := afero.ReadFile(factory.GetFS(), path)
	if err != nil {
		return nil, err
	}
	cfg = &ConfigSpec{}

	switch filepath.Ext(path) {
	case ".yaml", ".yml":
		if err = yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	case ".toml":
		if err = toml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func MustLoadFromPath(path string) (cfg *ConfigSpec) {
	cfg, err := LoadFromPath(path)
	if err != nil {
		panic(err)
	}
	return cfg
}

func Default() (cfg *ConfigSpec) {
	return &ConfigSpec{
		Theme: "simple",
		Plugins: PluginsSpec{
			OMZ:    []OMZPlugin{},
			Custom: PluginsList{},
		},
		Aliases: map[string]string{},
		Source: SourceSpec{
			Post: []string{},
			Pre:  []string{},
		},
		Configs: ConfigsSpec{
			User:       map[string]string{},
			OMZ:        map[string]string{},
			Paths:      []string{},
			ZshOptions: map[string]bool{},
		},
		Exports: map[string]string{},
		Themes:  ThemesSpec{},
	}
}
func MustLoad() (cfg *ConfigSpec) {
	return MustLoadFromPath(dfpath.ConfigFile())
}

func Load() (cfg *ConfigSpec, err error) {
	return LoadFromPath(dfpath.ConfigFile())
}
