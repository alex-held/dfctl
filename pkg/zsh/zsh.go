package zsh

import (
	"strings"
	"text/template"

	"github.com/alex-held/dfctl/pkg/config"
	"github.com/alex-held/dfctl/pkg/dfpath"
)

func Source() (rendered string, err error) {
	cfg, err := config.Load()
	if err != nil {
		return "", err
	}

	rendered, err = render(cfg)
	return rendered, err
}

func render(cfg *config.ConfigSpec) (rendered string, err error) {
	tpl := template.New("zshrc")
	parse, err := tpl.Parse(tmpl)
	if err != nil {
		return rendered, err
	}

	data := struct {
		OMZ_HOME    string
		Theme       string
		Plugins     []string
		Paths       []string
		Exports     map[string]string
		Aliases     map[string]string
		PostSources []string
		PreSources  []string
		UserConfigs map[string]string
		OMZConfigs  map[string]string
		ZshOptions  map[string]bool
	}{
		OMZ_HOME:    dfpath.OMZ(),
		Theme:       cfg.Theme,
		Paths:       cfg.Configs.Paths,
		Exports:     cfg.Exports,
		Aliases:     cfg.Aliases,
		PostSources: cfg.Source.Post,
		PreSources:  cfg.Source.Pre,
		UserConfigs: cfg.Configs.User,
		ZshOptions:  cfg.Configs.ZshOptions,
		OMZConfigs:  cfg.Configs.OMZ,
	}

	q, err := PluginsQuery()
	if err != nil {
		return rendered, err
	}
	q.SelectT(func(p *Plugin) string {
		return p.PluginName()
	}).ToSlice(&data.Plugins)

	sb := &strings.Builder{}

	if err := parse.Execute(sb, &data); err != nil {
		return rendered, err
	}

	rendered = sb.String()
	return rendered, nil
}

var tmpl = `
###############################################################################
# GLOBALS
##
export ZSH="{{ .OMZ_HOME }}"

###############################################################################
# EXPORTS
##
{{- if .Exports }}
{{- range $key, $val := .Exports }}
export {{ $key }}="{{ $val -}}"
{{- end -}}
{{ end }}

###############################################################################
# PATH
##
typeset -U path
{{- if .Paths }}
path+=(
	{{- range $path := .Paths }}
	{{ $path -}}
	{{ end }}
)
{{ end }}


###############################################################################
# OMZ CONFIG
##
ZSH_THEME="{{ .Theme }}"

{{- if .OMZConfigs }}
{{- range $option, $value := .OMZConfigs }}
{{ $option }}="{{ $value }}"
{{ end -}}
{{ end }}


###############################################################################
# PLUGINS
##
plugins=(
        {{- range $plugin := .Plugins }}
		{{ $plugin -}}
		{{ end }}
)


source $ZSH/oh-my-zsh.sh

###############################################################################
# USER CONFIG
##
{{- if .UserConfigs }}
{{- range $key, $val := .UserConfigs }}
export {{ $key }}="{{ $val -}}"
{{ end -}}
{{ end }}


###############################################################################
# ALIASES
##
{{- if .Aliases }}
{{- range $alias, $command := .Aliases }}
alias {{ $alias }}="{{ $command }}"
{{- end -}}
{{ end }}



###############################################################################
# OPTIONS
##
{{- if .ZshOptions }}
{{- range $option, $enabled := .ZshOptions }}
{{ if $enabled }}setopt {{ $option -}} {{ else }}unsetopt {{ $option -}} {{ end -}}
{{ end -}}
{{ end }}



###############################################################################
# POST SOURCE
##
{{- if .PostSources }}
{{- range $script := .PostSources }}
[[ ! -f {{ $script }} ]] || source {{ $script }}
{{ end }}
{{ end -}}
`
