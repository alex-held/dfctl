package zsh

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alex-held/dfctl/pkg/env"
)

func TestPlugin_Clone(t *testing.T) {

	tt := []struct {
		Name string
		In   PluginSpec
		Err  error
	}{
		{
			Name: "clones github plugin from github",
			In: PluginSpec{
				Name: "romkatv/powerlevel10k",
				Kind: PLUGIN_GITHUB,
			},
			Err: nil,
		},
		{
			Name: "clones git plugin from gitlab",
			In: PluginSpec{
				Name: "https://gitlab.com/IzzyOnDroid/repo",
				Kind: PLUGIN_GIT,
			},
			Err: nil,
		},
		{
			Name: "mustn't clone omz plugin",
			In: PluginSpec{
				Name: "brew",
				Kind: PLUGIN_OMZ,
			},
			Err: ErrOMZCloneNotSupported,
		},
	}
	for _, tt := range tt {
		t.Run(tt.Name, func(t *testing.T) {
			tmpPath := filepath.Join(t.TempDir(), "dfctl-zsh", "plugins_test")
			env.SetOverrides(&env.EnvConfig{OMZ: tmpPath})

			p := PluginFromSpec(&tt.In)

			pluginPath, err := p.Clone()
			fmt.Printf("plugin %s cloned into %s\n", tt.In.Name, pluginPath)

			if tt.Err != nil {
				assert.ErrorAs(t, err, ErrOMZCloneNotSupported)
				assert.NoDirExists(t, pluginPath)
			} else {
				assert.NoError(t, err)
				assert.DirExists(t, pluginPath)
			}
		})
	}
}

func TestListPlugins(t *testing.T) {
	got, err := ListPlugins()
	assert.NoError(t, err)
	fmt.Printf("len: %v", len(got))
}

func TestListInstalledPlugins(t *testing.T) {
	got, err := ListInstalledPlugins()
	assert.NoError(t, err)
	fmt.Printf("len: %v", len(got))
}

func TestListUninstalledPlugins(t *testing.T) {
	got, err := ListUninstalledPlugins()
	assert.NoError(t, err)
	fmt.Printf("len: %v", len(got))
}
