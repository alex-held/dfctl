package zsh

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ahmetb/go-linq"
	"github.com/go-git/go-git/v5"
	"github.com/rs/zerolog/log"

	"github.com/alex-held/dfctl-kit/pkg/env"

	"github.com/alex-held/dfctl/pkg/factory"
)

type Plugin struct {
	ID      string
	Repo    string
	Name    string
	Enabled bool
	Kind    RepoKind
}

func (p *Plugin) SetEnabled(enable bool) error {
	cfg, err := Load()
	if err != nil {
		return err
	}

	p.Enabled = enable

	for i := 0; i < len(cfg.Plugins.Custom); i++ {
		custom := cfg.Plugins.Custom[i]

		if custom.ID == p.ID {
			if enable {
				return nil
			}
			cfg.Plugins.Custom = append(cfg.Plugins.Custom[:i], cfg.Plugins.Custom[i+1:]...)
			return Save(cfg)
		}
	}

	// plugin is not yet in config
	if enable {
		cfg.Plugins.Custom = append(cfg.Plugins.Custom, *p.Spec())
		return Save(cfg)
	}

	return nil
}

func (p *Plugin) IsEnabled() bool {
	return p.Enabled
}

func PluginFromSpec(p *PluginSpec) *Plugin {
	return &Plugin{
		ID:      p.ID,
		Repo:    p.Repo,
		Name:    p.Name,
		Kind:    p.Kind,
		Enabled: p.Enabled,
	}
}

func strptr(s string) *string {
	return &s
}

func NewPlugin(repoUrn string, id, name *string) (p *Plugin) {
	repo := repoUrn[strings.Index(repoUrn, ":")+1:]
	kindStr := repoUrn[:strings.Index(repoUrn, ":")]

	if id == nil || *id == "" {
		id = strptr(filepath.Base(repoUrn))
	}
	if name == nil || *name == "" {
		name = strptr(*id)
	}

	return &Plugin{
		ID:      *id,
		Name:    *name,
		Repo:    repo,
		Kind:    ParsePluginKind(kindStr),
		Enabled: true,
	}
}

func (p *Plugin) Spec() *PluginSpec {
	return &PluginSpec{
		ID:      p.ID,
		Repo:    p.Repo,
		Name:    p.Name,
		Kind:    p.Kind,
		Enabled: p.Enabled,
	}
}

var ErrOMZCloneNotSupported = fmt.Errorf("plugin is of type OMZ and cannot be cloned")

func (p *Plugin) PluginName() string {
	return filepath.Base(p.Name)
}

func (p *Plugin) Clone() (pluginPath string, err error) {
	var url string

	name := filepath.Base(p.Name)
	pluginsDir := env.Plugins()
	pluginPath = filepath.Join(pluginsDir, name)

	switch p.Kind {
	case PLUGIN_GITHUB:
		url = fmt.Sprintf("https://github.com/%s", p.Name)
	case PLUGIN_GIT:
		url = p.Name
	case PLUGIN_OMZ:
		return "", ErrOMZCloneNotSupported
	}

	repo, err := git.PlainClone(pluginPath, false, &git.CloneOptions{URL: url})
	if err != nil {
		return pluginPath, err
	}
	_ = repo

	return pluginPath, nil
}

func (p *Plugin) Id() string { return p.ID }

func (p *Plugin) Install() (res InstallResult) {
	path := p.Path()

	if p.Kind == PLUGIN_OMZ {
		log.Debug().Msgf("plugin %s of kind omz does not need to be installed", p.ID)
		return InstallResult{Installed: false}
	}

	if _, statErr := os.Stat(path); statErr == nil {
		log.Debug().Msgf("plugin %s is already installed", p.ID)
		return InstallResult{Installed: false}
	}

	if _, err := git.PlainClone(path, false, &git.CloneOptions{URL: BuildRepositoryURI(p.Repo, p.Kind)}); err != nil {
		return InstallResult{Installed: false, Err: err}
	}

	cfg := MustLoad()

	if cfg.Plugins.ContainsWithRepo(p.Repo, p.Kind) {
		return InstallResult{Installed: true}
	}

	cfg.Plugins.Custom = append(cfg.Plugins.Custom, *p.Spec())
	if err := Save(cfg); err != nil {
		log.Error().Err(err).Msgf("unable to save plugin %s to config file", p.ID)
		return InstallResult{Installed: true, Err: err}
	}

	return InstallResult{Installed: true}
}

func BuildRepositoryURI(repo string, kind RepoKind) string {
	if kind == PLUGIN_GITHUB {
		return "https://github.com/" + repo
	}
	return repo
}

func (p *Plugin) Path() string {
	switch p.Kind {
	case PLUGIN_OMZ:
		return filepath.Join(env.OMZ(), "plugins", p.Name)
	default:
		return filepath.Join(env.Plugins(), p.Name)
	}
}

func PluginsQuery() (query linq.Query, err error) {
	cfg, err := Load()
	if err != nil {
		return linq.Query{}, err
	}

	return linq.
		From(cfg.Plugins).
		SelectT(func(p PluginSpec) *Plugin {
			return PluginFromSpec(&p)
		}), nil
}

func ListInstalledPlugins() (plugins []*Plugin, err error) {
	q, err := PluginsQuery()
	if err != nil {
		return plugins, err
	}

	q.WhereT(func(plugin *Plugin) bool {
		return plugin.IsInstalled()
	}).ToSlice(&plugins)

	return plugins, err
}

func (p *Plugin) IsInstalled() bool {
	_, err := factory.Default.Fs.Stat(p.Path())
	return err == nil
}

func (p *Plugin) GetKind() InstallableKind { return PluginInstallableKind }

func ListUninstalledPlugins() (plugins []*Plugin, err error) {
	q, err := PluginsQuery()
	if err != nil {
		return plugins, err
	}
	q.WhereT(func(plugin *Plugin) bool {
		return !plugin.IsInstalled()
	}).ToSlice(&plugins)

	return plugins, nil
}

func ListPlugins() (plugins []*Plugin, err error) {
	q, err := PluginsQuery()
	if err != nil {
		return plugins, err
	}
	q.ToSlice(&plugins)
	return plugins, err
}
