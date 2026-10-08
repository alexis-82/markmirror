package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Settings sono le preferenze persistenti dell'utente.
type Settings struct {
	Theme      string       `json:"theme"` // "system", "light" o "dark"
	SyncScroll bool         `json:"syncScroll"`
	Window     *WindowState `json:"window,omitempty"`
}

// WindowState è la posizione della finestra all'ultima chiusura. X, Y,
// Width e Height sono le dimensioni "normali" (non ingrandita), in pixel.
type WindowState struct {
	X         int32 `json:"x"`
	Y         int32 `json:"y"`
	Width     int32 `json:"width"`
	Height    int32 `json:"height"`
	Maximised bool  `json:"maximised"`
}

const (
	minWindowWidth  = 700
	minWindowHeight = 400
)

func defaultSettings() Settings {
	return Settings{Theme: "system", SyncScroll: true}
}

func settingsPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, appName, "settings.json"), nil
}

func loadSettings() Settings {
	s := defaultSettings()
	path, err := settingsPath()
	if err != nil {
		return s
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return s
	}
	return parseSettings(data)
}

// parseSettings interpreta settings.json, correggendo i valori non validi.
func parseSettings(data []byte) Settings {
	s := defaultSettings()
	if err := json.Unmarshal(data, &s); err != nil {
		return defaultSettings()
	}
	switch s.Theme {
	case "system", "light", "dark":
	default:
		s.Theme = "system"
	}
	if w := s.Window; w != nil {
		w.Width = max(w.Width, minWindowWidth)
		w.Height = max(w.Height, minWindowHeight)
	}
	return s
}

func (s Settings) save() error {
	path, err := settingsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// GetSettings restituisce le preferenze correnti al frontend.
func (a *App) GetSettings() Settings {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.settings
}

// updateSettings applica una modifica, la salva e la notifica al frontend.
func (a *App) updateSettings(change func(*Settings)) {
	a.mu.Lock()
	change(&a.settings)
	s := a.settings
	a.mu.Unlock()

	if err := s.save(); err != nil {
		runtime.LogErrorf(a.ctx, "salvataggio impostazioni: %v", err)
	}
	a.applyWindowTheme()
	runtime.EventsEmit(a.ctx, "settings:changed", s)
}

// applyWindowTheme allinea la barra del titolo di Windows al tema scelto.
func (a *App) applyWindowTheme() {
	theme := a.GetSettings().Theme
	applyMenuTheme(theme)
	switch theme {
	case "light":
		runtime.WindowSetLightTheme(a.ctx)
	case "dark":
		runtime.WindowSetDarkTheme(a.ctx)
	default:
		runtime.WindowSetSystemDefaultTheme(a.ctx)
	}
	refreshWindowFrame()
}
