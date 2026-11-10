package document

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCheckExternalChange(t *testing.T) {
	path := filepath.Join(t.TempDir(), "doc.md")
	os.WriteFile(path, []byte("uno\n"), 0o644)

	d := &Document{}
	if _, changed := d.CheckExternalChange(); changed {
		t.Fatal("no document open: change reported")
	}
	if _, err := d.Open(path); err != nil {
		t.Fatal(err)
	}
	if _, changed := d.CheckExternalChange(); changed {
		t.Fatal("unchanged file reported as changed")
	}

	// Modifica esterna: segnalata una sola volta.
	os.WriteFile(path, []byte("due, più lungo\n"), 0o644)
	if got, changed := d.CheckExternalChange(); !changed || got != path {
		t.Fatalf("external change = %q, %v; want %q, true", got, changed, path)
	}
	if _, changed := d.CheckExternalChange(); changed {
		t.Fatal("same change reported twice")
	}

	// I salvataggi dell'applicazione non sono modifiche esterne.
	if _, err := d.Save(path, "tre\n"); err != nil {
		t.Fatal(err)
	}
	if _, changed := d.CheckExternalChange(); changed {
		t.Fatal("own save reported as external change")
	}

	// Stessa dimensione, data diversa: è comunque una modifica.
	future := time.Now().Add(time.Hour)
	os.Chtimes(path, future, future)
	if _, changed := d.CheckExternalChange(); !changed {
		t.Fatal("modification time change not reported")
	}

	// File rimosso: nessuna segnalazione.
	os.Remove(path)
	if _, changed := d.CheckExternalChange(); changed {
		t.Fatal("removed file reported as changed")
	}
}
