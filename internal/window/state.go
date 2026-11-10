package window

import "markmirror-go/internal/settings"

// StateToSave sceglie la geometria da salvare alla chiusura e preferisce lo snapshot preso prima dello schermo intero.
func StateToSave(snapshot *settings.WindowState, current settings.WindowState, ok bool) (settings.WindowState, bool) {
	if snapshot != nil {
		return *snapshot, true
	}
	return current, ok
}
