package main

import (
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"markmirror-go/internal/settings"
	"markmirror-go/internal/window"
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
	window.ApplyMenuTheme(theme)
	switch theme {
	case "light":
		runtime.WindowSetLightTheme(a.ctx)
	case "dark":
		runtime.WindowSetDarkTheme(a.ctx)
	default:
		runtime.WindowSetSystemDefaultTheme(a.ctx)
	}
	window.RefreshFrame()
}

// windowStartState indica a Wails se aprire la finestra ingrandita.
func (a *App) windowStartState() options.WindowStartState {
	if w := a.GetSettings().Window; w != nil && w.Maximised {
		return options.Maximised
	}
	return options.Normal
}

// restoreWindow riporta la finestra alla posizione e dimensione salvate.
// Lo stato ingrandito è già applicato da Wails tramite windowStartState.
func (a *App) restoreWindow() {
	if w := a.GetSettings().Window; w != nil {
		window.RestorePlacement(*w)
	}
}

// toggleFullscreen entra o esce dallo schermo intero (menu e F11).
// Lo stato finale si decide prima dell'azione: su GTK il passaggio è
// asincrono e rileggerlo subito restituirebbe ancora il vecchio stato.
func (a *App) toggleFullscreen(cd *menu.CallbackData) {
	enter := !runtime.WindowIsFullscreen(a.ctx)
	if enter {
		a.snapshotWindow()
		runtime.WindowFullscreen(a.ctx)
	} else {
		runtime.WindowUnfullscreen(a.ctx)
		a.clearWindowSnapshot()
	}
	// Wails inverte già la spunta nativa al clic e con F11, quindi di norma
	// Checked vale già enter. Lo forziamo solo nel modello: se lo stato era
	// cambiato per vie esterne, la spunta si riallinea al toggle successivo.
	// Niente MenuUpdateApplicationMenu: su Linux (Wails v2.16) ricrea un
	// menubar mai mostrato e scollega quello visibile, e il clic successivo
	// manda in panic handleMenuItemClick.
	cd.MenuItem.Checked = enter
}

// snapshotWindow memorizza la geometria prima dello schermo intero, che
// altrimenti alla chiusura verrebbe salvata con le dimensioni del monitor.
func (a *App) snapshotWindow() {
	state, ok := window.Placement()
	if !ok {
		return
	}
	a.mu.Lock()
	a.preFullscreen = &state
	a.mu.Unlock()
}

func (a *App) clearWindowSnapshot() {
	a.mu.Lock()
	a.preFullscreen = nil
	a.mu.Unlock()
}

// saveWindow memorizza posizione e dimensione correnti della finestra.
func (a *App) saveWindow() {
	current, placed := window.Placement()
	a.mu.Lock()
	state, ok := window.StateToSave(a.preFullscreen, current, placed)
	if !ok {
		a.mu.Unlock()
		return
	}
	a.settings.Window = &state
	s := a.settings
	a.mu.Unlock()
	if err := s.Save(); err != nil {
		runtime.LogErrorf(a.ctx, "salvataggio impostazioni: %v", err)
	}
}
