// Package document gestisce il documento aperto: lettura e scrittura su
// disco, terminazioni di riga, modifiche esterne e immagini locali.
package document

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const bom = "\xef\xbb\xbf" // BOM UTF-8

// File descrive un documento aperto o salvato.
type File struct {
	Path    string `json:"path"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

// Document è il documento corrente. Il valore zero è un documento nuovo,
// non ancora salvato.
type Document struct {
	ioMu sync.Mutex // serializza letture/scritture del documento con il watcher

	mu       sync.Mutex
	path     string
	stamp    fileStamp // versione su disco del documento corrente
	crlf     bool
	modified bool
}

// Open legge un file e lo rende il documento corrente.
func (d *Document) Open(path string) (*File, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	d.ioMu.Lock()
	defer d.ioMu.Unlock()
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

	d.mu.Lock()
	d.path = abs
	d.stamp = stamp
	d.crlf = crlf
	d.modified = false
	d.mu.Unlock()

	return &File{Path: abs, Name: filepath.Base(abs), Content: content}, nil
}

// Reload rilegge dal disco il documento corrente (es. dopo una modifica
// fatta da un altro programma). Restituisce nil se il documento non è salvato.
func (d *Document) Reload() (*File, error) {
	path := d.Path()
	if path == "" {
		return nil, nil
	}
	return d.Open(path)
}

// Reset azzera il documento corrente.
func (d *Document) Reset() {
	d.mu.Lock()
	d.path = ""
	d.stamp = fileStamp{}
	d.crlf = false
	d.modified = false
	d.mu.Unlock()
}

// Save scrive content in path, con le terminazioni di riga del file
// originale, e lo rende il documento corrente.
func (d *Document) Save(path, content string) (*File, error) {
	d.mu.Lock()
	crlf := d.crlf
	d.mu.Unlock()
	if crlf {
		content = strings.ReplaceAll(content, "\n", "\r\n")
	}
	d.ioMu.Lock()
	defer d.ioMu.Unlock()
	if err := WriteAtomic(path, []byte(content)); err != nil {
		return nil, fmt.Errorf("impossibile salvare il file: %w", err)
	}
	stamp, _ := statStamp(path)

	d.mu.Lock()
	d.path = path
	d.stamp = stamp
	d.modified = false
	d.mu.Unlock()

	return &File{Path: path, Name: filepath.Base(path)}, nil
}

// WriteAtomic scrive su un file temporaneo e poi lo rinomina, così un
// errore a metà scrittura non tronca il documento esistente.
func WriteAtomic(path string, data []byte) error {
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

// SetModified aggiorna lo stato "modificato" e dice se è cambiato.
func (d *Document) SetModified(modified bool) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	changed := d.modified != modified
	d.modified = modified
	return changed
}

// Path restituisce il percorso del documento, o "" se non è salvato.
func (d *Document) Path() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.path
}

// Modified dice se ci sono modifiche non salvate.
func (d *Document) Modified() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.modified
}

// Status restituisce percorso e stato "modificato" letti insieme.
func (d *Document) Status() (path string, modified bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.path, d.modified
}

// Dir restituisce la cartella del documento corrente, o "" se non salvato.
func (d *Document) Dir() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.path == "" {
		return ""
	}
	return filepath.Dir(d.path)
}

// ImageLink converte un percorso assoluto in un riferimento per il markdown:
// relativo alla cartella del documento quando possibile, con slash in avanti.
func (d *Document) ImageLink(path string) string {
	ref := path
	if dir := d.Dir(); dir != "" {
		if rel, err := filepath.Rel(dir, path); err == nil {
			ref = rel
		}
	}
	return filepath.ToSlash(ref)
}

// ExportName è il nome proposto per un'esportazione: il nome del documento
// con l'estensione indicata, o "documento" se non è ancora salvato.
func (d *Document) ExportName(ext string) string {
	path := d.Path()
	if path == "" {
		return "documento" + ext
	}
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base)) + ext
}
