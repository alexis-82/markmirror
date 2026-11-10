//go:build linux

package update

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"syscall"
)

// MarkMirror si installa su Linux in quattro modi:
//   - pacchetto .deb (nfpm): /usr/bin/markmirror, aggiornato con apt;
//   - pacchetto .rpm (nfpm, Fedora): /usr/bin/markmirror, aggiornato con dnf;
//   - pacchetto .pkg.tar.zst (nfpm, Arch Linux): /usr/bin/markmirror, aggiornato con pacman;
//   - build/linux/install.sh: /usr/local/bin/markmirror.
//
// In tutti i casi l'installazione come root passa da markmirror-update.sh,
// che sta accanto al binario e riverifica lo SHA-256 del file su una copia
// di root prima di installarlo.
const (
	dpkgFileList = "/var/lib/dpkg/info/markmirror.list"
	pkgBinary    = "/usr/bin/markmirror"
	updateScript = "markmirror-update.sh"
	launcher     = "markmirror-launch"
)

// installedFromDeb indica se il binario in esecuzione appartiene al
// pacchetto Debian "markmirror".
func installedFromDeb() bool {
	if _, err := os.Stat(dpkgFileList); err != nil {
		return false
	}
	return runningExecutable() == pkgBinary
}

// installedFromRpm indica se il binario in esecuzione appartiene al
// pacchetto rpm "markmirror", chiedendolo al database rpm.
func installedFromRpm() bool {
	return ownedByPackage(runningExecutable(), func(path string) (string, error) {
		out, err := exec.Command("rpm", "-qf", "--qf", "%{NAME}", path).Output()
		return string(out), err
	})
}

// installedFromPacman indica se il binario in esecuzione appartiene al
// pacchetto Arch Linux "markmirror", chiedendolo al database di pacman.
func installedFromPacman() bool {
	return ownedByPackage(runningExecutable(), func(path string) (string, error) {
		out, err := exec.Command("pacman", "-Qqo", path).Output()
		return string(out), err
	})
}

// ownedByPackage indica se exe è il binario dei pacchetti e la query al
// gestore dei pacchetti (passata da fuori, per i test) dice che appartiene a
// "markmirror". Senza gestore, o con un file di nessun pacchetto (rpm e
// pacman escono con 1): false.
func ownedByPackage(exe string, query func(path string) (string, error)) bool {
	if exe != pkgBinary {
		return false
	}
	name, err := query(exe)
	return err == nil && strings.TrimSpace(name) == "markmirror"
}

func runningExecutable() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		return resolved
	}
	return exe
}

// PlatformAsset sceglie il file della release in base a come è installato
// MarkMirror.
func PlatformAsset(assets []Asset, version string) (Asset, error) {
	if installedFromDeb() {
		return PickDeb(assets, version, goruntime.GOARCH)
	}
	if installedFromRpm() {
		return PickRpm(assets, version, goruntime.GOARCH)
	}
	if installedFromPacman() {
		return PickPacman(assets, version, goruntime.GOARCH)
	}
	return PickAsset(assets, version, "linux", goruntime.GOARCH)
}

// installCommand restituisce il comando che installa path come root:
// pkexec <cartella del binario>/markmirror-update.sh <path> <sha256>.
func installCommand(pkexec, exe, path, sum string) []string {
	return []string{pkexec, filepath.Join(filepath.Dir(exe), updateScript), path, sum}
}

// Install installa il file scaricato tramite pkexec (richiesta grafica
// della password) e avvia la nuova versione; l'app si chiude subito dopo.
func Install(path, sum string) error {
	pkexec, err := exec.LookPath("pkexec")
	if err != nil {
		return errors.New("pkexec non disponibile: aggiorna a mano da terminale con: sudo " + updateScript)
	}
	exe := runningExecutable()
	args := installCommand(pkexec, exe, path, sum)
	if _, err := os.Stat(args[1]); err != nil {
		return fmt.Errorf("%s non trovato: aggiorna a mano scaricando la nuova versione da GitHub", args[1])
	}

	out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() == 126 {
			return errors.New("autorizzazione annullata: aggiornamento non installato")
		}
		return fmt.Errorf("installazione non riuscita: %v\n%s", err, strings.TrimSpace(string(out)))
	}
	os.RemoveAll(filepath.Dir(path))

	// Il launcher sta accanto al binario in tutti i modi di installazione.
	// Parte in una sessione propria, così sopravvive alla nostra uscita.
	relaunch := exec.Command(filepath.Join(filepath.Dir(exe), launcher))
	relaunch.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := relaunch.Start(); err != nil {
		return fmt.Errorf("aggiornamento installato, ma impossibile riavviare MarkMirror: %w", err)
	}
	return nil
}
