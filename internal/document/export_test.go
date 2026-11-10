package document

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExportName(t *testing.T) {
	d := &Document{}
	if got := d.ExportName(".html"); got != "documento.html" {
		t.Fatalf("unsaved document: exportName = %q", got)
	}

	path := filepath.Join(t.TempDir(), "note.di.oggi.md")
	os.WriteFile(path, []byte("x"), 0o644)
	if _, err := d.Open(path); err != nil {
		t.Fatal(err)
	}
	if got := d.ExportName(".html"); got != "note.di.oggi.html" {
		t.Fatalf("exportName = %q, want note.di.oggi.html", got)
	}
}
