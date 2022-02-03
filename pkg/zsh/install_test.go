package zsh

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alex-held/dfctl/pkg/config"
	"github.com/alex-held/dfctl/pkg/env"
	"github.com/alex-held/dfctl/pkg/testutils"
)

func TestInstallThemes(t *testing.T) {
	log := testutils.Logger(t)

	cfg := &config.ConfigSpec{
		Themes: config.ThemesSpec{
			{
				ID:   "powerlevel10k/powerlevel10k",
				Name: "powerlevel10k",
				Repo: "romkatv/powerlevel10k",
				Kind: config.PLUGIN_GITHUB,
			},
		},
	}

	dir := testutils.TempDir(t)
	themeDir := filepath.Join(dir, "omz", "custom", "themes")
	err := os.MkdirAll(themeDir, os.ModePerm)
	assert.NoError(t, err)

	env.SetOverrides(&env.EnvConfig{
		OMZ: filepath.Join(dir, "omz"),
	})

	fmt.Println("tmpDir: " + themeDir)
	results := InstallThemes(cfg)
	for theme, installResult := range results {
		log.Debug().Msgf("asserting install results of %s", theme.ID)
		assert.NoError(t, installResult.Err)
		assert.True(t, installResult.Installed)
		assert.DirExists(t, theme.Path())
	}
}

func TestInstallPlugin(t *testing.T) {
	log := testutils.Logger(t)
	cfg := &config.ConfigSpec{
		Plugins: config.PluginsSpec{
			Custom: config.PluginsList{
				{
					ID:   "zsh-autosuggestions",
					Name: "zsh-autosuggestions",
					Repo: "zsh-users/zsh-autosuggestions",
					Kind: config.PLUGIN_GITHUB,
				},
				{
					ID:   "fast-syntax-highlighting",
					Name: "fast-syntax-highlighting",
					Repo: "zdharma/fast-syntax-highlighting",
					Kind: config.PLUGIN_GITHUB,
				},
			},
		},
	}

	omzDir := testutils.TempDir(t, "omz")
	pluginsDir := filepath.Join(omzDir, "custom", "plugins")
	err := os.MkdirAll(pluginsDir, os.ModePerm)
	assert.NoError(t, err)

	env.SetOverrides(&env.EnvConfig{
		OMZ: omzDir,
	})

	results := InstallPlugins(cfg)
	for plugin, installResult := range results {
		log.Debug().Msgf("asserting install results of %s", plugin.ID)
		assert.NoError(t, installResult.Err)
		assert.True(t, installResult.Installed)
		assert.DirExists(t, plugin.Path())
	}
}

func TestListInstallable(t *testing.T) {
	log := testutils.Logger(t)

	cfg := &config.ConfigSpec{
		Plugins: config.PluginsSpec{
			Custom: config.PluginsList{

				{
					ID:   "zsh-autosuggestions",
					Name: "zsh-autosuggestions",
					Repo: "zsh-users/zsh-autosuggestions",
					Kind: config.PLUGIN_GITHUB,
				},
				{
					ID:   "fast-syntax-highlighting",
					Name: "fast-syntax-highlighting",
					Repo: "zdharma/fast-syntax-highlighting",
					Kind: config.PLUGIN_GITHUB,
				},
			},
		},
	}

	omzDir := testutils.TempDir(t, "omz")
	configFileDir := testutils.TempDir(t, "dfctl.toml")

	env.SetOverrides(&env.EnvConfig{
		OMZ:    omzDir,
		Config: configFileDir,
	})

	err := config.SaveToPath(cfg, configFileDir)
	assert.NoError(t, err)

	installables := ListInstallables()

	for _, i := range installables {
		log.Debug().Str("id", i.Id()).Str("path", i.Path()).Bool("isInstalled", i.IsInstalled()).Msgf("asserting installable")
	}
}

func TestListUninstalled(t *testing.T) {
	log := testutils.Logger(t)

	cfg := &config.ConfigSpec{
		Plugins: config.PluginsSpec{
			Custom: config.PluginsList{
				{
					ID:   "zsh-autosuggestions",
					Name: "zsh-autosuggestions",
					Repo: "zsh-users/zsh-autosuggestions",
					Kind: config.PLUGIN_GITHUB,
				},
				{
					ID:   "fast-syntax-highlighting",
					Name: "fast-syntax-highlighting",
					Repo: "zdharma-continuum/fast-syntax-highlighting",
					Kind: config.PLUGIN_GITHUB,
				},
			},
		},
	}

	omzDir := testutils.TempDir(t, "omz")
	configFileDir := testutils.TempDir(t, "dfctl.toml")

	env.SetOverrides(&env.EnvConfig{
		OMZ:    omzDir,
		Config: configFileDir,
	})

	err := config.SaveToPath(cfg, configFileDir)
	assert.NoError(t, err)

	uninstalledBefore := ListUninstalled()

	uninstalledBeforeLen := len(uninstalledBefore)
	assert.NotEqual(t, 0, uninstalledBeforeLen)

	for _, i := range uninstalledBefore {
		log.Debug().Str("id", i.Id()).Str("path", i.Path()).Bool("isInstalled", i.IsInstalled()).Msgf("asserting installable")
	}

	_ = Install(uninstalledBefore...)

	uninstalledAfter := ListUninstalled()
	uninstalledAfterLen := len(uninstalledAfter)

	assert.Equal(t, 0, uninstalledAfterLen)
}
