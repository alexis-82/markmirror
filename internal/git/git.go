// Package git esegue i comandi git sul documento corrente.
package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// gitTimeout limita ogni comando git: senza terminale una richiesta di
// credenziali non può ricevere risposta e il comando resterebbe appeso.
// È ampio perché il login di Git Credential Manager passa dal browser.
const gitTimeout = 3 * time.Minute

// Info descrive il repository che contiene il documento corrente.
type Info struct {
	Repo   string `json:"repo"`   // cartella radice; "" se il documento non è in un repo
	Branch string `json:"branch"` // "" se HEAD è staccato
}

// runGit esegue git nella cartella indicata e restituisce l'output ripulito.
// In caso di errore il messaggio è quello stampato da git.
func runGit(dir string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gitTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	hideConsole(cmd)
	out, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(out))
	if ctx.Err() != nil {
		return text, fmt.Errorf("git %s: tempo scaduto", args[0])
	}
	if err != nil {
		if text == "" {
			text = err.Error()
		}
		return text, errors.New(text)
	}
	return text, nil
}

// Status individua il repository che contiene path. Restituisce un Info
// vuoto se git non è installato o il file non sta in un repo.
func Status(path string) Info {
	if path == "" {
		return Info{}
	}
	if _, err := exec.LookPath("git"); err != nil {
		return Info{}
	}
	dir := filepath.Dir(path)
	root, err := runGit(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return Info{}
	}
	branch, _ := runGit(dir, "symbolic-ref", "--short", "-q", "HEAD")
	return Info{Repo: filepath.FromSlash(root), Branch: branch}
}

// HasChanges indica se il file path differisce dall'ultimo commit (o non è
// ancora tracciato).
func HasChanges(path string) (bool, error) {
	out, err := runGit(filepath.Dir(path), "status", "--porcelain", "--", filepath.Base(path))
	return out != "", err
}

// Commit crea un commit con il solo file path: le altre modifiche del
// repository, anche se già in staging, restano fuori.
func Commit(path, message string) error {
	if strings.TrimSpace(message) == "" {
		return errors.New("il messaggio di commit è vuoto")
	}
	dir, name := filepath.Dir(path), filepath.Base(path)
	// add serve per i file non ancora tracciati; con il percorso esplicito
	// commit include solo quel file (--only è implicito).
	if _, err := runGit(dir, "add", "--", name); err != nil {
		return err
	}
	_, err := runGit(dir, "commit", "-m", message, "--", name)
	return err
}

// Push invia il branch corrente del repository di path al suo remote. Se il
// branch non ha ancora un upstream lo crea sul remote del repository (origin
// se ce n'è più d'uno).
func Push(path string) error {
	dir := filepath.Dir(path)
	if _, err := runGit(dir, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}"); err == nil {
		_, err = runGit(dir, "push")
		return err
	}
	remotes, err := runGit(dir, "remote")
	if err != nil {
		return err
	}
	remote, err := pushRemote(strings.Fields(remotes))
	if err != nil {
		return err
	}
	_, err = runGit(dir, "push", "-u", remote, "HEAD")
	return err
}

// pushRemote sceglie il remote per un branch senza upstream.
func pushRemote(remotes []string) (string, error) {
	switch {
	case len(remotes) == 0:
		return "", errors.New("il repository non ha un remote configurato")
	case len(remotes) == 1:
		return remotes[0], nil
	}
	for _, r := range remotes {
		if r == "origin" {
			return r, nil
		}
	}
	return "", fmt.Errorf("il branch non ha un upstream e ci sono più remote (%s): impostalo con git push -u", strings.Join(remotes, ", "))
}
