package internal_test

import (
	"go/build"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

const module = "markmirror-go"

// forbidden sono gli import vietati ai package di internal/: il runtime Wails
// resta nella facciata App, e internal/ non deve dipendere dal package main.
var forbidden = map[string]string{
	"github.com/wailsapp/wails/v2/pkg/runtime": "runtime Wails",
	module: "package main",
}

// TestInternalPackagesStayWailsFree controlla gli import di ogni package di
// internal/, anche indiretti tra package del modulo, per Linux e Windows.
func TestInternalPackagesStayWailsFree(t *testing.T) {
	// Tutte le directory sotto internal/, anche annidate.
	var dirs []string
	err := filepath.WalkDir(".", func(path string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if e.IsDir() && path != "." {
			dirs = append(dirs, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, goos := range []string{"linux", "windows"} {
		ctx := build.Default
		ctx.GOOS = goos
		seen := map[string]bool{} // ogni package si controlla una volta per GOOS
		for _, dir := range dirs {
			pkg, err := ctx.ImportDir(dir, 0)
			if err != nil {
				continue // non è un package Go (o nessun file per questo GOOS)
			}
			path := module + "/internal/" + filepath.ToSlash(dir)
			if seen[path] {
				continue
			}
			seen[path] = true
			checkImports(t, &ctx, goos, path, pkg.Imports, seen)
		}
	}
}

func checkImports(t *testing.T, ctx *build.Context, goos, from string, imports []string, seen map[string]bool) {
	t.Helper()
	for _, imp := range imports {
		if what, bad := forbidden[imp]; bad {
			t.Errorf("%s (%s) importa %s (%s)", from, goos, imp, what)
		}
		if seen[imp] || !strings.HasPrefix(imp, module+"/internal/") {
			continue
		}
		seen[imp] = true
		dir := filepath.FromSlash(strings.TrimPrefix(imp, module+"/internal/"))
		pkg, err := ctx.ImportDir(dir, 0)
		if err != nil {
			t.Errorf("%s: %v", imp, err)
			continue
		}
		checkImports(t, ctx, goos, imp, pkg.Imports, seen)
	}
}
