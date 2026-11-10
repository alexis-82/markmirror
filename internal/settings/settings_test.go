package settings

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseSettings(t *testing.T) {
	// File di una versione precedente: nessuna finestra salvata.
	s := parse([]byte(`{"theme":"dark","syncScroll":false}`))
	if s.Theme != "dark" || s.SyncScroll || s.Window != nil {
		t.Fatalf("old settings = %+v", s)
	}

	// Valori non validi corretti; dimensioni sotto il minimo portate al minimo.
	s = parse([]byte(`{"theme":"blu","window":{"x":-10,"y":20,"width":100,"height":50,"maximised":true}}`))
	if s.Theme != "system" {
		t.Errorf("theme = %q, want system", s.Theme)
	}
	want := WindowState{X: -10, Y: 20, Width: MinWidth, Height: MinHeight, Maximised: true}
	if s.Window == nil || *s.Window != want {
		t.Errorf("window = %+v, want %+v", s.Window, want)
	}

	// JSON rovinato: impostazioni predefinite.
	if s := parse([]byte(`{`)); s != Default() {
		t.Errorf("broken file = %+v", s)
	}
}

func TestParseSettingsCheckUpdates(t *testing.T) {
	// File di una versione precedente all'updater: controllo attivo.
	if s := parse([]byte(`{"theme":"dark"}`)); !s.CheckUpdates {
		t.Errorf("missing key: CheckUpdates = false, want true")
	}
	if s := parse([]byte(`{"checkUpdates":false}`)); s.CheckUpdates {
		t.Errorf("checkUpdates false: CheckUpdates = true")
	}
}

func TestSettingsRoundTrip(t *testing.T) {
	in := Default()
	in.Window = &WindowState{X: 100, Y: 80, Width: 1300, Height: 900, Maximised: true}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	out := parse(data)
	if out.Window == nil || *out.Window != *in.Window {
		t.Fatalf("round trip window = %+v, want %+v", out.Window, in.Window)
	}
}

func TestFilePathKeepsAppFolder(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir()) // Linux
	t.Setenv("AppData", t.TempDir())         // Windows
	path, err := filePath()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("MarkMirror", "settings.json")
	if !strings.HasSuffix(path, want) {
		t.Fatalf("path = %q, want suffix %q", path, want)
	}
}
