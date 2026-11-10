//go:build windows

package window

// SystemDark dice se il sistema usa il tema scuro per le app.
func SystemDark() bool { return systemUsesDarkTheme() }
