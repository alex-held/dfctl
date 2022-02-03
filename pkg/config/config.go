package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog/log"

	"github.com/alex-held/dfctl/pkg/dfpath"
)

type OhMyZshPluginSpec struct {
	Name string
	URL  string
}

type OhMyZshPluginsSpec []OhMyZshPluginSpec

type PluginSpec struct {
	ID   string     `toml:"id,omitempty"`
	Repo string     `toml:"repo,omitempty"`
	Name string     `toml:"name,omitempty"`
	Kind PluginKind `toml:"kind,omitempty"`
}

type PluginKind string

const (
	PLUGIN_GITHUB PluginKind = "github"
	PLUGIN_GIT    PluginKind = "git"
	PLUGIN_OMZ    PluginKind = "omz"
)

type PluginsList []PluginSpec
type PluginsSpec struct {
	OMZ    []string    `json:"omz,omitempty"`
	Custom PluginsList `json:"custom,omitempty"`
}

func ParsePluginKind(kindStr string) PluginKind {
	var kind PluginKind
	switch strings.ToLower(kindStr) {
	case "omz":
		kind = PLUGIN_OMZ
	case "git":
		kind = PLUGIN_GIT
	case "gh":
		kind = PLUGIN_GITHUB
	default:
		err := fmt.Errorf("kind %s is not supported", kindStr)
		log.Error().Err(err).Msgf("failed to parse PluginKind")
		panic(err)
	}
	return kind
}

type PluginURN string

func (p PluginURN) GetScheme() PluginKind {
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
		if plugin == id {
			return true
		}
	}
	return false
}

func (s *PluginsSpec) ContainsWithRepo(repo string, kind PluginKind) bool {
	for _, pluginSpec := range s.Custom {
		if pluginSpec.Kind == kind && pluginSpec.Repo == repo {
			return true
		}
	}
	return false
}

type SourceSpec struct {
	Post []string `toml:"post,omitempty"`
	Pre  []string `toml:"pre,omitempty"`
}

type ConfigsSpec struct {
	User       map[string]string `toml:"user,omitempty"`
	OMZ        map[string]string `toml:"omz,omitempty"`
	Paths      []string          `toml:"path,omitempty"`
	ZshOptions map[string]bool   `toml:"zsh_options,omitempty"`
}

type ThemesSpec []ThemeSpec
type ThemeSpec struct {
	ID   string     `toml:"id,omitempty"`
	Name string     `toml:"name,omitempty"`
	Repo string     `toml:"repo,omitempty"`
	Kind PluginKind `toml:"kind,omitempty"`
}

type ConfigSpec struct {
	Theme   string            `toml:"theme,omitempty"`
	Plugins PluginsSpec       `toml:"plugins,omitempty"`
	Aliases map[string]string `toml:"aliases,omitempty"`
	Source  SourceSpec        `toml:"source,omitempty"`
	Configs ConfigsSpec       `toml:"configs,omitempty"`
	Exports map[string]string `toml:"exports,omitempty"`
	Themes  ThemesSpec        `toml:"themes,omitempty"`
}

func (cfg *ConfigSpec) Toml() (out string, err error) {
	buf := &bytes.Buffer{}
	if err = toml.NewEncoder(buf).Encode(cfg); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func SaveToPath(cfg *ConfigSpec, path string) (err error) {
	err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	err = toml.NewEncoder(file).Encode(cfg)
	if err != nil {
		return err
	}
	return nil
}

func Save(cfg *ConfigSpec) (err error) {
	return SaveToPath(cfg, dfpath.ConfigFile())
}

func LoadFromPath(path string) (cfg *ConfigSpec, err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg = &ConfigSpec{}
	err = toml.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
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

func MustLoad() (cfg *ConfigSpec) {
	return MustLoadFromPath(dfpath.ConfigFile())
}

func Load() (cfg *ConfigSpec, err error) {
	return LoadFromPath(dfpath.ConfigFile())
}
