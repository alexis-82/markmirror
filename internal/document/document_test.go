package document

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenSavePreservesLineEndingsAndStripsBOM(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(path, []byte(bom+"# Titolo\r\n\r\ntesto\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	d := &Document{}
	res, err := d.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if want := "# Titolo\n\ntesto\n"; res.Content != want {
		t.Fatalf("content = %q, want %q", res.Content, want)
	}

	if _, err := d.Save(path, res.Content+"altro\n"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if want := "# Titolo\r\n\r\ntesto\r\naltro\r\n"; string(data) != want {
		t.Fatalf("saved = %q, want %q", data, want)
	}

	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("expected no leftover temp files, got %d entries", len(entries))
	}
}

func TestImageLinkIsRelativeToDocument(t *testing.T) {
	dir := t.TempDir()
	doc := filepath.Join(dir, "doc.md")
	os.WriteFile(doc, []byte("x"), 0o644)

	d := &Document{}
	if _, err := d.Open(doc); err != nil {
		t.Fatal(err)
	}
	got := d.ImageLink(filepath.Join(dir, "img", "foto.png"))
	if got != "img/foto.png" {
		t.Fatalf("ImageLink = %q", got)
	}
}

func TestReloadReadsExternalChanges(t *testing.T) {
	d := &Document{}
	if res, err := d.Reload(); res != nil || err != nil {
		t.Fatalf("unsaved document: Reload = %v, %v; want nil, nil", res, err)
	}

	path := filepath.Join(t.TempDir(), "doc.md")
	os.WriteFile(path, []byte("prima\n"), 0o644)
	if _, err := d.Open(path); err != nil {
		t.Fatal(err)
	}
	d.SetModified(true)

	os.WriteFile(path, []byte("dopo\r\n"), 0o644)
	res, err := d.Reload()
	if err != nil {
		t.Fatal(err)
	}
	if res.Content != "dopo\n" {
		t.Fatalf("content = %q, want %q", res.Content, "dopo\n")
	}
	if d.Modified() {
		t.Fatal("document still marked as modified after reload")
	}
}
