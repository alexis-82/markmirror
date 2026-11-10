package main

import (
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"markmirror-go/internal/settings"
)

// GetSettings restituisce le preferenze correnti al frontend.
func (a *App) GetSettings() settings.Settings {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.settings
}

// updateSettings applica una modifica, la salva e la notifica al frontend.
func (a *App) updateSettings(change func(*settings.Settings)) {
	a.mu.Lock()
	change(&a.settings)
	s := a.settings
	a.mu.Unlock()

	if err := s.Save(); err != nil {
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
