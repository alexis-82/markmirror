package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIsNewer(t *testing.T) {
	cases := []struct {
		latest, current string
		want            bool
	}{
		{"v1.0.1", "v1.0.0", true},
		{"v1.10.0", "v1.9.9", true},
		{"v2.0.0", "v1.99.99", true},
		{"v1.0.0", "v1.0.0", false},
		{"v1.0.0", "v1.0.1", false},
		{"v1.2.0-beta", "v1.1.0", true},
	}
	for _, c := range cases {
		got, err := IsNewer(c.latest, c.current)
		if err != nil || got != c.want {
			t.Errorf("IsNewer(%s, %s) = %v, %v; want %v", c.latest, c.current, got, err, c.want)
		}
	}
	for _, bad := range []string{"dev", "1.0.0", "v1.0", "v1.x.0", ""} {
		if _, err := IsNewer("v1.0.0", bad); err == nil {
			t.Errorf("expected error for current %q", bad)
		}
	}
}

func TestAssetNames(t *testing.T) {
	want := Names{
		WindowsInstaller: "MarkMirror-2.1.0-amd64-installer.exe",
		WindowsPortable:  "MarkMirror-2.1.0-amd64-portable.exe",
		LinuxTarball:     "MarkMirror-2.1.0-linux-amd64.tar.gz",
		Deb:              "markmirror_2.1.0_amd64.deb",
		Rpm:              "markmirror-2.1.0-1.x86_64.rpm",
		Pacman:           "markmirror-2.1.0-1-x86_64.pkg.tar.zst",
	}
	for _, v := range []string{"v2.1.0", "2.1.0"} {
		if got := AssetNames(v, "amd64"); got != want {
			t.Errorf("AssetNames(%q) = %+v, want %+v", v, got, want)
		}
	}
}

// release210 sono i file pubblicati dalla pipeline per v2.1.0, più file di
// altre versioni e la portable, che non devono mai essere scelti.
var release210 = []Asset{
	{Name: "MarkMirror-2.1.0-amd64-portable.exe"},
	{Name: "MarkMirror-2.0.9-amd64-installer.exe"},
	{Name: "markmirror-2.1.0-1.aarch64.rpm"},
	{Name: "MARKMIRROR-2.1.0-AMD64-INSTALLER.EXE"},
	{Name: "MarkMirror-2.1.0-linux-amd64.tar.gz"},
	{Name: "MarkMirror-2.1.0-linux-arm64.tar.gz"},
	{Name: "markmirror_2.1.0_amd64.deb"},
	{Name: "markmirror-2.1.0-1.x86_64.rpm"},
	{Name: "markmirror-2.1.0-1-aarch64.pkg.tar.zst"},
	{Name: "markmirror-2.1.0-1-x86_64.pkg.tar.zst"},
	{Name: "SHA256SUMS.txt"},
}

// release200 sono i file della v2.0.0, pubblicata prima dell'updater.
var release200 = []Asset{
	{Name: "MarkMirror-2.0.0-amd64-installer.exe"},
	{Name: "MarkMirror-2.0.0-amd64_portable.exe"},
	{Name: "MarkMirror-2.0.0-linux"},
}

func TestPickAsset(t *testing.T) {
	if a, err := PickAsset(release210, "v2.1.0", "windows", "amd64"); err != nil || a.Name != "MARKMIRROR-2.1.0-AMD64-INSTALLER.EXE" {
		t.Fatalf("windows: %v %v", a, err)
	}
	if a, err := PickAsset(release210, "v2.1.0", "linux", "arm64"); err != nil || a.Name != "MarkMirror-2.1.0-linux-arm64.tar.gz" {
		t.Fatalf("linux: %v %v", a, err)
	}
	if _, err := PickAsset(release210, "v2.1.0", "darwin", "arm64"); !errors.Is(err, ErrNoAsset) {
		t.Fatalf("darwin: expected ErrNoAsset, got %v", err)
	}
	// Solo portable e setup di un'altra versione: nessun setup da avviare.
	if _, err := PickAsset(release210[:2], "v2.1.0", "windows", "amd64"); !errors.Is(err, ErrNoAsset) {
		t.Fatalf("windows without setup: expected ErrNoAsset, got %v", err)
	}
	// Il setup della v2.0.0 ha già il nome nuovo; il file Linux no.
	if _, err := PickAsset(release200, "v2.0.0", "linux", "amd64"); !errors.Is(err, ErrNoAsset) {
		t.Fatalf("v2.0.0 linux: expected ErrNoAsset, got %v", err)
	}
}

