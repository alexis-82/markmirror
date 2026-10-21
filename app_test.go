package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
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

	a := NewApp("")
	res, err := a.OpenPath(path)
	if err != nil {
		t.Fatal(err)
	}
	if want := "# Titolo\n\ntesto\n"; res.Content != want {
		t.Fatalf("content = %q, want %q", res.Content, want)
	}

	if _, err := a.SaveFile(res.Content + "altro\n"); err != nil {
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

	a := NewApp("")
	if _, err := a.OpenPath(doc); err != nil {
		t.Fatal(err)
	}
	got := a.ImageLink(filepath.Join(dir, "img", "foto.png"))
	if got != "img/foto.png" {
		t.Fatalf("ImageLink = %q", got)
	}
}

func TestLocalFilesServesOnlyImages(t *testing.T) {
	dir := t.TempDir()
	doc := filepath.Join(dir, "doc.md")
	os.WriteFile(doc, []byte("x"), 0o644)
	os.MkdirAll(filepath.Join(dir, "img"), 0o755)
	os.WriteFile(filepath.Join(dir, "img", "a b.png"), []byte("\x89PNG\r\n\x1a\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "secret.txt"), []byte("segreto"), 0o644)

	a := NewApp("")
	if _, err := a.OpenPath(doc); err != nil {
		t.Fatal(err)
	}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusTeapot) })
	handler := a.localFiles(next)

	get := func(target string) int {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
		return rec.Code
	}
	q := func(p string) string { return localFilePath + "?path=" + url.QueryEscape(p) }

	cases := []struct {
		name   string
		target string
		want   int
	}{
		{"relative image", q("img/a b.png"), http.StatusOK},
		{"absolute image", q(filepath.Join(dir, "img", "a b.png")), http.StatusOK},
		{"non-image refused", q("secret.txt"), http.StatusNotFound},
		{"missing image", q("img/nope.png"), http.StatusNotFound},
		{"other paths passed through", "/index.html", http.StatusTeapot},
	}
	for _, c := range cases {
		if got := get(c.target); got != c.want {
			t.Errorf("%s: status = %d, want %d", c.name, got, c.want)
		}
	}
}

func TestReloadFileReadsExternalChanges(t *testing.T) {
	a := NewApp("")
	if res, err := a.ReloadFile(); res != nil || err != nil {
		t.Fatalf("unsaved document: ReloadFile = %v, %v; want nil, nil", res, err)
	}

	path := filepath.Join(t.TempDir(), "doc.md")
	os.WriteFile(path, []byte("prima\n"), 0o644)
	if _, err := a.OpenPath(path); err != nil {
		t.Fatal(err)
	}
	a.SetModified(true)

	os.WriteFile(path, []byte("dopo\r\n"), 0o644)
	res, err := a.ReloadFile()
	if err != nil {
		t.Fatal(err)
	}
	if res.Content != "dopo\n" {
		t.Fatalf("content = %q, want %q", res.Content, "dopo\n")
	}
	if a.modified {
		t.Fatal("document still marked as modified after reload")
	}
}

func TestWindowTitle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nome.md")
	cases := []struct {
		file     string
		modified bool
		want     string
	}{
		{"", false, "Senza titolo - MarkMirror"},
		{"", true, "*Senza titolo - MarkMirror"},
		{path, false, path + " - MarkMirror"},
		{path, true, "*" + path + " - MarkMirror"},
	}
	for _, c := range cases {
		if got := windowTitle(c.file, c.modified); got != c.want {
			t.Errorf("windowTitle(%q, %v) = %q, want %q", c.file, c.modified, got, c.want)
		}
	}
}

// TestWindowTitleFollowsDocument percorre i criteri di accettazione della
// spec sulle operazioni reali dell'app. Salva con nome passa da writeFile,
// perché la finestra di dialogo richiede il runtime di Wails.
func TestWindowTitleFollowsDocument(t *testing.T) {
	root := t.TempDir()
	dirA := filepath.Join(root, "a")
	dirB := filepath.Join(root, "b")
	dirC := filepath.Join(root, "c")
	for _, d := range []string{dirA, dirB, dirC} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	readmeA := filepath.Join(dirA, "README.md")
	readmeB := filepath.Join(dirB, "README.md")
	os.WriteFile(readmeA, []byte("# A\n"), 0o644)
	os.WriteFile(readmeB, []byte("# B\n"), 0o644)

	title := func(a *App) string {
		a.mu.Lock()
		defer a.mu.Unlock()
		return windowTitle(a.currentFile, a.modified)
	}
	expect := func(a *App, want string) {
		t.Helper()
		if got := title(a); got != want {
			t.Fatalf("title = %q, want %q", got, want)
		}
	}

	// Avvio con un percorso relativo (doppio clic / riga di comando).
	t.Chdir(root)
	a := NewApp(filepath.Join("a", "README.md"))
	expect(a, "Senza titolo - MarkMirror")
	if _, err := a.InitialFile(); err != nil {
		t.Fatal(err)
	}
	expect(a, readmeA+" - MarkMirror")

	// Due file omonimi in cartelle diverse sono distinguibili.
	if _, err := a.OpenPath(readmeB); err != nil {
		t.Fatal(err)
	}
	expect(a, readmeB+" - MarkMirror")

	// "*" su modifica, sparisce al salvataggio.
	a.SetModified(true)
	expect(a, "*"+readmeB+" - MarkMirror")
	if _, err := a.SaveFile("# B modificato\n"); err != nil {
		t.Fatal(err)
	}
	expect(a, readmeB+" - MarkMirror")

	// Salva con nome in un'altra cartella.
	a.SetModified(true)
	other := filepath.Join(dirC, "copia.md")
	if _, err := a.writeFile(other, "# copia\n"); err != nil {
		t.Fatal(err)
	}
	expect(a, other+" - MarkMirror")

	// Nuovo documento.
	a.NewFile()
	expect(a, "Senza titolo - MarkMirror")
	a.SetModified(true)
	expect(a, "*Senza titolo - MarkMirror")
}
