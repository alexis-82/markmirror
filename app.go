package main

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"markmirror-go/internal/document"
	"markmirror-go/internal/settings"
)

const appName = "MarkMirror"

var markdownFilters = []runtime.FileFilter{
	{DisplayName: "File Markdown (*.md, *.markdown)", Pattern: "*.md;*.markdown"},
	{DisplayName: "File di testo (*.txt)", Pattern: "*.txt"},
	{DisplayName: "Tutti i file (*.*)", Pattern: "*.*"},
}

// App è la facciata esposta al frontend: i metodi pubblici sono i binding
// Wails e delegano ai package di internal/; qui restano dialoghi ed eventi.
type App struct {
	ctx context.Context

	doc *document.Document

	mu          sync.Mutex
	quitting    bool
	initialFile string
	settings    settings.Settings

	updates updater
	gitItem *menu.MenuItem // "Commit e push", abilitata solo dentro un repo

	preFullscreen *settings.WindowState // geometria prima dello schermo intero
}

func NewApp(initialFile string) *App {
	a := &App{doc: &document.Document{}, initialFile: initialFile, settings: settings.Load()}
	a.updates = newUpdater(a)
	return a
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.restoreWindow()
	a.applyWindowTheme()
	a.updateTitle()
	go a.doc.Watch(ctx, func(path string) {
		runtime.EventsEmit(ctx, "file:changed", path)
	})
	go a.cleanupUpdates()
}

// beforeClose blocca la chiusura se ci sono modifiche non salvate e chiede
// al frontend di gestire la conferma; il frontend chiamerà Quit.
// Viene chiamata anche da Quit, quindi la finestra si salva qui.
func (a *App) beforeClose(ctx context.Context) bool {
	a.mu.Lock()
	block := a.doc.Modified() && !a.quitting
	a.mu.Unlock()
	if block {
		runtime.EventsEmit(ctx, "app:close-requested")
	} else {
		a.saveWindow()
	}
	return block
}

// Quit chiude l'applicazione senza ulteriori conferme.
func (a *App) Quit() {
	a.mu.Lock()
	a.quitting = true
	a.mu.Unlock()
	runtime.Quit(a.ctx)
}

// InitialFile apre il file passato da riga di comando, se presente.
func (a *App) InitialFile() (*document.File, error) {
	a.mu.Lock()
	path := a.initialFile
	a.initialFile = ""
	a.mu.Unlock()
	if path == "" {
		return nil, nil
	}
	return a.OpenPath(path)
}

// OpenFile mostra la finestra di apertura. Restituisce nil se annullata.
func (a *App) OpenFile() (*document.File, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "Apri file",
		DefaultDirectory: a.doc.Dir(),
		Filters:          markdownFilters,
	})
	if err != nil || path == "" {
		return nil, err
	}
	return a.OpenPath(path)
}

// OpenPath legge un file e lo rende il documento corrente.
func (a *App) OpenPath(path string) (*document.File, error) {
	f, err := a.doc.Open(path)
	if err != nil {
		return nil, err
	}
	a.updateTitle()
	return f, nil
}

// ReloadFile rilegge dal disco il documento corrente (es. dopo una modifica
// fatta da un altro programma). Restituisce nil se il documento non è salvato.
func (a *App) ReloadFile() (*document.File, error) {
	f, err := a.doc.Reload()
	if f != nil {
		a.updateTitle()
	}
	return f, err
}

// NewFile azzera il documento corrente.
func (a *App) NewFile() {
	a.doc.Reset()
	a.updateTitle()
}

// SaveFile salva sul file corrente, o chiede un nome se non esiste ancora.
// Restituisce nil se l'utente annulla.
func (a *App) SaveFile(content string) (*document.File, error) {
	path := a.doc.Path()
	if path == "" {
		return a.SaveFileAs(content)
	}
	return a.writeFile(path, content)
}

// SaveFileAs chiede sempre il nome del file. Restituisce nil se annullata.
func (a *App) SaveFileAs(content string) (*document.File, error) {
	name := "documento.md"
	if current := a.doc.Path(); current != "" {
		name = filepath.Base(current)
	}

	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:            "Salva file",
		DefaultDirectory: a.doc.Dir(),
		DefaultFilename:  name,
		Filters:          markdownFilters,
	})
	if err != nil || path == "" {
		return nil, err
	}
	if filepath.Ext(path) == "" {
		path += ".md"
	}
	return a.writeFile(path, content)
}

func (a *App) writeFile(path, content string) (*document.File, error) {
	f, err := a.doc.Save(path, content)
	if err != nil {
		return nil, err
	}
	a.updateTitle()
	return f, nil
}

// SetModified aggiorna lo stato "modificato" (usato per titolo e chiusura).
func (a *App) SetModified(modified bool) {
	if a.doc.SetModified(modified) {
		a.updateTitle()
	}
}

func (a *App) updateTitle() {
	if a.ctx == nil {
		return
	}
	runtime.WindowSetTitle(a.ctx, windowTitle(a.doc.Status()))
}

// windowTitle compone il titolo della finestra: percorso completo del
// documento (o "Senza titolo"), "*" in testa se ci sono modifiche non salvate.
func windowTitle(file string, modified bool) string {
	name := "Senza titolo"
	if file != "" {
		name = file
	}
	if modified {
		name = "*" + name
	}
	return name + " - " + appName
}

// PickImage fa scegliere un'immagine e restituisce il percorso da usare nel
// markdown, oppure "" se annullata.
func (a *App) PickImage() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "Seleziona un'immagine",
		DefaultDirectory: a.doc.Dir(),
		Filters: []runtime.FileFilter{
			{DisplayName: "Immagini", Pattern: "*.png;*.jpg;*.jpeg;*.gif;*.bmp;*.webp;*.svg"},
			{DisplayName: "Tutti i file (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil || path == "" {
		return "", err
	}
	return a.ImageLink(path), nil
}

// ImageLink converte un percorso assoluto in un riferimento per il markdown:
// relativo alla cartella del documento quando possibile, con slash in avanti.
func (a *App) ImageLink(path string) string {
	return a.doc.ImageLink(path)
}

// ExportHTML chiede dove salvare l'HTML generato dal frontend e lo scrive.
// Restituisce il percorso del file, oppure "" se l'utente annulla.
func (a *App) ExportHTML(html string) (string, error) {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:            "Esporta in HTML",
		DefaultDirectory: a.doc.Dir(),
		DefaultFilename:  a.doc.ExportName(".html"),
		Filters: []runtime.FileFilter{
			{DisplayName: "Pagina web (*.html)", Pattern: "*.html;*.htm"},
		},
	})
	if err != nil || path == "" {
		return "", err
	}
	if filepath.Ext(path) == "" {
		path += ".html"
	}
	if err := document.WriteAtomic(path, []byte(html)); err != nil {
		return "", fmt.Errorf("impossibile esportare il file: %w", err)
	}
	return path, nil
}
