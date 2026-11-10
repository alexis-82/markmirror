package main

import (
	goruntime "runtime"
	"testing"
)

func TestWindowStateToSave(t *testing.T) {
	current := WindowState{X: 10, Y: 20, Width: 800, Height: 600}
	snapshot := &WindowState{X: 1, Y: 2, Width: 1000, Height: 700, Maximised: true}

	tests := []struct {
		name     string
		snapshot *WindowState
		current  WindowState
		ok       bool
		want     WindowState
		wantOK   bool
	}{
		{"senza snapshot usa la geometria corrente", nil, current, true, current, true},
		{"senza snapshot e senza geometria non salva", nil, WindowState{}, false, WindowState{}, false},
		{"lo snapshot vince sulla geometria corrente", snapshot, current, true, *snapshot, true},
		{"lo snapshot basta anche senza geometria", snapshot, WindowState{}, false, *snapshot, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOK := windowStateToSave(tt.snapshot, tt.current, tt.ok)
			if got != tt.want || gotOK != tt.wantOK {
				t.Errorf("got (%+v, %v), want (%+v, %v)", got, gotOK, tt.want, tt.wantOK)
			}
		})
	}
}

// useTempConfigDir fa scrivere le impostazioni in una cartella temporanea.
func useTempConfigDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir) // Linux
	t.Setenv("AppData", dir)         // Windows
	t.Setenv("HOME", dir)            // macOS
}

func TestSaveWindowUsesPreFullscreenSnapshot(t *testing.T) {
	useTempConfigDir(t)
	snap := WindowState{X: 50, Y: 60, Width: 1100, Height: 750, Maximised: true}
	a := &App{settings: defaultSettings()}
	a.preFullscreen = &snap

	a.saveWindow()

	got := loadSettings().Window
	if got == nil || *got != snap {
		t.Fatalf("finestra salvata = %+v, want %+v", got, snap)
	}
}

func TestSaveWindowAfterClearKeepsPreviousGeometry(t *testing.T) {
	if goruntime.GOOS == "windows" {
		t.Skip("su Windows windowPlacement legge la finestra reale")
	}
	useTempConfigDir(t)
	prev := WindowState{X: 1, Y: 2, Width: 900, Height: 600}
	a := &App{settings: defaultSettings()}
	a.settings.Window = &prev
	a.preFullscreen = &WindowState{X: 9, Y: 9, Width: 9, Height: 9}
	a.clearWindowSnapshot()

	a.saveWindow()

	// Senza snapshot e senza placement (no-op fuori da Windows) non si salva nulla.
	if got := loadSettings().Window; got != nil {
		t.Fatalf("finestra salvata = %+v, want nessun salvataggio", got)
	}
	if a.settings.Window == nil || *a.settings.Window != prev {
		t.Fatalf("settings in memoria = %+v, want %+v", a.settings.Window, prev)
	}
}
