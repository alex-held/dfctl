package color

import (
	"fmt"
	"strings"
)

const (
	Black AnsiiCode = iota + 30
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White

	DarkGray AnsiiCode = 90
)

func ParseColorOrDefault(c string, def AnsiiCode) AnsiiCode {
	switch strings.ToLower(c) {
	case "black":
		return Black
	case "red":
		return Red
	case "green":
		return Green
	case "yellow":
		return Yellow
	case "blue":
		return Blue
	case "cyan":
		return Cyan
	case "magenta":
		return Magenta
	case "white":
		return White
	case "darkgray":
		return DarkGray
	default:
		return def
	}
}

const (
	Normal       AnsiiCode = 0
	Bold                   = 1
	Underlined             = 4
	Blinking               = 5
	ReverseVideo           = 7
)

type AnsiiCode int

func (a AnsiiCode) Paint(i interface{}) string {
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", a, i)
}

// Colorize returns the string s wrapped in ANSI code c, unless disabled is true.
func Colorize(i interface{}, opts ...AnsiiCode) (colorized string) {
	colorized = fmt.Sprintf("%v", i)
	for _, opt := range opts {
		colorized = opt.Paint(colorized)
	}
	return colorized
}
