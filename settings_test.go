package main

import (
	"encoding/json"
	"testing"
)

func TestParseSettings(t *testing.T) {
	// File di una versione precedente: nessuna finestra salvata.
	s := parseSettings([]byte(`{"theme":"dark","syncScroll":false}`))
	if s.Theme != "dark" || s.SyncScroll || s.Window != nil {
		t.Fatalf("old settings = %+v", s)
	}

	// Valori non validi corretti; dimensioni sotto il minimo portate al minimo.
	s = parseSettings([]byte(`{"theme":"blu","window":{"x":-10,"y":20,"width":100,"height":50,"maximised":true}}`))
	if s.Theme != "system" {
		t.Errorf("theme = %q, want system", s.Theme)
	}
	want := WindowState{X: -10, Y: 20, Width: minWindowWidth, Height: minWindowHeight, Maximised: true}
	if s.Window == nil || *s.Window != want {
		t.Errorf("window = %+v, want %+v", s.Window, want)
	}

	// JSON rovinato: impostazioni predefinite.
	if s := parseSettings([]byte(`{`)); s != defaultSettings() {
		t.Errorf("broken file = %+v", s)
	}
}

func TestSettingsRoundTrip(t *testing.T) {
	in := defaultSettings()
	in.Window = &WindowState{X: 100, Y: 80, Width: 1300, Height: 900, Maximised: true}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	out := parseSettings(data)
	if out.Window == nil || *out.Window != *in.Window {
		t.Fatalf("round trip window = %+v, want %+v", out.Window, in.Window)
	}
}
