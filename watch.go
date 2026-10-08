package main

import (
	"os"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// watchInterval è l'intervallo di controllo delle modifiche esterne al file.
const watchInterval = time.Second

// fileStamp identifica una versione del file su disco.
type fileStamp struct {
	modTime time.Time
	size    int64
}

func (s fileStamp) equal(o fileStamp) bool {
	return s.size == o.size && s.modTime.Equal(o.modTime)
}

func statStamp(path string) (fileStamp, bool) {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return fileStamp{}, false
	}
	return fileStamp{modTime: info.ModTime(), size: info.Size()}, true
}

// watchFile controlla periodicamente il documento corrente e avvisa il
// frontend ("file:changed", con il percorso) quando cambia su disco.
func (a *App) watchFile() {
	ticker := time.NewTicker(watchInterval)
	defer ticker.Stop()
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			if path, changed := a.checkExternalChange(); changed {
				runtime.EventsEmit(a.ctx, "file:changed", path)
			}
		}
	}
}

// checkExternalChange segnala se il documento corrente è cambiato su disco
// dall'ultimo controllo. Ogni versione viene segnalata una sola volta; un file
// rimosso o illeggibile non viene segnalato.
func (a *App) checkExternalChange() (string, bool) {
	// Salta il controllo mentre l'applicazione sta leggendo o scrivendo il file.
	if !a.ioMu.TryLock() {
		return "", false
	}
	defer a.ioMu.Unlock()

	a.mu.Lock()
	path, last := a.currentFile, a.stamp
	a.mu.Unlock()
	if path == "" {
		return "", false
	}
	stamp, ok := statStamp(path)
	if !ok || stamp.equal(last) {
		return "", false
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	if a.currentFile != path {
		return "", false
	}
	a.stamp = stamp
	return path, true
}
