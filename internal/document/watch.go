package document

import (
	"context"
	"os"
	"time"
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

// Watch controlla periodicamente il documento corrente e chiama onChange
// (con il percorso) quando cambia su disco, finché ctx non termina.
func (d *Document) Watch(ctx context.Context, onChange func(path string)) {
	ticker := time.NewTicker(watchInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if path, changed := d.CheckExternalChange(); changed {
				onChange(path)
			}
		}
	}
}

// CheckExternalChange segnala se il documento corrente è cambiato su disco
// dall'ultimo controllo. Ogni versione viene segnalata una sola volta; un file
// rimosso o illeggibile non viene segnalato.
func (d *Document) CheckExternalChange() (string, bool) {
	// Salta il controllo mentre l'applicazione sta leggendo o scrivendo il file.
	if !d.ioMu.TryLock() {
		return "", false
	}
	defer d.ioMu.Unlock()

	d.mu.Lock()
	path, last := d.path, d.stamp
	d.mu.Unlock()
	if path == "" {
		return "", false
	}
	stamp, ok := statStamp(path)
	if !ok || stamp.equal(last) {
		return "", false
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	if d.path != path {
		return "", false
	}
	d.stamp = stamp
	return path, true
}
