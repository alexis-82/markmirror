package main

import (
	"errors"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"markmirror-go/internal/git"
)

// GitStatus restituisce il repository del documento corrente e abilita la
// voce di menu "Commit e push" solo se il documento ne fa parte.
func (a *App) GitStatus() git.Info {
	a.mu.Lock()
	path := a.currentFile
	a.mu.Unlock()
	info := git.Status(path)
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
	return git.HasChanges(path)
}

// GitCommit crea un commit con il solo documento corrente.
func (a *App) GitCommit(message string) error {
	path, err := a.gitFile()
	if err != nil {
		return err
	}
	return git.Commit(path, message)
}

// GitPush invia il branch corrente al suo remote.
func (a *App) GitPush() error {
	path, err := a.gitFile()
	if err != nil {
		return err
	}
	return git.Push(path)
}
