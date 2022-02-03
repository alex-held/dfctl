package env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOverrideDefaults(t *testing.T) {
	override := &EnvConfig{
		Home:    "/foo",
		Plugins: "/bar",
		OMZ:     "/test",
	}

	sut := &EnvConfig{
		Home:    "~/.config/dfctl",
		Plugins: "~/.config/dfctl/plugins",
		OMZ:     "~/.config/dfctl/omz",
	}

	SetOverrides(override, nil)
	sut.overrideDefaults()

	assert.Equal(t, "/foo", sut.GetHomePath())
	assert.Equal(t, "/bar", sut.GetPluginsPath())
	assert.Equal(t, "/test", sut.GetOMZPath())
}

func TestLoad(t *testing.T) {
	tt := []struct {
		name      string
		env       map[string]string
		want      EnvConfig
		overrides *EnvConfig
	}{
		{
			name: "with environment values",
			env: map[string]string{
				"DFCTL_HOME":    "/env/home",
				"DFCTL_PLUGINS": "/env/plugins",
				"DFCTL_OMZ":     "/env/omz",
			},
			want: EnvConfig{
				Home:    "/env/home",
				Plugins: "/env/plugins",
				OMZ:     "/env/omz",
			},
		},
		{
			name: "with environment values and overrides",
			env: map[string]string{
				"DFCTL_HOME":    "/env/home",
				"DFCTL_PLUGINS": "/env/plugins",
				"DFCTL_OMZ":     "/env/omz",
			},
			want: EnvConfig{
				Home:    "/override/home",
				Plugins: "/env/plugins",
				OMZ:     "/env/omz",
			},
			overrides: &EnvConfig{
				Home: "/override/home",
			},
		},
		{
			name: "defaults",
			env: map[string]string{
				"HOME": "$HOME",
			},
			want: EnvConfig{
				Home:    "$HOME/.config/dfctl",
				Plugins: "$HOME/.config/dfctl/plugins",
				OMZ:     "$HOME/.config/dfctl/omz",
			},
		},
		{
			name: "defaults with overrides",
			env: map[string]string{
				"HOME": "$HOME",
			},
			want: EnvConfig{
				Home:    "/override/home",
				Plugins: "$HOME/.config/dfctl/plugins",
				OMZ:     "/override/omz",
			},
			overrides: &EnvConfig{
				Home: "/override/home",
				OMZ:  "/override/omz",
			},
		},
	}

	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {

			SetEnvironmentOverrides(tt.env)
			SetOverrides(tt.overrides, tt.env)

			gotEnv, err := Load()
			got := gotEnv.(*EnvConfig)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, *got)
		})
	}
}

func TestEnvironment_With(t *testing.T) {
	sut := Environment{"A": "A_VAL"}
	want := Environment{"A": "A_VAL", "B": "B_VAL"}
	got := sut.With(Environment{"B": "B_VAL"})
	assert.Equal(t, want, got)
}

func TestOsEnvironment_Without(t *testing.T) {
	sut := Environment{"A": "A_VAL", "B": "B_VAL"}
	want := Environment{"B": "B_VAL"}
	got := sut.Without("A")
	assert.Equal(t, want, got)
}

func TestOsEnvironment(t *testing.T) {
	want := len(os.Environ())
	got := len(OsEnvironment())
	assert.Equal(t, want, got)
}

func TestNewEnvironment(t *testing.T) {
	tt := []struct {
		name    string
		options []EnvironmentOption
		want    Environment
	}{
		{
			name:    "with os.Environ()",
			options: []EnvironmentOption{WithOsEnviron()},
			want:    OsEnvironment(),
		},
		{
			name:    "without os.Environ()",
			options: []EnvironmentOption{},
			want:    Environment{},
		},
		{
			name:    "with os.Environ_() | except HOME, UID",
			options: []EnvironmentOption{With(Environment{"HOME": "HOME_VALUE", "UID": "UID_VALUE", "OTHER": "OTHER_VALUE"}), Except("HOME", "UID")},
			want:    Environment{"OTHER": "OTHER_VALUE"},
		},
		{
			name:    "with environment",
			options: []EnvironmentOption{With(Environment{"FOO": "FOO_VALUE", "BAR": "BAR_VALUE"})},
			want:    Environment{"FOO": "FOO_VALUE", "BAR": "BAR_VALUE"},
		},
	}
	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {
			got := NewEnvironment(tt.options...)
			assert.Equal(t, tt.want, got)
		})
	}
}
