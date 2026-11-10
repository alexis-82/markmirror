//go:build !windows

package main

import "markmirror-go/internal/settings"

func refreshWindowFrame() {}

func applyMenuTheme(string) {}

func windowPlacement() (settings.WindowState, bool) { return settings.WindowState{}, false }

func restoreWindowPlacement(settings.WindowState) {}
