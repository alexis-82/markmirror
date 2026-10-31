// Package update controlla su GitHub Releases se esiste una versione più
// recente di MarkMirror e scarica il file della release per la piattaforma
// corrente. Non sa nulla di Wails né di come il file viene installato.
package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	Repo                  = "alexis-82/markmirror"
	DefaultAPIURL         = "https://api.github.com/repos/" + Repo + "/releases/latest"
	DefaultDownloadPrefix = "https://github.com/" + Repo + "/releases/download/"
	// TempPrefix è il prefisso delle cartelle temporanee dei download.
	TempPrefix = "markmirror-update-"
)

var (
	// ErrNoRelease: il repository non ha ancora una release pubblicata.
	ErrNoRelease = errors.New("nessuna release pubblicata")
	// ErrNoAsset: la release non contiene un file per questo sistema.
	ErrNoAsset = errors.New("la release non contiene un file per questo sistema")
)

type Asset struct {
	Name   string
	URL    string
	Size   int64
	SHA256 string // esadecimale, vuoto se GitHub non ha fornito il digest
}

type Release struct {
	Version string // tag, es. v1.2.3
	Notes   string // testo della release (Markdown)
	PageURL string
	Assets  []Asset
}

type Checker struct {
	Client *http.Client
	APIURL string
	// DownloadPrefix: i file si scaricano solo da URL che iniziano così.
	DownloadPrefix string
	UserAgent      string
}

func NewChecker(currentVersion string) *Checker {
	return &Checker{
		// Nessun timeout globale: interromperebbe i download lunghi; le
		// richieste usano il context.
		Client:         &http.Client{},
		APIURL:         DefaultAPIURL,
		DownloadPrefix: DefaultDownloadPrefix,
		UserAgent:      "MarkMirror/" + currentVersion,
	}
}

