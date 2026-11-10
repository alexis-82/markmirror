package main

import (
	goruntime "runtime"
	"testing"

	"markmirror-go/internal/settings"
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
