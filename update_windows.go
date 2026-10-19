//go:build windows

package main

import (
	"fmt"
	"path/filepath"
	goruntime "runtime"

	"golang.org/x/sys/windows"

	"markmirror-go/internal/update"
)

// installUpdate avvia il setup scaricato; l'app si chiude subito dopo.
// ShellExecute (a differenza di exec.Command) gestisce anche un setup che
// chiede l'elevazione. L'hash non serve: il file non passa a un altro utente.
func installUpdate(path, _ string) error {
	err := windows.ShellExecute(0, windows.StringToUTF16Ptr("open"), windows.StringToUTF16Ptr(path),
		nil, windows.StringToUTF16Ptr(filepath.Dir(path)), windows.SW_SHOWNORMAL)
	if err != nil {
		return fmt.Errorf("impossibile avviare il setup: %w", err)
	}
	return nil
}

// platformAsset sceglie il setup: anche la versione portable si aggiorna così.
func platformAsset(assets []update.Asset, version string) (update.Asset, error) {
	return update.PickAsset(assets, version, "windows", goruntime.GOARCH)
}
