// Package settings gestisce le preferenze persistenti (settings.json).
package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Settings sono le preferenze persistenti dell'utente.
type Settings struct {
	Theme        string       `json:"theme"` // "system", "light" o "dark"
	SyncScroll   bool         `json:"syncScroll"`
	CheckUpdates bool         `json:"checkUpdates"` // controllo aggiornamenti all'avvio
	Window       *WindowState `json:"window,omitempty"`
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
	MinWidth  = 700
	MinHeight = 400
)

// appDir è la cartella delle impostazioni sotto la config dell'utente. Deve
// restare uguale al nome dell'app: lì ci sono le impostazioni già salvate.
const appDir = "MarkMirror"

// Default restituisce le preferenze iniziali.
func Default() Settings {
	return Settings{Theme: "system", SyncScroll: true, CheckUpdates: true}
}

func filePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, appDir, "settings.json"), nil
}

// Load legge settings.json; se manca o non è leggibile usa i valori predefiniti.
func Load() Settings {
	s := Default()
	path, err := filePath()
	if err != nil {
		return s
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return s
	}
	return Parse(data)
}

// Parse interpreta settings.json, correggendo i valori non validi.
func Parse(data []byte) Settings {
	s := Default()
	if err := json.Unmarshal(data, &s); err != nil {
		return Default()
	}
	switch s.Theme {
	case "system", "light", "dark":
	default:
		s.Theme = "system"
	}
	if w := s.Window; w != nil {
		w.Width = max(w.Width, MinWidth)
		w.Height = max(w.Height, MinHeight)
	}
	return s
}

// Save scrive le preferenze in settings.json.
func (s Settings) Save() error {
	path, err := filePath()
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
