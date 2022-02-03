package zsh

import (
	"os"
	"path/filepath"

	"github.com/ahmetb/go-linq"
	"github.com/go-git/go-git/v5"
	"github.com/rs/zerolog/log"

	"github.com/alex-held/dfctl/pkg/config"
	"github.com/alex-held/dfctl/pkg/dfpath"
)

type Theme struct {
	*config.ThemeSpec
}

func (theme *Theme) GetKind() InstallableKind { return ThemeKind }
func (theme *Theme) Id() string               { return theme.ID }

func (theme *Theme) IsInstalled() bool {
	_, err := os.Stat(theme.Path())
	return err == nil
}

func (theme *Theme) Path() string {
	if theme.Kind == config.PLUGIN_OMZ {
		return filepath.Join(dfpath.OMZ(), "themes", theme.Name)
	}
	return filepath.Join(dfpath.Themes(), theme.Name)
}

func (theme *Theme) Install() (res InstallResult) {
	path := theme.Path()
	if theme.Kind == config.PLUGIN_OMZ {
		log.Debug().Msgf("plugin %s of kind omz does not need to be installed", theme.ID)
		return InstallResult{Installed: false}
	}
	if _, statErr := os.Stat(path); statErr == nil {
		log.Debug().Msgf("plugin %s is already installed", theme.ID)
		return InstallResult{Installed: false}
	}
	if _, err := git.PlainClone(path, false, &git.CloneOptions{URL: BuildRepositoryURI(theme.Repo, theme.Kind)}); err != nil {
		return InstallResult{Installed: false, Err: err}
	}
	return InstallResult{Installed: true}
}

type Installable interface {
	Id() string
	Install() (result InstallResult)
	IsInstalled() bool
	Path() string
	GetKind() InstallableKind
}

type InstallResult struct {
	Installed bool
	Err       error
}

func InstallThemes(cfg *config.ConfigSpec) (installed map[Theme]InstallResult) {
	installed = map[Theme]InstallResult{}
	for _, theme := range cfg.Themes {
		t := Theme{ThemeSpec: &theme}
		installed[t] = t.Install()
	}
	return installed
}

func Install(installables ...Installable) (results map[Installable]InstallResult) {
	results = map[Installable]InstallResult{}
	for _, installable := range installables {
		results[installable] = installable.Install()
	}
	return results
}

func InstallPlugins(cfg *config.ConfigSpec) (results map[Plugin]InstallResult) {
	results = map[Plugin]InstallResult{}
	for _, plugin := range cfg.Plugins.Custom {
		p := PluginFromSpec(&plugin)
		results[*p] = p.Install()
	}
	return results
}

type OMZPlugin struct {
	id string
}

func (p *OMZPlugin) Id() string {
	return p.id
}

func (p *OMZPlugin) Install() (result InstallResult) {
	return InstallResult{Installed: false}
}

func (p *OMZPlugin) IsInstalled() bool {
	return true
}

func (p *OMZPlugin) Path() string {
	return filepath.Join(dfpath.OMZ(), "plugins", p.Id())
}

func (p *OMZPlugin) GetKind() InstallableKind {
	return PluginKind
}

func QueryInstallable() (query linq.Query) {
	cfg, err := config.Load()
	if err != nil {
		log.Error().Err(err).Msgf("unable to load config")
	}

	return linq.
		From(cfg.Themes).
		SelectT(func(theme config.ThemeSpec) Installable {
			return &Theme{&theme}
		}).
		Concat(linq.
			From(cfg.Plugins.Custom).
			SelectT(func(plugin config.PluginSpec) Installable {
				return PluginFromSpec(&plugin)
			}),
		).
		Concat(linq.
			From(cfg.Plugins.OMZ).
			SelectT(func(plugin string) Installable {
				return &OMZPlugin{id: plugin}
			}))
}

func ListInstallables(predicateFns ...InstallablePredicateFn) (result []Installable) {
	query := QueryInstallable()
	for _, predicateFn := range predicateFns {
		query = query.WhereT(predicateFn)
	}

	query.ToSlice(&result)
	return result
}

func ListInstalled() (result []Installable) {
	QueryInstalled().
		WhereT(func(installable Installable) bool {
			return installable.IsInstalled()
		}).
		ToSlice(&result)

	return result
}

type InstallablePredicateFn func(installable Installable) bool

func query(predicateFn InstallablePredicateFn) (query linq.Query) {
	query = QueryInstallable()
	return query.WhereT(predicateFn)
}

func QueryInstalled() linq.Query {
	return query(func(i Installable) bool {
		return i.IsInstalled()
	})
}

func QueryUninstalled() linq.Query {
	return query(func(i Installable) bool {
		return !i.IsInstalled()
	})
}

func InstalledFilterFn(installed bool) InstallablePredicateFn {
	return func(i Installable) bool {
		return i.IsInstalled() == installed
	}
}
func PluginFilterFn() InstallablePredicateFn {
	return func(i Installable) bool {
		switch i.(type) {
		case *Plugin:
			return true
		default:
			return false
		}
	}
}

func ThemeFilterFn() InstallablePredicateFn {
	return func(i Installable) bool {
		switch i.(type) {
		case *Theme:
			return true
		default:
			return false
		}
	}
}

type InstallableKind int

func (i InstallableKind) String() string {
	switch i {
	case PluginKind:
		return "plugin"
	case ThemeKind:
		return "theme"
	default:
		panic("implement me!")
	}
}

const (
	PluginKind InstallableKind = iota
	ThemeKind
)

func KindFilterFn(kind InstallableKind) InstallablePredicateFn {
	return func(i Installable) bool {
		return i.GetKind() == kind
	}
}

func ListUninstalled() (result []Installable) {
	QueryInstallable().
		WhereT(func(installable Installable) bool {
			return !installable.IsInstalled()
		}).
		ToSlice(&result)
	return result
}
