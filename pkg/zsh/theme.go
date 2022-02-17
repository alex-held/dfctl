package zsh

import (
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/rs/zerolog/log"

	"github.com/alex-held/dfctl-kit/pkg/env"
)

type Theme struct {
	*ThemeSpec
}

func (theme *Theme) SetEnabled(enable bool) error {
	cfg := MustLoad()
	if enable {
		cfg.Theme = theme.Name
		return Save(cfg)
	}
	cfg.Theme = Default().Theme
	return Save(cfg)
}

func (theme *Theme) IsEnabled() bool {
	cfg := MustLoad()
	return cfg.Theme == theme.ID
}

func (theme *Theme) GetKind() InstallableKind { return ThemeInstallableKind }
func (theme *Theme) Id() string               { return theme.ID }

func (theme *Theme) IsInstalled() bool {
	_, err := os.Stat(theme.Path())
	return err == nil
}

func (theme *Theme) Path() string {
	if theme.Kind == PLUGIN_OMZ {
		return filepath.Join(env.OMZ(), "themes", theme.Name)
	}
	return filepath.Join(env.Themes(), theme.Name)
}

func (theme *Theme) Install() (res InstallResult) {
	path := theme.Path()
	if theme.Kind == PLUGIN_OMZ {
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
