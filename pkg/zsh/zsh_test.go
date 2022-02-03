package zsh

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alex-held/dfctl/pkg/config"
)

func TestRender(t *testing.T) {
	cfg := &config.ConfigSpec{
		Theme: "powerlevel10k/powerlevel10k",
		Exports: map[string]string{
			"GOPATH":                          "$HOME/go",
			"GOBIN":                           "$HOME/go/bin",
			"GOROOT":                          "$HOME/.devctl/sdks/go/current",
			"GRADLE_HOME":                     "$HOME/.gradle",
			"GO15VENDOREXPERIMENT":            "1",
			"GO111MODULE":                     "on",
			"NVM_DIR":                         "$HOME/.nvm",
			"VISUAL":                          "code-insiders",
			"PAGER":                           "bat",
			"BAT_CONFIG_PATH":                 "$HOME/.config/bat/bat.conf",
			"BROWSER":                         "chrome",
			"ZSH_AUTOSUGGEST_BUFFER_MAX_SIZE": "20",
			"LANG":                            "en_US.UTF-8",
			"LC_TYPE":                         "en_US.UTF-8",
			"LC_ALL":                          "en_US.UTF-8",
			"JB_GOLAND_DIR":                   "$HOME/Library/Application Support/JetBrains/Toolbox/apps/Goland/ch-0/211.7142.13",
			"FZF_DEFAULT_COMMAND":             "rg --files --no-ignore --hidden --follow -g '!{.git,node_modules}/*' 2> /dev/null",
			"FZF_DEFAULT_OPTS":                "--ansi --layout=default --info=inline --height=50% --multi --preview-window=right:50% --preview-window=sharp --preview-window=cycle --preview '([[ -f {} ]] && (bat --style=numbers --color=always --theme=gruvbox-dark --line-range :500 {} || cat {})) || ([[ -d {} ]] && (tree -C {} | less)) || echo {} 2> /dev/null | head -200' --prompt='λ -> ' --pointer='|>' --marker='✓' --bind 'ctrl-e:execute(nvim {} < /dev/tty > /dev/tty 2>&1)' > selected --bind 'ctrl-v:execute(code {+})'",
			"FZF_CTRL_T_COMMAND":              "$FZF_DEFAULT_COMMAND",
		},
		Plugins: config.PluginsSpec{
			OMZ: []string{
				"ag",
				"autojump",
				"brew",
				"colored-man-pages",
				"docker",
				"extract",
				"fd",
				"fzf",
				"gh",
				"git",
				"golang",
				"man",
				"nmap",
				"node",
				"npm",
				"nvm",
				"pip",
				"pipenv",
				"ripgrep",
				"sdk",
				"ssh-agent",
				"sudo",
				"yarn",
			},
			Custom: config.PluginsList{
				{
					ID:   "zsh-autosuggestions",
					Name: "zsh-autosuggestions",
					Repo: "zsh-users/zsh-autosuggestions",
					Kind: config.PLUGIN_GITHUB,
				},
				{
					ID:   "zsh-completions",
					Name: "zsh-completions",
					Repo: "zsh-users/zsh-completions",
					Kind: config.PLUGIN_GITHUB,
				},
				// {
				// 	ID:   "zstyle-completions",
				// 	Name: "zstyle-completions",
				// 	Repo: "zshzoo/zstyle-completions",
				// 	Kind: config.PLUGIN_GITHUB,
				// },
				{
					ID:   "fast-syntax-highlighting",
					Name: "fast-syntax-highlighting",
					Repo: "zdharma-continuum/fast-syntax-highlighting",
					Kind: config.PLUGIN_GITHUB,
				},
				{
					ID:   "zsh-fzf-history-search",
					Name: "zsh-fzf-history-search",
					Repo: "joshskidmore/zsh-fzf-history-search",
					Kind: config.PLUGIN_GITHUB,
				},
				{
					ID:   "fzf-tab",
					Name: "fzf-tab",
					Repo: "Aloxaf/fzf-tab",
					Kind: config.PLUGIN_GITHUB,
				},
				{
					ID:   "zfzf",
					Name: "zfzf",
					Repo: "b0o/zfzf",
					Kind: config.PLUGIN_GITHUB,
				},
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
			"k":         "kubectl",
			"ls":        "exa -b --links --long -a --git",
			"l":         "exa -@ --git  -H -g -a --group-directories-first --long --modified",
			"zshconfig": "dfctl config edit",
			"reload!":   "source <(dfctl zsh source)",
			"cdg":       "cd $GOPATH/src/github.com/alex-held",
			"cdr":       "cd ~/source/repos",
			"dl":        "cd ~/Downloads",
			"gs":        "git status --find-renames --untracked-files --ahead-behind --verbose",
			"grep":      "grep --color=auto",
			"fgrep":     "fgrep --color=auto",
			"egrep":     "egrep --color=auto",
			"sudo":      "sudo ",
			"flush":     "dscacheutil -flushcache && killall -HUP mDNSResponder",
		},
		Source: config.SourceSpec{
			Post: []string{
				"~/.p10k.zsh",
			},
		},
		Configs: config.ConfigsSpec{
			ZshOptions: map[string]bool{
				"BEEP":             false,
				"no_beep":          true,
				"case_glob":        false,
				"globdots":         true,
				"extendedglob":     true,
				"autocd":           true,
				"brace_ccl":        true,
				"combining_chars":  true,
				"rc_quotes":        true,
				"mail_warning":     false,
				"long_list_jobs":   true,
				"auto_resume":      true,
				"notify":           true,
				"bg_nice":          false,
				"hup":              false,
				"check_jobs":       false,
				"correct":          true,
				"complete_in_word": true,
				"path_dirs":        true,
				"auto_menu":        false,
				"auto_list":        false,
				"always_to_end":    true,
				"menu_complete":    true,
				"COMPLETE_ALIASES": true,
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
	_ = config.Save(cfg)
	rendered, err := render(cfg)
	assert.NoError(t, err)

	fmt.Println(rendered)
}