type ghRelease struct {
	TagName string `json:"tag_name"`
	Body    string `json:"body"`
	HTMLURL string `json:"html_url"`
	Assets  []struct {
		Name               string `json:"name"`
		Size               int64  `json:"size"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Digest             string `json:"digest"` // "sha256:<hex>" sui caricamenti recenti
	} `json:"assets"`
}

// Latest restituisce l'ultima release pubblicata (non bozza, non prerelease).
func (c *Checker) Latest(ctx context.Context) (Release, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.APIURL, nil)
	if err != nil {
		return Release{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", c.UserAgent)
	resp, err := c.Client.Do(req)
	if err != nil {
		return Release{}, fmt.Errorf("impossibile contattare GitHub: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return Release{}, ErrNoRelease
	}
	if resp.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("risposta inattesa da GitHub: %s", resp.Status)
	}
	var gr ghRelease
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&gr); err != nil {
		return Release{}, fmt.Errorf("risposta di GitHub non valida: %w", err)
	}
	if _, err := parseVersion(gr.TagName); err != nil {
		return Release{}, fmt.Errorf("tag dell'ultima release non valido %q (atteso vX.Y.Z)", gr.TagName)
	}
	r := Release{Version: gr.TagName, Notes: gr.Body, PageURL: gr.HTMLURL}
	for _, a := range gr.Assets {
		sum := ""
		if d, ok := strings.CutPrefix(a.Digest, "sha256:"); ok {
			sum = strings.ToLower(d)
		}
		r.Assets = append(r.Assets, Asset{Name: a.Name, URL: a.BrowserDownloadURL, Size: a.Size, SHA256: sum})
	}
	return r, nil
}

// parseVersion interpreta "vX.Y.Z" (un eventuale "-suffisso" è ignorato).
func parseVersion(v string) ([3]int, error) {
	var out [3]int
	s, ok := strings.CutPrefix(strings.TrimSpace(v), "v")
	if !ok {
		return out, fmt.Errorf("versione senza prefisso v: %q", v)
	}
	s, _, _ = strings.Cut(s, "-")
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return out, fmt.Errorf("versione non valida: %q", v)
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return out, fmt.Errorf("versione non valida: %q", v)
		}
		out[i] = n
	}
	return out, nil
}

// IsNewer indica se latest è una versione più alta di current.
func IsNewer(latest, current string) (bool, error) {
	l, err := parseVersion(latest)
	if err != nil {
		return false, err
	}
	c, err := parseVersion(current)
	if err != nil {
		return false, err
	}
	for i := range l {
		if l[i] != c[i] {
			return l[i] > c[i], nil
		}
	}
	return false, nil
}

// Names sono i nomi dei file di una release. La pipeline di release
// (.github/workflows/release.yml, build/linux/package.sh, nfpm.yaml) deve
// produrre esattamente questi nomi.
type Names struct {
	WindowsInstaller string
	WindowsPortable  string
	LinuxTarball     string
	Deb              string
	Rpm              string
	Pacman           string
}

// rpmArch converte le architetture Go nei nomi usati da rpm (e da nfpm).
var rpmArch = map[string]string{"amd64": "x86_64", "arm64": "aarch64"}

// AssetNames restituisce i nomi dei file della release version ("vX.Y.Z" o
// "X.Y.Z") per l'architettura goarch.
func AssetNames(version, goarch string) Names {
	v := strings.TrimPrefix(version, "v")
	return Names{
		WindowsInstaller: "MarkMirror-" + v + "-" + goarch + "-installer.exe",
		WindowsPortable:  "MarkMirror-" + v + "-" + goarch + "-portable.exe",
		LinuxTarball:     "MarkMirror-" + v + "-linux-" + goarch + ".tar.gz",
		Deb:              "markmirror_" + v + "_" + goarch + ".deb",
		Rpm:              "markmirror-" + v + "-1." + rpmArch[goarch] + ".rpm",
		Pacman:           "markmirror-" + v + "-1-" + rpmArch[goarch] + ".pkg.tar.zst",
	}
}

func pick(assets []Asset, name string) (Asset, error) {
	for _, a := range assets {
		if strings.EqualFold(a.Name, name) {
			return a, nil
		}
	}
	return Asset{}, ErrNoAsset
}

// PickAsset restituisce il file da installare su goos/goarch: il setup su
// Windows, il .tar.gz su Linux.
func PickAsset(assets []Asset, version, goos, goarch string) (Asset, error) {
	names := AssetNames(version, goarch)
	switch goos {
	case "windows":
		return pick(assets, names.WindowsInstaller)
	case "linux":
		return pick(assets, names.LinuxTarball)
	}
	return Asset{}, ErrNoAsset
}

// PickDeb restituisce il pacchetto Debian per goarch.
func PickDeb(assets []Asset, version, goarch string) (Asset, error) {
	return pick(assets, AssetNames(version, goarch).Deb)
}

// PickRpm restituisce il pacchetto Fedora per goarch.
func PickRpm(assets []Asset, version, goarch string) (Asset, error) {
	if _, ok := rpmArch[goarch]; !ok {
		return Asset{}, ErrNoAsset
	}
	return pick(assets, AssetNames(version, goarch).Rpm)
}

// PickPacman restituisce il pacchetto Arch Linux per goarch (stessi nomi di
// architettura di rpm).
func PickPacman(assets []Asset, version, goarch string) (Asset, error) {
	if _, ok := rpmArch[goarch]; !ok {
		return Asset{}, ErrNoAsset
	}
	return pick(assets, AssetNames(version, goarch).Pacman)
}

// Download salva il file in una nuova cartella temporanea privata e ne
// verifica la dimensione e, se noto, lo SHA-256. Restituisce il percorso e lo
// SHA-256 calcolato. progress può essere nil.
func (c *Checker) Download(ctx context.Context, a Asset, progress func(done, total int64)) (path, sum string, err error) {
	if !strings.HasPrefix(a.URL, c.DownloadPrefix) {
		return "", "", fmt.Errorf("indirizzo di download non consentito: %s", a.URL)
	}
	if a.Name == "" || a.Name == "." || a.Name == ".." || strings.ContainsAny(a.Name, `/\`) {
		return "", "", fmt.Errorf("nome file non valido: %q", a.Name)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.URL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", c.UserAgent)
	resp, err := c.Client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("download non riuscito: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("download non riuscito: %s", resp.Status)
	}

	dir, err := os.MkdirTemp("", TempPrefix)
	if err != nil {
		return "", "", err
	}
	path = filepath.Join(dir, a.Name)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o700)
	if err != nil {
		os.RemoveAll(dir)
		return "", "", err
	}
	fail := func(err error) (string, string, error) {
		f.Close()
		os.RemoveAll(dir)
		return "", "", err
	}

	h := sha256.New()
	var done int64
	buf := make([]byte, 64*1024)
	// Mai leggere più della dimensione annunciata (+1 per accorgersi di un
	// corpo più lungo).
	body := io.LimitReader(resp.Body, a.Size+1)
	for {
		n, rerr := body.Read(buf)
		if n > 0 {
			if _, err := f.Write(buf[:n]); err != nil {
				return fail(err)
			}
			h.Write(buf[:n])
			done += int64(n)
			if progress != nil {
				progress(done, a.Size)
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return fail(fmt.Errorf("download interrotto: %w", rerr))
		}
	}
	if done != a.Size {
		return fail(fmt.Errorf("file scaricato incompleto o diverso (%d byte invece di %d)", done, a.Size))
	}
	sum = hex.EncodeToString(h.Sum(nil))
	if a.SHA256 != "" && sum != a.SHA256 {
		return fail(errors.New("il file scaricato non corrisponde al checksum SHA-256 della release"))
	}
	if err := f.Close(); err != nil {
		os.RemoveAll(dir)
		return "", "", err
	}
	return path, sum, nil
}

// CleanupStale rimuove le cartelle di dir con il prefisso indicato più
// vecchie di olderThan (download interrotti, setup già eseguiti).
func CleanupStale(dir, prefix string, olderThan time.Duration) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	limit := time.Now().Add(-olderThan)
	var errs []error
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), prefix) {
			continue
		}
		info, err := e.Info()
		if err != nil || !info.ModTime().Before(limit) {
			continue
		}
		if err := os.RemoveAll(filepath.Join(dir, e.Name())); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
