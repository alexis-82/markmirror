package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ExportHTML chiede dove salvare l'HTML generato dal frontend e lo scrive.
// Restituisce il percorso del file, oppure "" se l'utente annulla.
func (a *App) ExportHTML(html string) (string, error) {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:            "Esporta in HTML",
		DefaultDirectory: a.docDir(),
		DefaultFilename:  a.exportName(".html"),
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
	if err := writeFileAtomic(path, []byte(html)); err != nil {
		return "", fmt.Errorf("impossibile esportare il file: %w", err)
	}
	return path, nil
}

// exportName è il nome proposto per un'esportazione: il nome del documento
// con l'estensione indicata, o "documento" se non è ancora salvato.
func (a *App) exportName(ext string) string {
	a.mu.Lock()
	path := a.currentFile
	a.mu.Unlock()
	if path == "" {
		return "documento" + ext
	}
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base)) + ext
}
