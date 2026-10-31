package main

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"markmirror-go/internal/update"
)

// I file della release sono prodotti da script diversi (workflow, package.sh,
// nfpm.yaml) ma l'updater li cerca con update.AssetNames: se un nome cambia
// da una parte sola, l'aggiornamento non trova più il file.

func readRepoFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return strings.ReplaceAll(string(data), "\r\n", "\n")
}

func TestReleaseWorkflowExpectsAssetNames(t *testing.T) {
	wf := readRepoFile(t, ".github/workflows/release.yml")
	// Elenco controllato dal passo "Nomi dei file", con $V = versione.
	expanded := strings.NewReplacer("${V}", "2.1.0", "$V", "2.1.0").Replace(wf)
	n := update.AssetNames("v2.1.0", "amd64")
	for _, name := range []string{n.WindowsInstaller, n.WindowsPortable, n.LinuxTarball, n.Deb, n.Rpm, n.Pacman} {
		if !strings.Contains(expanded, `"`+name+`"`) {
			t.Errorf("release.yml non controlla %s", name)
		}
	}
	// La portable è l'exe di wails build, copiato con il nome della release.
	if !strings.Contains(wf, `"build\bin\MarkMirror-$ver-amd64-portable.exe"`) {
		t.Error("release.yml: manca la copia della portable")
	}
	if !strings.Contains(wf, `/DVERSION=$ver build\windows\installer\markmirror.nsi`) {
		t.Error("release.yml: il setup non è costruito con markmirror.nsi e la versione del tag")
	}
	if !strings.Contains(wf, "if-no-files-found: error") {
		t.Error("release.yml: gli artifact mancanti non fanno fallire la build")
	}
	if !strings.Contains(wf, `grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$'`) {
		t.Error("release.yml: manca la validazione del formato del tag")
	}
}

func TestInstallerScriptName(t *testing.T) {
	nsi := readRepoFile(t, "build/windows/installer/markmirror.nsi")
	m := regexp.MustCompile(`(?m)^OutFile "\.\.\\\.\.\\bin\\(.+)"$`).FindStringSubmatch(nsi)
	if m == nil {
		t.Fatal("markmirror.nsi: OutFile non trovato")
	}
	got := strings.NewReplacer("${APP_NAME}", "MarkMirror", "${VERSION}", "2.1.0").Replace(m[1])
	if want := update.AssetNames("v2.1.0", "amd64").WindowsInstaller; got != want {
		t.Errorf("markmirror.nsi produce %s, l'updater cerca %s", got, want)
	}
	// L'updater avvia il setup senza elevazione: deve restare per utente.
	if !strings.Contains(nsi, "RequestExecutionLevel user") {
		t.Error("markmirror.nsi: il setup non è più per utente")
	}
}

func TestPackageScriptTarballName(t *testing.T) {
	sh := readRepoFile(t, "build/linux/package.sh")
	m := regexp.MustCompile(`(?m)^OUT="build/bin/(.+)"$`).FindStringSubmatch(sh)
	if m == nil {
		t.Fatal("package.sh: riga OUT= non trovata")
	}
	got := strings.NewReplacer("$VERSION", "2.1.0", "$ARCH", "amd64").Replace(m[1])
	if want := update.AssetNames("v2.1.0", "amd64").LinuxTarball; got != want {
		t.Errorf("package.sh produce %s, l'updater cerca %s", got, want)
	}
}

func TestNfpmPackageName(t *testing.T) {
	cfg := readRepoFile(t, "build/linux/nfpm.yaml")
	// nfpm nomina i pacchetti <name>_<version>_<arch>.deb,
	// <name>-<version>-1.<arch>.rpm e <name>-<version>-1-<arch>.pkg.tar.zst:
	// basta che il nome sia "markmirror".
	if !regexp.MustCompile(`(?m)^name: markmirror$`).MatchString(cfg) {
		t.Error("nfpm.yaml: name diverso da markmirror")
	}
	n := update.AssetNames("2.1.0", "amd64")
	if !strings.HasPrefix(n.Deb, "markmirror_") || !strings.HasPrefix(n.Rpm, "markmirror-") || !strings.HasPrefix(n.Pacman, "markmirror-") {
		t.Errorf("AssetNames non segue il nome del pacchetto: %s %s %s", n.Deb, n.Rpm, n.Pacman)
	}
	// Lo script di installazione deve stare accanto al binario anche nei pacchetti.
	for _, dst := range []string{"dst: /usr/bin/markmirror\n", "dst: /usr/bin/markmirror-launch\n", "dst: /usr/bin/markmirror-update.sh\n"} {
		if !strings.Contains(cfg, dst) {
			t.Errorf("nfpm.yaml: manca %s", strings.TrimSpace(dst))
		}
	}
}
