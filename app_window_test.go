package main

import (
	"net/http"
	"net/http/httptest"
	goruntime "runtime"
	"strings"
	"testing"

	"github.com/wailsapp/wails/v2/pkg/options"

	"markmirror-go/internal/settings"
	"markmirror-go/internal/window"
)

// useTempConfigDir fa scrivere le impostazioni in una cartella temporanea.
func useTempConfigDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir) // Linux
	t.Setenv("AppData", dir)         // Windows
	t.Setenv("HOME", dir)            // macOS
}

func TestSaveWindowUsesPreFullscreenSnapshot(t *testing.T) {
	useTempConfigDir(t)
	snap := settings.WindowState{X: 50, Y: 60, Width: 1100, Height: 750, Maximised: true}
	a := &App{settings: settings.Default()}
	a.preFullscreen = &snap

	a.saveWindow()

	got := settings.Load().Window
	if got == nil || *got != snap {
		t.Fatalf("finestra salvata = %+v, want %+v", got, snap)
	}
}

func TestSaveWindowAfterClearKeepsPreviousGeometry(t *testing.T) {
	if goruntime.GOOS == "windows" {
		t.Skip("su Windows windowPlacement legge la finestra reale")
	}
	useTempConfigDir(t)
	prev := settings.WindowState{X: 1, Y: 2, Width: 900, Height: 600}
	a := &App{settings: settings.Default()}
	a.settings.Window = &prev
	a.preFullscreen = &settings.WindowState{X: 9, Y: 9, Width: 9, Height: 9}
	a.clearWindowSnapshot()

	a.saveWindow()

	// Senza snapshot e senza placement (no-op fuori da Windows) non si salva nulla.
	if got := settings.Load().Window; got != nil {
		t.Fatalf("finestra salvata = %+v, want nessun salvataggio", got)
	}
	if a.settings.Window == nil || *a.settings.Window != prev {
		t.Fatalf("settings in memoria = %+v, want %+v", a.settings.Window, prev)
	}
}

func TestWindowBackgroundFollowsTheme(t *testing.T) {
	dark := options.RGBA{R: 0x1e, G: 0x1e, B: 0x1e, A: 255}
	light := options.RGBA{R: 255, G: 255, B: 255, A: 255}
	system := light
	if window.SystemDark() {
		system = dark
	}
	cases := []struct {
		theme string
		want  options.RGBA
	}{
		{"dark", dark},
		{"light", light},
		{"system", system},
	}
	for _, c := range cases {
		a := &App{settings: settings.Settings{Theme: c.theme}}
		if got := *a.windowBackground(); got != c.want {
			t.Errorf("theme %q: background = %+v, want %+v", c.theme, got, c.want)
		}
	}
}

func TestThemedIndexWritesSavedTheme(t *testing.T) {
	const page = `<!DOCTYPE html><html lang="it"><head></head></html>`
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(page))
	})
	get := func(theme, path string) string {
		a := &App{settings: settings.Settings{Theme: theme}}
		rec := httptest.NewRecorder()
		a.themedIndex(next).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		return rec.Body.String()
	}

	if got := get("dark", "/"); !strings.Contains(got, `<html data-theme="dark" lang="it">`) {
		t.Errorf("dark: %s", got)
	}
	if got := get("light", "/index.html"); !strings.Contains(got, `<html data-theme="light" lang="it">`) {
		t.Errorf("light: %s", got)
	}
	// "system" lo decide lo script della pagina; gli altri file passano intatti.
	if got := get("system", "/"); got != page {
		t.Errorf("system: %s", got)
	}
	if got := get("dark", "/assets/app.js"); got != page {
		t.Errorf("other path modified: %s", got)
	}
}
