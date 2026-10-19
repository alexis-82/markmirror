package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
	"testing"

	"markmirror-go/internal/update"
)

// fakeGitHub serve una "ultima release" con il file per il sistema corrente.
type fakeGitHub struct {
	srv     *httptest.Server
	tag     string
	payload []byte
	status  int
	noAsset bool
}

func newFakeGitHub(t *testing.T, tag string) *fakeGitHub {
	f := &fakeGitHub{tag: tag, payload: []byte("installer"), status: http.StatusOK}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/dl/") {
			w.Write(f.payload)
			return
		}
		if f.status != http.StatusOK {
			w.WriteHeader(f.status)
			return
		}
		names := update.AssetNames(f.tag, goruntime.GOARCH)
		name := names.WindowsInstaller
		if goruntime.GOOS == "linux" {
			name = names.LinuxTarball
		}
		if f.noAsset {
			name = names.WindowsPortable
		}
		sum := sha256.Sum256(f.payload)
		fmt.Fprintf(w, `{"tag_name":%q,"body":"Note","html_url":"x","assets":[{"name":%q,"size":%d,"browser_download_url":%q,"digest":"sha256:%s"}]}`,
			f.tag, name, len(f.payload), f.srv.URL+"/dl/"+name, hex.EncodeToString(sum[:]))
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func withVersion(t *testing.T, v string) {
	old := version
	version = v
	t.Cleanup(func() { version = old })
}

// installRecord registra gli avvii dell'installer e le chiusure invece di
// eseguirli.
type installRecord struct {
	mu        sync.Mutex
	installed []string // contenuto dei file passati all'installer
	sums      []string
	quits     int
	events    int
}

func updatableApp(t *testing.T, gh *fakeGitHub) (*App, *installRecord) {
	t.Helper()
	if goruntime.GOOS != "windows" && goruntime.GOOS != "linux" {
		t.Skip("nessun file installabile su questo sistema")
	}
	// I download restano nella cartella del test, non in quella di sistema.
	tmp := t.TempDir()
	for _, env := range []string{"TMP", "TEMP", "TMPDIR"} {
		t.Setenv(env, tmp)
	}
	a := &App{ctx: context.Background(), settings: defaultSettings()}
	c := update.NewChecker(version)
	c.Client = gh.srv.Client()
	c.APIURL = gh.srv.URL + "/api"
	c.DownloadPrefix = gh.srv.URL + "/dl/"
	rec := &installRecord{}
	a.updates = updater{
		checker: c,
		install: func(p, sum string) error {
			data, _ := os.ReadFile(p)
			rec.mu.Lock()
			rec.installed = append(rec.installed, string(data))
			rec.sums = append(rec.sums, sum)
			rec.mu.Unlock()
			os.RemoveAll(filepath.Dir(p))
			return nil
		},
		quit: func() { rec.mu.Lock(); rec.quits++; rec.mu.Unlock() },
		emit: func(event string, _ any) {
			if event == "update:progress" {
				rec.mu.Lock()
				rec.events++
				rec.mu.Unlock()
			}
		},
	}
	return a, rec
}

func TestUpdate_CheckStates(t *testing.T) {
	withVersion(t, "v1.0.0")
	gh := newFakeGitHub(t, "v1.1.0")
	a, _ := updatableApp(t, gh)

	info, err := a.CheckUpdate()
	if err != nil || !info.Available || !info.CanInstall || info.Dev || info.Latest != "v1.1.0" || info.Current != "v1.0.0" || info.Notes != "Note" {
		t.Fatalf("newer release: %+v %v", info, err)
	}

	gh.tag = "v1.0.0"
	if info, _ := a.CheckUpdate(); info.Available || info.CanInstall || info.Dev {
		t.Fatalf("same version must not be offered: %+v", info)
	}

	gh.tag, gh.noAsset = "v1.2.0", true
	if info, _ := a.CheckUpdate(); !info.Available || info.CanInstall {
		t.Fatalf("release without installable file: %+v", info)
	}
	gh.noAsset = false

	withVersion(t, "dev")
	gh.tag = "v9.9.9"
	if info, _ := a.CheckUpdate(); info.Available || !info.Dev || info.Latest != "v9.9.9" {
		t.Fatalf("dev build must not be offered updates: %+v", info)
	}

	gh.status = http.StatusNotFound
	if info, err := a.CheckUpdate(); err != nil || !info.NoRelease {
		t.Fatalf("no release yet: %+v %v", info, err)
	}

	gh.srv.Close()
	if _, err := a.CheckUpdate(); err == nil {
		t.Fatal("offline: expected an error")
	}
}

