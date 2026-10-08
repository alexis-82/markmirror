package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// localFilePath è l'endpoint da cui l'anteprima carica le immagini locali:
// /localfile?path=<percorso relativo al documento o assoluto>
const localFilePath = "/localfile"

var imageExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".bmp": true,
	".webp": true, ".svg": true, ".ico": true, ".avif": true,
}

// localFiles è un middleware dell'AssetServer che serve le immagini locali
// referenziate dal documento. Serve solo file con estensione di immagine.
func (a *App) localFiles(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != localFilePath {
			next.ServeHTTP(w, r)
			return
		}
		path, ok := a.resolveLocal(r.URL.Query().Get("path"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		f, err := os.Open(path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()
		info, err := f.Stat()
		if err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeContent(w, r, info.Name(), info.ModTime(), f)
	})
}

// resolveLocal risolve un riferimento rispetto alla cartella del documento
// (o alla cartella di lavoro se il documento non è ancora salvato).
func (a *App) resolveLocal(ref string) (string, bool) {
	if ref == "" {
		return "", false
	}
	path := filepath.FromSlash(ref)
	if !filepath.IsAbs(path) {
		base := a.docDir()
		if base == "" {
			wd, err := os.Getwd()
			if err != nil {
				return "", false
			}
			base = wd
		}
		path = filepath.Join(base, path)
	}
	if !imageExts[strings.ToLower(filepath.Ext(path))] {
		return "", false
	}
	return path, true
}
