package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	appName    = "MarkMirror"
	appVersion = "2.0.0"
	bom        = "\xef\xbb\xbf" // BOM UTF-8
)

var markdownFilters = []runtime.FileFilter{
	{DisplayName: "File Markdown (*.md, *.markdown)", Pattern: "*.md;*.markdown"},
	{DisplayName: "File di testo (*.txt)", Pattern: "*.txt"},
	{DisplayName: "Tutti i file (*.*)", Pattern: "*.*"},
}

// FileResult descrive un documento aperto o salvato.
type FileResult struct {
	Path    string `json:"path"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

// App contiene lo stato del documento corrente ed espone i metodi al frontend.
type App struct {
	ctx context.Context

	ioMu sync.Mutex // serializza letture/scritture del documento con il watcher

	mu          sync.Mutex
	currentFile string
	stamp       fileStamp // versione su disco del documento corrente
	crlf        bool
	modified    bool
	quitting    bool
	initialFile string
	settings    Settings
}

func NewApp(initialFile string) *App {
	return &App{initialFile: initialFile, settings: loadSettings()}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.applyWindowTheme()
	a.updateTitle()
	go a.watchFile()
}

// beforeClose blocca la chiusura se ci sono modifiche non salvate e chiede
// al frontend di gestire la conferma; il frontend chiamerà Quit.
func (a *App) beforeClose(ctx context.Context) bool {
	a.mu.Lock()
	block := a.modified && !a.quitting
	a.mu.Unlock()
	if block {
		runtime.EventsEmit(ctx, "app:close-requested")
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
func (a *App) InitialFile() (*FileResult, error) {
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
func (a *App) OpenFile() (*FileResult, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "Apri file",
		DefaultDirectory: a.docDir(),
		Filters:          markdownFilters,
	})
	if err != nil || path == "" {
		return nil, err
	}
	return a.OpenPath(path)
}

// OpenPath legge un file e lo rende il documento corrente.
func (a *App) OpenPath(path string) (*FileResult, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	a.ioMu.Lock()
	defer a.ioMu.Unlock()
	// Stat prima della lettura: se il file cambia nel mezzo, il watcher lo
	// segnalerà di nuovo invece di perdere la modifica.
	stamp, _ := statStamp(abs)
	data, err := os.ReadFile(abs)
	if err != nil {
		return nil, fmt.Errorf("impossibile aprire il file: %w", err)
	}
	content := strings.TrimPrefix(string(data), bom)
	crlf := strings.Contains(content, "\r\n")
	content = strings.ReplaceAll(content, "\r\n", "\n")

	a.mu.Lock()
	a.currentFile = abs
	a.stamp = stamp
	a.crlf = crlf
	a.modified = false
	a.mu.Unlock()
	a.updateTitle()

	return &FileResult{Path: abs, Name: filepath.Base(abs), Content: content}, nil
}

// ReloadFile rilegge dal disco il documento corrente (es. dopo una modifica
// fatta da un altro programma). Restituisce nil se il documento non è salvato.
func (a *App) ReloadFile() (*FileResult, error) {
	a.mu.Lock()
	path := a.currentFile
	a.mu.Unlock()
	if path == "" {
		return nil, nil
	}
	return a.OpenPath(path)
}

// NewFile azzera il documento corrente.
func (a *App) NewFile() {
	a.mu.Lock()
	a.currentFile = ""
	a.stamp = fileStamp{}
	a.crlf = false
	a.modified = false
	a.mu.Unlock()
	a.updateTitle()
}

// SaveFile salva sul file corrente, o chiede un nome se non esiste ancora.
// Restituisce nil se l'utente annulla.
func (a *App) SaveFile(content string) (*FileResult, error) {
	a.mu.Lock()
	path := a.currentFile
	a.mu.Unlock()
	if path == "" {
		return a.SaveFileAs(content)
	}
	return a.writeFile(path, content)
}

// SaveFileAs chiede sempre il nome del file. Restituisce nil se annullata.
func (a *App) SaveFileAs(content string) (*FileResult, error) {
	a.mu.Lock()
	name := "documento.md"
	if a.currentFile != "" {
		name = filepath.Base(a.currentFile)
	}
	a.mu.Unlock()

	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:            "Salva file",
		DefaultDirectory: a.docDir(),
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

func (a *App) writeFile(path, content string) (*FileResult, error) {
	a.mu.Lock()
	crlf := a.crlf
	a.mu.Unlock()
	if crlf {
		content = strings.ReplaceAll(content, "\n", "\r\n")
	}
	a.ioMu.Lock()
	defer a.ioMu.Unlock()
	if err := writeFileAtomic(path, []byte(content)); err != nil {
		return nil, fmt.Errorf("impossibile salvare il file: %w", err)
	}
	stamp, _ := statStamp(path)

	a.mu.Lock()
	a.currentFile = path
	a.stamp = stamp
	a.modified = false
	a.mu.Unlock()
	a.updateTitle()

	return &FileResult{Path: path, Name: filepath.Base(path)}, nil
}

// writeFileAtomic scrive su un file temporaneo e poi lo rinomina, così un
// errore a metà scrittura non tronca il documento esistente.
func writeFileAtomic(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".markmirror-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
}

// SetModified aggiorna lo stato "modificato" (usato per titolo e chiusura).
func (a *App) SetModified(modified bool) {
	a.mu.Lock()
	changed := a.modified != modified
	a.modified = modified
	a.mu.Unlock()
	if changed {
		a.updateTitle()
	}
}

func (a *App) updateTitle() {
	if a.ctx == nil {
		return
	}
	a.mu.Lock()
	name := "Senza titolo"
	if a.currentFile != "" {
		name = filepath.Base(a.currentFile)
	}
	if a.modified {
		name = "*" + name
	}
	a.mu.Unlock()
	runtime.WindowSetTitle(a.ctx, name+" - "+appName)
}

// docDir restituisce la cartella del documento corrente, o "" se non salvato.
func (a *App) docDir() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.currentFile == "" {
		return ""
	}
	return filepath.Dir(a.currentFile)
}

// PickImage fa scegliere un'immagine e restituisce il percorso da usare nel
// markdown, oppure "" se annullata.
func (a *App) PickImage() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "Seleziona un'immagine",
		DefaultDirectory: a.docDir(),
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
	ref := path
	if dir := a.docDir(); dir != "" {
		if rel, err := filepath.Rel(dir, path); err == nil {
			ref = rel
		}
	}
	return filepath.ToSlash(ref)
}
