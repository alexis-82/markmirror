package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Settings sono le preferenze persistenti dell'utente.
type Settings struct {
	Theme      string `json:"theme"` // "system", "light" o "dark"
	SyncScroll bool   `json:"syncScroll"`
}

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
	if err := json.Unmarshal(data, &s); err != nil {
		return defaultSettings()
	}
	switch s.Theme {
	case "system", "light", "dark":
	default:
		s.Theme = "system"
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
