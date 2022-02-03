package dfpath

import (
	"path/filepath"

	"github.com/alex-held/dfctl/pkg/env"
)

func Home() string       { return env.MustLoad().Home }
func ConfigFile() string { return env.MustLoad().Config }
func OMZ() string        { return env.MustLoad().OMZ }
func Themes() string     { return filepath.Join(OMZ(), "custom", "themes") }
func Plugins() string    { return filepath.Join(OMZ(), "custom", "plugins") }
