package window

import (
	"testing"

	"markmirror-go/internal/settings"
)

func TestStateToSave(t *testing.T) {
	current := settings.WindowState{X: 10, Y: 20, Width: 800, Height: 600}
	snapshot := &settings.WindowState{X: 1, Y: 2, Width: 1000, Height: 700, Maximised: true}

	tests := []struct {
		name     string
		snapshot *settings.WindowState
		current  settings.WindowState
		ok       bool
		want     settings.WindowState
		wantOK   bool
	}{
		{"senza snapshot usa la geometria corrente", nil, current, true, current, true},
		{"senza snapshot e senza geometria non salva", nil, settings.WindowState{}, false, settings.WindowState{}, false},
		{"lo snapshot vince sulla geometria corrente", snapshot, current, true, *snapshot, true},
		{"lo snapshot basta anche senza geometria", snapshot, settings.WindowState{}, false, *snapshot, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOK := StateToSave(tt.snapshot, tt.current, tt.ok)
			if got != tt.want || gotOK != tt.wantOK {
				t.Errorf("got (%+v, %v), want (%+v, %v)", got, gotOK, tt.want, tt.wantOK)
			}
		})
	}
}
