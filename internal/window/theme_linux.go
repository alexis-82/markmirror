//go:build linux

package window

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// SystemDark dice se il desktop usa il tema scuro. Serve prima che la pagina
// sia caricata (sfondo della finestra all'avvio), quando prefers-color-scheme
// non è ancora disponibile. Legge le stesse impostazioni di GTK, da cui
// WebKitGTK ricava il tema: GTK_THEME, poi settings.ini di GTK 3 (scritto
// anche da KDE), infine il color-scheme di GNOME. Nel dubbio vale "chiaro".
func SystemDark() bool {
	if theme := os.Getenv("GTK_THEME"); theme != "" {
		return isDarkThemeName(theme)
	}
	if dir, err := os.UserConfigDir(); err == nil {
		if f, err := os.Open(filepath.Join(dir, "gtk-3.0", "settings.ini")); err == nil {
			dark, found := gtkSettingsDark(f)
			f.Close()
			if found {
				return dark
			}
		}
	}
	return gnomePrefersDark()
}

// gtkSettingsDark legge settings.ini di GTK: gtk-application-prefer-dark-theme
// oppure un tema con variante scura nel nome. found è false se il file non
// dice nulla sul tema.
func gtkSettingsDark(r io.Reader) (dark, found bool) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		key, value, ok := strings.Cut(sc.Text(), "=")
		if !ok {
			continue
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		switch key {
		case "gtk-application-prefer-dark-theme":
			if value == "true" || value == "1" {
				return true, true
			}
			found = true
		case "gtk-theme-name":
			if isDarkThemeName(value) {
				return true, true
			}
			found = true
		}
	}
	return false, found
}

// isDarkThemeName riconosce i temi GTK scuri: "Adwaita:dark", "Breeze-Dark"...
func isDarkThemeName(name string) bool {
	name = strings.ToLower(name)
	return strings.HasSuffix(name, ":dark") || strings.HasSuffix(name, "-dark")
}

// gnomePrefersDark chiede a gsettings il color-scheme di GNOME. Se gsettings
// manca o non risponde in fretta, vale "chiaro": non deve ritardare l'avvio.
func gnomePrefersDark() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	out, err := exec.CommandContext(ctx, "gsettings", "get", "org.gnome.desktop.interface", "color-scheme").Output()
	return err == nil && strings.TrimSpace(string(out)) == "'prefer-dark'"
}
