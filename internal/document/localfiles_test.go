package document

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalFilesServesOnlyImages(t *testing.T) {
	dir := t.TempDir()
	doc := filepath.Join(dir, "doc.md")
	os.WriteFile(doc, []byte("x"), 0o644)
	os.MkdirAll(filepath.Join(dir, "img"), 0o755)
	os.WriteFile(filepath.Join(dir, "img", "a b.png"), []byte("\x89PNG\r\n\x1a\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "secret.txt"), []byte("segreto"), 0o644)

	d := &Document{}
	if _, err := d.Open(doc); err != nil {
		t.Fatal(err)
	}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusTeapot) })
	handler := d.LocalFiles(next)

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
