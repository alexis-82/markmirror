package main

import (
	"os"
	"path/filepath"
	"testing"
)

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
		return windowTitle(a.doc.Status())
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

// I metodi git della facciata lavorano solo su un documento salvato.
func TestGitRequiresSavedDocument(t *testing.T) {
	a := NewApp("")
	if info := a.GitStatus(); info.Repo != "" {
		t.Fatalf("unsaved document: Repo = %q, want empty", info.Repo)
	}
	if _, err := a.GitHasChanges(); err == nil {
		t.Error("GitHasChanges: expected error for unsaved document")
	}
	if err := a.GitCommit("msg"); err == nil {
		t.Error("GitCommit: expected error for unsaved document")
	}
	if err := a.GitPush(); err == nil {
		t.Error("GitPush: expected error for unsaved document")
	}
}
