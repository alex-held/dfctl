package zsh

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/alex-held/dfctl/pkg/dfpath"
)

type OMZPlugin struct {
	ID string
}

func (p *OMZPlugin) MarshalTOML() ([]byte, error) {
	return []byte(p.ID), nil
}

func (p *OMZPlugin) UnmarshalText(text []byte) error {
	p.ID = string(text)
	return nil
}

var ErrUnmarshalInvalidTypeCast = errors.New("unable to cast data type to expected")

func (p *OMZPlugin) UnmarshalTOML(i interface{}) error {
	if id, ok := i.(string); ok {
		p.ID = id
		return nil
	}
	return fmt.Errorf("cannot %v (%T) cast to string: %w", i, i, ErrUnmarshalInvalidTypeCast)
}

func (p *OMZPlugin) Enable(enable bool) error {
	cfg, err := Load()
	if err != nil {
		return err
	}

	for i := 0; i < len(cfg.Plugins.OMZ); i++ {
		omz := cfg.Plugins.OMZ[i]

		if omz.ID == p.ID {
			if enable {
				return nil
			}
			cfg.Plugins.OMZ = append(cfg.Plugins.OMZ[:i], cfg.Plugins.OMZ[i+1:]...)
			return Save(cfg)
		}
	}

	// plugin is not yet in config
	if enable {
		cfg.Plugins.OMZ = append(cfg.Plugins.OMZ, *p)
		return Save(cfg)
	}

	return nil
}

func (p *OMZPlugin) IsEnabled() bool {
	cfg := MustLoad()
	return cfg.Plugins.ContainsOMZ(p.ID)
}

func (p *OMZPlugin) Id() string {
	return p.ID
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
	return PluginInstallableKind
}
