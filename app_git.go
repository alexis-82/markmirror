package main

import (
	"errors"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"markmirror-go/internal/git"
)

// GitStatus restituisce il repository del documento corrente e abilita la
// voce di menu "Commit e push" solo se il documento ne fa parte.
func (a *App) GitStatus() git.Info {
	info := git.Status(a.doc.Path())
	if a.gitItem != nil && a.gitItem.Disabled != (info.Repo == "") {
		a.gitItem.Disabled = info.Repo == ""
		runtime.MenuUpdateApplicationMenu(a.ctx)
	}
	return info
}

// gitFile restituisce il documento corrente, o un errore se non è salvato.
func (a *App) gitFile() (string, error) {
	path := a.doc.Path()
	if path == "" {
		return "", errors.New("il documento non è salvato")
	}
	return path, nil
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
