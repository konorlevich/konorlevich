package web

import (
	"fmt"
	"io/fs"
	"maps"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/konorlevich/konorlevich/internal/cv"
	"github.com/konorlevich/konorlevich/internal/render"
	"github.com/konorlevich/konorlevich/internal/site"
)

// Options configures the site handler. The two filesystems come from //go:embed
// in package main, so the running binary carries everything it serves.
type Options struct {
	Static    fs.FS // the static/ subtree
	Templates fs.FS // the tree containing templates/
	CV        cv.CV
	Site      site.Config
	Log       logrus.FieldLogger
	BuildTime time.Time
}

// Server holds every response the site can produce, prepared at boot.
type Server struct {
	blobs    map[string]*Blob // exact-path responses (pages + ops + documents)
	assets   *Assets
	notFound *Blob
	log      logrus.FieldLogger

	// documentNames carries the download filename for /cv/download*.
	documentNames map[string]string
}

// New builds every page, document and ops file up front. Any failure here is a
// startup failure: a broken template or a missing asset can never reach a user.
func New(opts Options) (*Server, error) {
	assets, err := LoadAssets(opts.Static, opts.BuildTime)
	if err != nil {
		return nil, fmt.Errorf("load assets: %w", err)
	}

	renderer := NewRenderer(assets, opts.Site, opts.CV)
	pages, sitemapPaths, err := renderer.RenderAll(opts.Templates, opts.BuildTime)
	if err != nil {
		return nil, err
	}

	ops, err := buildOps(opts.Site, opts.CV, sitemapPaths, assets, opts.BuildTime)
	if err != nil {
		return nil, err
	}

	s := &Server{
		blobs:         make(map[string]*Blob),
		assets:        assets,
		log:           opts.Log,
		documentNames: make(map[string]string),
	}
	maps.Copy(s.blobs, pages)
	maps.Copy(s.blobs, ops)

	notFound, ok := pages[NotFoundPath]
	if !ok {
		return nil, fmt.Errorf("web: 404 page was not rendered")
	}
	s.notFound = notFound
	delete(s.blobs, NotFoundPath) // reachable only via the catch-all, never by URL

	// Documents are pure functions of the CV, so they are rendered once too.
	pdf, err := render.PDF(opts.CV)
	if err != nil {
		return nil, err
	}
	s.blobs["/cv/download"] = NewBlob("application/pdf", pdf, htmlCacheControl, opts.BuildTime)
	s.documentNames["/cv/download"] = render.Filename(opts.CV, "pdf")

	s.blobs["/cv/download.md"] = NewBlob("text/markdown; charset=utf-8",
		render.Markdown(opts.CV), htmlCacheControl, opts.BuildTime)
	s.documentNames["/cv/download.md"] = render.Filename(opts.CV, "md")

	return s, nil
}

// Handler returns the routed handler for the whole site.
func (s *Server) Handler() *http.ServeMux {
	mux := http.NewServeMux()

	// {$} anchors the pattern to exactly "/" — without it this would swallow
	// every unmatched path and answer 200 for URLs that do not exist.
	mux.HandleFunc("GET /{$}", s.serveBlob("/"))
	mux.HandleFunc("GET /privacy", s.serveBlob("/privacy"))

	// /cv was an alias for the home page; a permanent redirect keeps the old
	// link working while leaving exactly one indexable URL for the content.
	mux.HandleFunc("GET /cv", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	})

	mux.HandleFunc("GET /cv/download", s.serveDocument("/cv/download"))
	mux.HandleFunc("GET /cv/download.md", s.serveDocument("/cv/download.md"))

	for _, p := range []string{"/robots.txt", "/sitemap.xml", "/llms.txt", "/site.webmanifest"} {
		mux.HandleFunc("GET "+p, s.serveBlob(p))
	}

	// The ICO must live at the site root — that is where every browser and
	// crawler looks for it before reading any markup.
	mux.HandleFunc("GET /favicon.ico", s.serveAssetAt("/static/favicon.ico"))

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	mux.HandleFunc("GET /static/", s.serveStatic)

	// Catch-all: anything not matched above is a real 404.
	mux.HandleFunc("GET /", s.serveNotFound)

	return mux
}

func (s *Server) serveBlob(path string) http.HandlerFunc {
	blob, ok := s.blobs[path]
	if !ok {
		// Registering a route with no body is a programming error; fail loudly
		// on the first request rather than serving something misleading.
		return func(w http.ResponseWriter, r *http.Request) {
			s.log.WithField("path", path).Error("web: no prepared response for route")
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		blob.Serve(w, r, http.StatusOK)
	}
}

func (s *Server) serveDocument(path string) http.HandlerFunc {
	blob := s.blobs[path]
	name := s.documentNames[path]
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", name))
		blob.Serve(w, r, http.StatusOK)
	}
}

func (s *Server) serveAssetAt(assetPath string) http.HandlerFunc {
	blob, ok := s.assets.Blob(assetPath)
	if !ok {
		return s.serveNotFound
	}
	return func(w http.ResponseWriter, r *http.Request) {
		blob.Serve(w, r, http.StatusOK)
	}
}

func (s *Server) serveStatic(w http.ResponseWriter, r *http.Request) {
	blob, ok := s.assets.Blob(r.URL.Path)
	if !ok {
		s.serveNotFound(w, r)
		return
	}
	blob.Serve(w, r, http.StatusOK)
}

// serveNotFound returns the designed 404 page with a real 404 status, so
// crawlers drop the URL instead of indexing a soft error.
func (s *Server) serveNotFound(w http.ResponseWriter, r *http.Request) {
	s.notFound.Serve(w, r, http.StatusNotFound)
}
