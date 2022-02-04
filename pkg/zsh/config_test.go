package zsh

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Save(t *testing.T) {
	cfg := &ConfigSpec{
		Plugins: PluginsSpec{
			OMZ: OMZPluginList("brew"),
			Custom: PluginsList{
				{
					Name: "romkatv/powerlevel10k",
					Kind: PLUGIN_GITHUB,
				},
				{
					Name: "https://gitlab.com/IzzyOnDroid/repo",
					Kind: PLUGIN_GIT,
				},
			},
		},
	}

	tmpFile := filepath.Join(os.TempDir(), "dfctl-zsh.config.toml")
	fmt.Printf("config-file: %s\n", tmpFile)
	err := Save(cfg)
	assert.NoError(t, err)
}

func TestConfig_Load(t *testing.T) {
	cfg, err := Load()
	fmt.Printf("%#v", *cfg)
	assert.NoError(t, err)
}