func TestUpdate_InstallLaunchesAndQuits(t *testing.T) {
	withVersion(t, "v1.0.0")
	gh := newFakeGitHub(t, "v1.1.0")
	a, rec := updatableApp(t, gh)

	if err := a.InstallUpdate(); err == nil {
		t.Fatal("install before a check must fail")
	}

	a.CheckUpdate()
	if err := a.InstallUpdate(); err != nil {
		t.Fatal(err)
	}
	want := sha256.Sum256([]byte("installer"))
	if len(rec.installed) != 1 || rec.installed[0] != "installer" || rec.sums[0] != hex.EncodeToString(want[:]) || rec.quits != 1 {
		t.Fatalf("installer not launched / app not quit: %+v", rec)
	}
	if rec.events == 0 {
		t.Fatal("no download progress events")
	}
}

func TestUpdate_NothingToInstall(t *testing.T) {
	withVersion(t, "v1.1.0")
	gh := newFakeGitHub(t, "v1.1.0")
	a, rec := updatableApp(t, gh)
	a.CheckUpdate()
	if err := a.InstallUpdate(); err == nil {
		t.Fatal("same version: expected error")
	}
	if len(rec.installed) != 0 || rec.quits != 0 {
		t.Fatal("installer launched without an update")
	}
}

func TestUpdate_CorruptDownloadNotInstalled(t *testing.T) {
	withVersion(t, "v1.0.0")
	gh := newFakeGitHub(t, "v1.1.0")
	a, rec := updatableApp(t, gh)
	a.CheckUpdate() // digest calcolato su "installer"
	gh.payload = []byte("tampered!")
	if err := a.InstallUpdate(); err == nil {
		t.Fatal("expected verification error")
	}
	if len(rec.installed) != 0 || rec.quits != 0 {
		t.Fatal("corrupt installer launched")
	}
}

func TestUpdate_InstallErrorKeepsAppOpen(t *testing.T) {
	withVersion(t, "v1.0.0")
	gh := newFakeGitHub(t, "v1.1.0")
	a, rec := updatableApp(t, gh)
	a.CheckUpdate()
	a.updates.install = func(p, _ string) error {
		os.RemoveAll(filepath.Dir(p))
		return fmt.Errorf("autorizzazione annullata")
	}
	if err := a.InstallUpdate(); err == nil || !strings.Contains(err.Error(), "annullata") {
		t.Fatalf("expected installer error, got %v", err)
	}
	if rec.quits != 0 {
		t.Fatal("app quit although the installer failed")
	}
	// Dopo un errore si può riprovare.
	a.updates.install = func(p, _ string) error { os.RemoveAll(filepath.Dir(p)); return nil }
	if err := a.InstallUpdate(); err != nil {
		t.Fatalf("retry: %v", err)
	}
}

func TestUpdate_ConcurrentInstallRejected(t *testing.T) {
	withVersion(t, "v1.0.0")
	gh := newFakeGitHub(t, "v1.1.0")
	a, _ := updatableApp(t, gh)
	a.CheckUpdate()

	// La prima installazione resta bloccata dentro l'installer.
	release := make(chan struct{})
	entered := make(chan struct{})
	a.updates.install = func(p, _ string) error {
		close(entered)
		<-release
		os.RemoveAll(filepath.Dir(p))
		return nil
	}
	done := make(chan error)
	go func() { done <- a.InstallUpdate() }()
	<-entered
	if err := a.InstallUpdate(); err == nil || !strings.Contains(err.Error(), "in corso") {
		t.Fatalf("second install while busy: expected rejection, got %v", err)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}
