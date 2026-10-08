//go:build !windows

package main

func refreshWindowFrame() {}

func applyMenuTheme(string) {}

func windowPlacement() (WindowState, bool) { return WindowState{}, false }

func restoreWindowPlacement(WindowState) {}
