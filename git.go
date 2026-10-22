package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// gitTimeout limita ogni comando git: senza terminale una richiesta di
// credenziali non può ricevere risposta e il comando resterebbe appeso.
// È ampio perché il login di Git Credential Manager passa dal browser.
const gitTimeout = 3 * time.Minute

// GitInfo descrive il repository che contiene il documento corrente.
type GitInfo struct {
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

// gitInfo individua il repository che contiene path. Restituisce un GitInfo
// vuoto se git non è installato o il file non sta in un repo.
func gitInfo(path string) GitInfo {
	if path == "" {
		return GitInfo{}
	}
	if _, err := exec.LookPath("git"); err != nil {
		return GitInfo{}
	}
	dir := filepath.Dir(path)
	root, err := runGit(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return GitInfo{}
	}
	branch, _ := runGit(dir, "symbolic-ref", "--short", "-q", "HEAD")
	return GitInfo{Repo: filepath.FromSlash(root), Branch: branch}
}

// GitStatus restituisce il repository del documento corrente e abilita la
// voce di menu "Commit e push" solo se il documento ne fa parte.
func (a *App) GitStatus() GitInfo {
	a.mu.Lock()
	path := a.currentFile
	a.mu.Unlock()
	info := gitInfo(path)
	if a.gitItem != nil && a.gitItem.Disabled != (info.Repo == "") {
		a.gitItem.Disabled = info.Repo == ""
		runtime.MenuUpdateApplicationMenu(a.ctx)
	}
	return info
}

// gitFile restituisce il documento corrente, o un errore se non è salvato.
func (a *App) gitFile() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.currentFile == "" {
		return "", errors.New("il documento non è salvato")
	}
	return a.currentFile, nil
}

// GitHasChanges indica se il documento corrente differisce dall'ultimo
// commit (o non è ancora tracciato).
func (a *App) GitHasChanges() (bool, error) {
	path, err := a.gitFile()
	if err != nil {
		return false, err
	}
	out, err := runGit(filepath.Dir(path), "status", "--porcelain", "--", filepath.Base(path))
	return out != "", err
}

// GitCommit crea un commit con il solo documento corrente: le altre
// modifiche del repository, anche se già in staging, restano fuori.
func (a *App) GitCommit(message string) error {
	path, err := a.gitFile()
	if err != nil {
		return err
	}
	if strings.TrimSpace(message) == "" {
		return errors.New("il messaggio di commit è vuoto")
	}
	dir, name := filepath.Dir(path), filepath.Base(path)
	// add serve per i file non ancora tracciati; con il percorso esplicito
	// commit include solo quel file (--only è implicito).
	if _, err := runGit(dir, "add", "--", name); err != nil {
		return err
	}
	_, err = runGit(dir, "commit", "-m", message, "--", name)
	return err
}

// GitPush invia il branch corrente al suo remote. Se il branch non ha ancora
// un upstream lo crea sul remote del repository (origin se ce n'è più d'uno).
func (a *App) GitPush() error {
	path, err := a.gitFile()
	if err != nil {
		return err
	}
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