func TestPickDeb(t *testing.T) {
	if a, err := PickDeb(release210, "v2.1.0", "amd64"); err != nil || a.Name != "markmirror_2.1.0_amd64.deb" {
		t.Fatalf("amd64: %v %v", a, err)
	}
	if _, err := PickDeb(release210, "v2.1.0", "arm64"); !errors.Is(err, ErrNoAsset) {
		t.Fatalf("arm64: expected ErrNoAsset, got %v", err)
	}
	if _, err := PickDeb(release210, "v2.2.0", "amd64"); !errors.Is(err, ErrNoAsset) {
		t.Fatalf("other version: expected ErrNoAsset, got %v", err)
	}
	if _, err := PickDeb(release200, "v2.0.0", "amd64"); !errors.Is(err, ErrNoAsset) {
		t.Fatalf("v2.0.0: expected ErrNoAsset, got %v", err)
	}
}

func TestPickRpm(t *testing.T) {
	if a, err := PickRpm(release210, "v2.1.0", "amd64"); err != nil || a.Name != "markmirror-2.1.0-1.x86_64.rpm" {
		t.Fatalf("amd64: %v %v", a, err)
	}
	if a, err := PickRpm(release210, "v2.1.0", "arm64"); err != nil || a.Name != "markmirror-2.1.0-1.aarch64.rpm" {
		t.Fatalf("arm64: %v %v", a, err)
	}
	if _, err := PickRpm(release210, "v2.1.0", "riscv64"); !errors.Is(err, ErrNoAsset) {
		t.Fatalf("unknown arch: expected ErrNoAsset, got %v", err)
	}
	if _, err := PickRpm(release200, "v2.0.0", "amd64"); !errors.Is(err, ErrNoAsset) {
		t.Fatalf("v2.0.0: expected ErrNoAsset, got %v", err)
	}
}

func TestPickPacman(t *testing.T) {
	if a, err := PickPacman(release210, "v2.1.0", "amd64"); err != nil || a.Name != "markmirror-2.1.0-1-x86_64.pkg.tar.zst" {
		t.Fatalf("amd64: %v %v", a, err)
	}
	if a, err := PickPacman(release210, "v2.1.0", "arm64"); err != nil || a.Name != "markmirror-2.1.0-1-aarch64.pkg.tar.zst" {
		t.Fatalf("arm64: %v %v", a, err)
	}
	if _, err := PickPacman(release210, "v2.1.0", "riscv64"); !errors.Is(err, ErrNoAsset) {
		t.Fatalf("unknown arch: expected ErrNoAsset, got %v", err)
	}
	if _, err := PickPacman(release210, "v2.2.0", "amd64"); !errors.Is(err, ErrNoAsset) {
		t.Fatalf("other version: expected ErrNoAsset, got %v", err)
	}
	if _, err := PickPacman(release200, "v2.0.0", "amd64"); !errors.Is(err, ErrNoAsset) {
		t.Fatalf("v2.0.0: expected ErrNoAsset, got %v", err)
	}
}

func checker(srv *httptest.Server) *Checker {
	c := NewChecker("v1.0.0")
	c.Client = srv.Client()
	c.APIURL = srv.URL + "/api/latest"
	c.DownloadPrefix = srv.URL + "/dl/"
	return c
}

