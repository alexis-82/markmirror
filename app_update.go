package main

import (
	"errors"
	"os"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"markmirror-go/internal/update"
)

// UpdateInfo è il risultato di un controllo aggiornamenti, per il frontend.
type UpdateInfo struct {
	Current    string `json:"current"`
	Latest     string `json:"latest"`
	Available  bool   `json:"available"`  // latest è più recente di current
	CanInstall bool   `json:"canInstall"` // c'è un file da installare su questo sistema
	NoRelease  bool   `json:"noRelease"`  // il repository non ha ancora release
	Dev        bool   `json:"dev"`        // build di sviluppo: nessun aggiornamento proposto
	Notes      string `json:"notes"`
}

// updater contiene lo stato degli aggiornamenti. checker, install, quit ed
// emit sono sostituibili nei test; install avvia l'installer del sistema
// (update_<os>.go) con il file scaricato e il suo SHA-256.
type updater struct {
	mu      sync.Mutex
	checker *update.Checker
	latest  *update.Release
	busy    bool
	install func(path, sum string) error
	quit    func()
	emit    func(event string, data any)
}

func newUpdater(a *App) updater {
	return updater{
		checker: update.NewChecker(version),
		install: update.Install,
		quit:    a.Quit,
		emit:    func(event string, data any) { runtime.EventsEmit(a.ctx, event, data) },
	}
}

// cleanupUpdates rimuove i download rimasti da sessioni precedenti. Le
// cartelle recenti restano: il setup potrebbe essere ancora in esecuzione.
func (a *App) cleanupUpdates() {
	if err := update.CleanupStale(os.TempDir(), update.TempPrefix, 24*time.Hour); err != nil {
		runtime.LogErrorf(a.ctx, "pulizia aggiornamenti: %v", err)
	}
}

// GetVersion restituisce la versione dell'app ("dev" per le build di sviluppo).
func (a *App) GetVersion() string {
	return version
}

// CheckUpdate chiede a GitHub l'ultima release.
func (a *App) CheckUpdate() (UpdateInfo, error) {
	info := UpdateInfo{Current: version}
	r, err := a.updates.checker.Latest(a.ctx)
	if errors.Is(err, update.ErrNoRelease) {
		info.NoRelease = true
		return info, nil
	}
	if err != nil {
		return info, err
	}
	a.updates.mu.Lock()
	a.updates.latest = &r
	a.updates.mu.Unlock()

	info.Latest = r.Version
	info.Notes = r.Notes
	newer, err := update.IsNewer(r.Version, version)
	info.Available = err == nil && newer
	info.Dev = err != nil // la versione corrente non è vX.Y.Z
	_, assetErr := update.PlatformAsset(r.Assets, r.Version)
	info.CanInstall = info.Available && assetErr == nil
	return info, nil
}

// InstallUpdate scarica il file della release per questo sistema, avvia
// l'installazione e chiude l'app. Restituisce solo in caso di errore. Il
// frontend chiede prima di salvare le modifiche.
func (a *App) InstallUpdate() error {
	u := &a.updates
	u.mu.Lock()
	if u.busy {
		u.mu.Unlock()
		return errors.New("aggiornamento già in corso")
	}
	r := u.latest
	u.busy = true
	u.mu.Unlock()
	defer func() {
		u.mu.Lock()
		u.busy = false
		u.mu.Unlock()
	}()

	if r == nil {
		return errors.New("controlla prima gli aggiornamenti")
	}
	if newer, err := update.IsNewer(r.Version, version); err != nil || !newer {
		return errors.New("nessun aggiornamento da installare")
	}
	asset, err := update.PlatformAsset(r.Assets, r.Version)
	if err != nil {
		return err
	}

	path, sum, err := u.checker.Download(a.ctx, asset, func(done, total int64) {
		u.emit("update:progress", map[string]any{"done": done, "total": total})
	})
	if err != nil {
		return err
	}
	if err := u.install(path, sum); err != nil {
		return err
	}
	u.quit()
	return nil
}
