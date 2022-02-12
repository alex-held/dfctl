package zsh

import (
	"os"
	"path/filepath"

	"github.com/ahmetb/go-linq"
	"github.com/rs/zerolog/log"

	"github.com/alex-held/dfctl/pkg/dfpath"
)

func InstallThemes(cfg *ConfigSpec) (installed map[Theme]InstallResult) {
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

func InstallPlugins(cfg *ConfigSpec) (results map[Plugin]InstallResult) {
	results = map[Plugin]InstallResult{}
	for _, plugin := range cfg.Plugins.Custom {
		p := PluginFromSpec(&plugin)
		results[*p] = p.Install()
	}
	return results
}

func QueryInstallable() (query linq.Query) {
	cfg, err := Load()
	if err != nil {
		log.Error().Err(err).Msgf("unable to load config")
	}

	return linq.
		From(cfg.Themes).
		SelectT(func(theme ThemeSpec) Installable {
			return &Theme{&theme}
		}).
		Concat(linq.
			From(cfg.Plugins.Custom).
			SelectT(func(plugin PluginSpec) Installable {
				return PluginFromSpec(&plugin)
			}),
		).
		Concat(linq.
			From(MustGetOMZPlugins()),
		)
}

func MustGetOMZPlugins() (plugins []Installable) {
	plugins, err := GetOMZPlugins()
	if err != nil {
		log.Error().Err(err).Msgf("getting omz plugins failed.")
		panic(err)
	}
	return plugins
}

func GetOMZPlugins() (plugins []Installable, err error) {
	path := filepath.Join(dfpath.OMZ(), "plugins")
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return plugins, err
	}

	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			plugins = append(plugins, &OMZPlugin{ID: dirEntry.Name()})
		}
	}
	return plugins, err
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
	case PluginInstallableKind:
		return "plugin"
	case ThemeInstallableKind:
		return "theme"
	default:
		panic("implement me!")
	}
}

const (
	PluginInstallableKind InstallableKind = iota
	ThemeInstallableKind
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
