//go:build linux

package update

import (
	"errors"
	"reflect"
	"testing"
)

func TestOwnedByPackage(t *testing.T) {
	owner := func(name string, err error) func(string) (string, error) {
		return func(string) (string, error) { return name, err }
	}

	if !ownedByPackage("/usr/bin/markmirror", owner("markmirror", nil)) {
		t.Fatal("binary of the markmirror package: expected true")
	}
	if ownedByPackage("/usr/bin/markmirror", owner("", errors.New(`exec: "rpm": executable file not found in $PATH`))) {
		t.Fatal("rpm not installed: expected false")
	}
	if ownedByPackage("/usr/bin/markmirror", owner("file /usr/bin/markmirror is not owned by any package", errors.New("exit status 1"))) {
		t.Fatal("file owned by no package: expected false")
	}
	if ownedByPackage("/usr/bin/markmirror", owner("other", nil)) {
		t.Fatal("binary of another package: expected false")
	}
	// pacman -Qqo stampa il nome seguito da un a capo
	if !ownedByPackage("/usr/bin/markmirror", owner("markmirror\n", nil)) {
		t.Fatal("pacman output with newline: expected true")
	}
	if ownedByPackage("/usr/bin/markmirror", owner("", errors.New(`exec: "pacman": executable file not found in $PATH`))) {
		t.Fatal("pacman not installed: expected false")
	}

	called := false
	query := func(string) (string, error) { called = true; return "markmirror", nil }
	if ownedByPackage("/usr/local/bin/markmirror", query) || called {
		t.Fatal("install.sh layout: expected false without querying the package manager")
	}
}

func TestInstallCommand(t *testing.T) {
	for exe, script := range map[string]string{
		"/usr/bin/markmirror":       "/usr/bin/markmirror-update.sh",
		"/usr/local/bin/markmirror": "/usr/local/bin/markmirror-update.sh",
	} {
		got := installCommand("/usr/bin/pkexec", exe, "/tmp/markmirror-update-1/f.deb", "abc")
		want := []string{"/usr/bin/pkexec", script, "/tmp/markmirror-update-1/f.deb", "abc"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("installCommand(%s) = %v, want %v", exe, got, want)
		}
	}
}