func TestLatest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "MarkMirror/v1.0.0" {
			t.Errorf("missing user agent: %q", r.Header.Get("User-Agent"))
		}
		fmt.Fprint(w, `{"tag_name":"v1.2.0","body":"## Novità\n- x","html_url":"https://github.com/r/1",
			"assets":[{"name":"MarkMirror-1.2.0-amd64-installer.exe","size":10,"browser_download_url":"https://h/a.exe","digest":"sha256:ABCD"}]}`)
	}))
	defer srv.Close()

	r, err := checker(srv).Latest(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if r.Version != "v1.2.0" || !strings.Contains(r.Notes, "Novità") || len(r.Assets) != 1 || r.Assets[0].SHA256 != "abcd" {
		t.Fatalf("unexpected release %+v", r)
	}
}

func TestLatestErrors(t *testing.T) {
	status := http.StatusNotFound
	body := `{"message":"Not Found"}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		fmt.Fprint(w, body)
	}))
	defer srv.Close()
	c := checker(srv)

	if _, err := c.Latest(context.Background()); !errors.Is(err, ErrNoRelease) {
		t.Fatalf("404: expected ErrNoRelease, got %v", err)
	}
	status = http.StatusInternalServerError
	if _, err := c.Latest(context.Background()); err == nil || errors.Is(err, ErrNoRelease) {
		t.Fatalf("500: expected generic error, got %v", err)
	}
	status, body = http.StatusOK, `{"tag_name":`
	if _, err := c.Latest(context.Background()); err == nil {
		t.Fatal("broken JSON: expected error")
	}
	body = `{"tag_name":"release-1"}`
	if _, err := c.Latest(context.Background()); err == nil {
		t.Fatal("bad tag: expected error")
	}

	srv.Close() // offline
	if _, err := c.Latest(context.Background()); err == nil {
		t.Fatal("offline: expected error")
	}
}

// isolateTempDir fa usare a os.TempDir una cartella solo di questo test:
// altri package di test, in parallelo, creano anch'essi cartelle
// markmirror-update-* nella cartella temporanea di sistema.
func isolateTempDir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	for _, env := range []string{"TMP", "TEMP", "TMPDIR"} {
		t.Setenv(env, dir)
	}
	if os.TempDir() != dir {
		t.Fatalf("os.TempDir() = %s, want %s", os.TempDir(), dir)
	}
}

func TestDownload(t *testing.T) {
	isolateTempDir(t)
	payload := []byte("setup-binary-content")
	sum := sha256.Sum256(payload)
	served := payload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(served)
	}))
	defer srv.Close()
	c := checker(srv)
	name := "MarkMirror-1.2.0-amd64-installer.exe"
	asset := Asset{Name: name, URL: srv.URL + "/dl/v1.2.0/" + name, Size: int64(len(payload)), SHA256: hex.EncodeToString(sum[:])}

	var last int64
	path, gotSum, err := c.Download(context.Background(), asset, func(done, total int64) { last = done })
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(filepath.Dir(path))
	got, _ := os.ReadFile(path)
	if string(got) != string(payload) || filepath.Base(path) != asset.Name || last != asset.Size {
		t.Fatalf("download mismatch: %q %s progress=%d", got, path, last)
	}
	if gotSum != asset.SHA256 {
		t.Fatalf("sum = %s, want %s", gotSum, asset.SHA256)
	}
	if !strings.HasPrefix(filepath.Base(filepath.Dir(path)), TempPrefix) {
		t.Fatalf("temp dir %s without prefix %s", filepath.Dir(path), TempPrefix)
	}

	// Senza digest di GitHub: l'hash viene comunque calcolato.
	noDigest := asset
	noDigest.SHA256 = ""
	path2, gotSum, err := c.Download(context.Background(), noDigest, nil)
	if err != nil || gotSum != asset.SHA256 {
		t.Fatalf("no digest: sum=%s err=%v", gotSum, err)
	}
	os.RemoveAll(filepath.Dir(path2))

	// Ogni errore deve rimuovere la cartella temporanea.
	before := tempDirs(t)
	assertNoLeftovers := func(what string) {
		t.Helper()
		if after := tempDirs(t); after != before {
			t.Fatalf("%s: temp dirs %d -> %d", what, before, after)
		}
	}

	bad := asset
	bad.SHA256 = strings.Repeat("0", 64)
	if _, _, err := c.Download(context.Background(), bad, nil); err == nil || !strings.Contains(err.Error(), "SHA-256") {
		t.Fatalf("expected checksum error, got %v", err)
	}
	assertNoLeftovers("checksum")

	// Corpo troncato o più lungo.
	for _, b := range [][]byte{payload[:5], append(append([]byte{}, payload...), 'x')} {
		served = b
		if _, _, err := c.Download(context.Background(), asset, nil); err == nil {
			t.Fatalf("expected size error for %d bytes", len(b))
		}
		assertNoLeftovers("size")
	}
	served = payload

	evil := asset
	evil.URL = srv.URL + "/other/" + name
	if _, _, err := c.Download(context.Background(), evil, nil); err == nil || !strings.Contains(err.Error(), "non consentito") {
		t.Fatalf("expected prefix error, got %v", err)
	}

	for _, n := range []string{"../evil.exe", `..\evil.exe`, "", ".."} {
		trav := asset
		trav.Name = n
		if _, _, err := c.Download(context.Background(), trav, nil); err == nil {
			t.Fatalf("expected invalid name error for %q", n)
		}
	}
	assertNoLeftovers("names")
}

func tempDirs(t *testing.T) int {
	t.Helper()
	entries, err := os.ReadDir(os.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), TempPrefix) {
			n++
		}
	}
	return n
}

func TestCleanupStale(t *testing.T) {
	dir := t.TempDir()
	old := time.Now().Add(-48 * time.Hour)
	mk := func(name string, dirEntry bool, mod time.Time) string {
		p := filepath.Join(dir, name)
		var err error
		if dirEntry {
			err = os.Mkdir(p, 0o700)
			if err == nil {
				err = os.WriteFile(filepath.Join(p, "setup.exe"), []byte("x"), 0o600)
			}
		} else {
			err = os.WriteFile(p, []byte("x"), 0o600)
		}
		if err == nil {
			err = os.Chtimes(p, mod, mod)
		}
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	stale := mk(TempPrefix+"111", true, old)
	recent := mk(TempPrefix+"222", true, time.Now())
	file := mk(TempPrefix+"file", false, old)
	other := mk("other-update-333", true, old)

	if err := CleanupStale(dir, TempPrefix, 24*time.Hour); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Errorf("stale dir not removed: %v", err)
	}
	for _, p := range []string{recent, file, other} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("%s removed: %v", p, err)
		}
	}
}

func TestDefaultPrefixMatchesRepo(t *testing.T) {
	if !strings.HasPrefix("https://github.com/alexis-82/markmirror/releases/download/v2.1.0/MarkMirror-2.1.0-amd64-installer.exe", DefaultDownloadPrefix) {
		t.Fatal("default prefix does not match the repo download URLs")
	}
}

// Controllo facoltativo contro la vera API di GitHub:
// MARKMIRROR_LIVE_GITHUB=1 go test ./internal/update -run Live
func TestLiveGitHub(t *testing.T) {
	if os.Getenv("MARKMIRROR_LIVE_GITHUB") != "1" {
		t.Skip("set MARKMIRROR_LIVE_GITHUB=1 to query api.github.com")
	}
	r, err := NewChecker("v0.0.0").Latest(context.Background())
	if errors.Is(err, ErrNoRelease) {
		t.Log("no release published yet")
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("latest %s, assets: %d", r.Version, len(r.Assets))
	for _, a := range r.Assets {
		if !strings.HasPrefix(a.URL, DefaultDownloadPrefix) {
			t.Errorf("asset URL outside the allowed prefix: %s", a.URL)
		}
	}
}
