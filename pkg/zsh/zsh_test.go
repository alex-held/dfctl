package zsh

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alex-held/dfctl-zsh/pkg/config"
)

func TestRender(t *testing.T) {
	cfg := &config.ConfigSpec{
		Theme: "powerlevel10k/powerlevel10k",
		Exports: map[string]string{
			"GOPATH":               "$HOME/go",
			"GOBIN":                "$HOME/go/bin",
			"GRADLE_HOME":          "$HOME/.gradle",
			"GO15VENDOREXPERIMENT": "1",
			"GO111MODULE":          "on",
		},
		Plugins: config.PluginsSpec{
			{
				Name: "brew",
				ID:   "brew",
				Kind: config.PLUGIN_OMZ,
			},
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
		Themes: config.ThemesSpec{
			{
				ID:   "powerlevel10k/powerlevel10k",
				Name: "powerlevel10k",
				Repo: "romkatv/powerlevel10k",
				Kind: config.PLUGIN_GITHUB,
			},
		},
		Aliases: map[string]string{
			"k": "kubectl",
		},
		Source: config.SourceSpec{
			Post: []string{
				"~/.p10k.zsh",
			},
		},
		Configs: config.ConfigsSpec{
			ZshOptions: map[string]bool{
				"BEEP":                    false,
				"no_beep":                 true,
				"case_glob":               false,
				"globdots":                true,
				"extendedglob":            true,
				"autocd":                  true,
				"brace_ccl":               true,
				"combining_chars":         true,
				"combining_charrc_quotes": true,
				"mail_warning":            false,
				"long_list_jobs":          true,
				"auto_resume":             true,
				"notify":                  true,
				"bg_nice":                 false,
				"hup":                     false,
				"check_jobs":              false,
				"correct":                 true,
				"complete_in_word":        true,
				"path_dirs":               true,
				"menu_complete":           true,
			},
			User: map[string]string{
				"EDITOR": "vim",
				"LANG":   "en_US.UTF-8",
			},
			Paths: []string{
				"$GOBIN",
				"$HOME/.devctl/sdks/go/current/bin",
			},
			OMZ: map[string]string{
				"ENABLE_CORRECTION": "true",
			},
		},
	}

	_ = config.SaveToPath(cfg, "/Users/dev/tmp/dfctl.toml")
	rendered, err := render(cfg)
	assert.NoError(t, err)

	fmt.Println(rendered)
}
