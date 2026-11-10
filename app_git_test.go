package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// newGitRepo crea un repository con un primo commit e un remote bare
// ("origin") senza upstream impostato. Restituisce la cartella di lavoro e
// quella del remote.
func newGitRepo(t *testing.T) (string, string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git non installato")
	}
	base := t.TempDir()
	work, remote := filepath.Join(base, "work"), filepath.Join(base, "remote.git")
	mustGit(t, base, "init", "-q", "--bare", remote)
	mustGit(t, base, "init", "-q", work)
	mustGit(t, work, "config", "user.name", "Test")
	mustGit(t, work, "config", "user.email", "test@example.com")
	mustGit(t, work, "config", "commit.gpgsign", "false")
	mustGit(t, work, "remote", "add", "origin", remote)
	writeFile(t, filepath.Join(work, "base.md"), "base\n")
	mustGit(t, work, "add", "base.md")
	mustGit(t, work, "commit", "-q", "-m", "base")
	return work, remote
}

func mustGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func openDoc(t *testing.T, path string) *App {
	t.Helper()
	a := NewApp("")
	if _, err := a.OpenPath(path); err != nil {
		t.Fatal(err)
	}
	return a
}

func TestGitStatusOutsideRepo(t *testing.T) {
	doc := filepath.Join(t.TempDir(), "doc.md")
	writeFile(t, doc, "x")
	if info := openDoc(t, doc).GitStatus(); info.Repo != "" {
		t.Fatalf("Repo = %q, want empty", info.Repo)
	}
	if info := NewApp("").GitStatus(); info.Repo != "" {
		t.Fatalf("unsaved document: Repo = %q, want empty", info.Repo)
	}
}

func TestGitStatusInsideRepo(t *testing.T) {
	work, _ := newGitRepo(t)
	sub := filepath.Join(work, "docs")
	os.Mkdir(sub, 0o755)
	doc := filepath.Join(sub, "doc.md")
	writeFile(t, doc, "x")

	info := openDoc(t, doc).GitStatus()
	want, _ := filepath.EvalSymlinks(work)
	got, _ := filepath.EvalSymlinks(info.Repo)
	if !strings.EqualFold(got, want) {
		t.Fatalf("Repo = %q, want %q", info.Repo, work)
	}
	if info.Branch == "" {
		t.Fatal("Branch is empty")
	}
}

func TestGitCommitOnlyCurrentFile(t *testing.T) {
	work, _ := newGitRepo(t)
	doc := filepath.Join(work, "doc.md")
	writeFile(t, doc, "nuovo\n")
	// Un altro file già in staging non deve finire nel commit.
	writeFile(t, filepath.Join(work, "other.md"), "altro\n")
	mustGit(t, work, "add", "other.md")

	a := openDoc(t, doc)
	if changed, err := a.GitHasChanges(); err != nil || !changed {
		t.Fatalf("GitHasChanges = %v, %v; want true (untracked)", changed, err)
	}
	if err := a.GitCommit("Aggiunge doc"); err != nil {
		t.Fatal(err)
	}

	files := mustGit(t, work, "show", "--name-only", "--format=", "HEAD")
	if files != "doc.md" {
		t.Fatalf("committed files = %q, want doc.md", files)
	}
	if msg := mustGit(t, work, "log", "-1", "--format=%s"); msg != "Aggiunge doc" {
		t.Fatalf("message = %q", msg)
	}
	if staged := mustGit(t, work, "diff", "--cached", "--name-only"); staged != "other.md" {
		t.Fatalf("staged after commit = %q, want other.md", staged)
	}
	if changed, _ := a.GitHasChanges(); changed {
		t.Fatal("GitHasChanges = true after commit")
	}
}

func TestGitCommitRejectsEmptyMessage(t *testing.T) {
	work, _ := newGitRepo(t)
	doc := filepath.Join(work, "base.md")
	writeFile(t, doc, "modificato\n")
	if err := openDoc(t, doc).GitCommit("  \n"); err == nil {
		t.Fatal("expected error for empty message")
	}
}

func TestGitPushSetsUpstream(t *testing.T) {
	work, remote := newGitRepo(t)
	doc := filepath.Join(work, "base.md")
	a := openDoc(t, doc)

	if err := a.GitPush(); err != nil {
		t.Fatal(err)
	}
	branch := mustGit(t, work, "symbolic-ref", "--short", "HEAD")
	if got, want := mustGit(t, remote, "rev-parse", branch), mustGit(t, work, "rev-parse", "HEAD"); got != want {
		t.Fatalf("remote %s = %s, want %s", branch, got, want)
	}
	mustGit(t, work, "rev-parse", "--abbrev-ref", "@{u}") // upstream ora impostato

	// Secondo push con upstream: usa il ramo normale.
	writeFile(t, doc, "modificato\n")
	if err := a.GitCommit("Modifica"); err != nil {
		t.Fatal(err)
	}
	if err := a.GitPush(); err != nil {
		t.Fatal(err)
	}
	if got, want := mustGit(t, remote, "rev-parse", branch), mustGit(t, work, "rev-parse", "HEAD"); got != want {
		t.Fatalf("after second push remote = %s, want %s", got, want)
	}
}

func TestGitPushWithoutRemote(t *testing.T) {
	work, _ := newGitRepo(t)
	mustGit(t, work, "remote", "remove", "origin")
	if err := openDoc(t, filepath.Join(work, "base.md")).GitPush(); err == nil {
		t.Fatal("expected error without remote")
	}
}
