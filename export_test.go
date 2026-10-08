package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExportName(t *testing.T) {
	a := NewApp("")
	if got := a.exportName(".html"); got != "documento.html" {
		t.Fatalf("unsaved document: exportName = %q", got)
	}

	path := filepath.Join(t.TempDir(), "note.di.oggi.md")
	os.WriteFile(path, []byte("x"), 0o644)
	if _, err := a.OpenPath(path); err != nil {
		t.Fatal(err)
	}
	if got := a.exportName(".html"); got != "note.di.oggi.html" {
		t.Fatalf("exportName = %q, want note.di.oggi.html", got)
	}
}
