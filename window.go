package main

import (
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

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
		restoreWindowPlacement(*w)
	}
}

// saveWindow memorizza posizione e dimensione correnti della finestra.
func (a *App) saveWindow() {
	state, ok := windowPlacement()
	if !ok {
		return
	}
	a.mu.Lock()
	a.settings.Window = &state
	s := a.settings
	a.mu.Unlock()
	if err := s.save(); err != nil {
		runtime.LogErrorf(a.ctx, "salvataggio impostazioni: %v", err)
	}
}
