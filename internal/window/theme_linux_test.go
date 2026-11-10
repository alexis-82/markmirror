//go:build linux

package window

import (
	"strings"
	"testing"
)

func TestGtkSettingsDark(t *testing.T) {
	cases := []struct {
		name        string
		ini         string
		dark, found bool
	}{
		{"KDE Breeze Dark", "[Settings]\ngtk-application-prefer-dark-theme=true\ngtk-theme-name=Breeze\n", true, true},
		{"prefer-dark come numero", "[Settings]\ngtk-application-prefer-dark-theme = 1\n", true, true},
		{"tema scuro nel nome", "[Settings]\ngtk-theme-name=Adwaita-dark\n", true, true},
		{"tema chiaro esplicito", "[Settings]\ngtk-application-prefer-dark-theme=false\ngtk-theme-name=Breeze\n", false, true},
		{"nessuna informazione", "[Settings]\ngtk-cursor-theme-name=breeze_cursors\n", false, false},
		{"file vuoto", "", false, false},
	}
	for _, c := range cases {
		dark, found := gtkSettingsDark(strings.NewReader(c.ini))
		if dark != c.dark || found != c.found {
			t.Errorf("%s: dark=%v found=%v, want dark=%v found=%v", c.name, dark, found, c.dark, c.found)
		}
	}
}

func TestIsDarkThemeName(t *testing.T) {
	for name, want := range map[string]bool{
		"Adwaita:dark": true, "Breeze-Dark": true, "Adwaita-dark": true,
		"Breeze": false, "Adwaita": false, "Darkly": false,
	} {
		if got := isDarkThemeName(name); got != want {
			t.Errorf("isDarkThemeName(%q) = %v, want %v", name, got, want)
		}
	}
}
