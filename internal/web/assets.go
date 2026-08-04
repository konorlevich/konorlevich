package web

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"mime"
	"path"
	"strings"
	"time"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
)

// immutableCacheControl is safe because every asset URL carries a content hash:
// a changed file yields a new URL, so the old one can be cached forever.
const immutableCacheControl = "public, max-age=31536000, immutable"

// Assets is the boot-time asset store. Every file under the embedded static/
// tree is minified (CSS/JS), hashed, precompressed and kept in memory; requests
// only ever pick a buffer and write it.
//
// The minifier runs here rather than in a pre-commit hook so the served bytes
// can never be stale relative to the source — the same guarantee the hook was
// there to provide, enforced by construction instead of by discipline.
type Assets struct {
	blobs map[string]*Blob  // "/static/css/styles.css" -> blob
	urls  map[string]string // "css/styles.css" -> "/static/css/styles.css?v=abc12345"
	texts map[string]string // "css/styles.css" -> minified source, for inlining
}

// LoadAssets walks fsys (rooted at the static/ directory) and builds the store.
func LoadAssets(fsys fs.FS, buildTime time.Time) (*Assets, error) {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/javascript", js.Minify)

	a := &Assets{
		blobs: make(map[string]*Blob),
		urls:  make(map[string]string),
		texts: make(map[string]string),
	}

	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		body, err := fs.ReadFile(fsys, p)
		if err != nil {
			return fmt.Errorf("read asset %s: %w", p, err)
		}

		contentType := contentTypeOf(p)
		if minified, err := minifyIfText(m, contentType, body); err != nil {
			return fmt.Errorf("minify %s: %w", p, err)
		} else if minified != nil {
			body = minified
			a.texts[p] = string(body)
		}

		sum := sha256.Sum256(body)
		version := hex.EncodeToString(sum[:])[:8]

		urlPath := "/static/" + p
		a.blobs[urlPath] = NewBlob(contentType, body, immutableCacheControl, buildTime)
		a.urls[p] = urlPath + "?v=" + version
		return nil
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

// minifyIfText returns the minified body for CSS/JS, or nil for anything else.
func minifyIfText(m *minify.M, contentType string, body []byte) ([]byte, error) {
	mediaType, _, _ := strings.Cut(contentType, ";")
	switch strings.TrimSpace(mediaType) {
	case "text/css", "text/javascript":
		return m.Bytes(mediaType, body)
	}
	return nil, nil
}

// contentTypeOf maps a path to its media type, registering the two types Go's
// table gets wrong or misses for our purposes.
func contentTypeOf(p string) string {
	switch path.Ext(p) {
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "text/javascript; charset=utf-8"
	case ".woff2":
		return "font/woff2"
	case ".webmanifest":
		return "application/manifest+json"
	}
	if ct := mime.TypeByExtension(path.Ext(p)); ct != "" {
		return ct
	}
	return "application/octet-stream"
}

// URL returns the content-hashed public URL for an asset path relative to
// static/, e.g. URL("css/styles.css"). Unknown paths fall back to the plain
// path so a typo shows up as a 404 rather than a silent blank.
func (a *Assets) URL(rel string) string {
	if u, ok := a.urls[rel]; ok {
		return u
	}
	return "/static/" + rel
}

// Text returns the minified source of a CSS/JS asset, for inlining into a page.
func (a *Assets) Text(rel string) (string, bool) {
	t, ok := a.texts[rel]
	return t, ok
}

// Blob looks up a prepared asset by its request path ("/static/css/x.css").
func (a *Assets) Blob(urlPath string) (*Blob, bool) {
	b, ok := a.blobs[urlPath]
	return b, ok
}
