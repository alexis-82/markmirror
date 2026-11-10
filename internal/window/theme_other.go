//go:build !windows && !linux

package window

// SystemDark: su questi sistemi il tema non è noto prima che la pagina sia
// caricata, quindi vale "chiaro".
func SystemDark() bool { return false }
