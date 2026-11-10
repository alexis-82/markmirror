//go:build !windows

package window

import "markmirror-go/internal/settings"

func RefreshFrame() {}

func ApplyMenuTheme(string) {}

func Placement() (settings.WindowState, bool) { return settings.WindowState{}, false }

func RestorePlacement(settings.WindowState) {}
