package web

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"strings"
	"time"

	"github.com/konorlevich/konorlevich/internal/cv"
	"github.com/konorlevich/konorlevich/internal/site"
)

// htmlCacheControl keeps HTML revalidating so a redeploy is visible at once,
// while the ETag still makes the repeat request a 304.
const htmlCacheControl = "no-cache"

// PageData is the template context for every page. cv.CV is embedded so
// templates keep reading .Name / .Links directly.
type PageData struct {
	cv.CV
	Site        site.Config
	Canonical   string
	Title       string
	Description string
	Robots      string // empty means index,follow (the default)
	OGType      string
	IsHome      bool
}

// pageSpec describes one server-rendered page.
type pageSpec struct {
	path        string // request path, also the sitemap entry
	file        string // template file under templates/pages
	title       string
	description string
	robots      string
	ogType      string
	isHome      bool
	inSitemap   bool
}

// Renderer builds every static page once at boot.
type Renderer struct {
	assets *Assets
	site   site.Config
	cv     cv.CV
	funcs  template.FuncMap
}

// NewRenderer wires the template helpers against the loaded assets.
func NewRenderer(assets *Assets, cfg site.Config, content cv.CV) *Renderer {
	r := &Renderer{assets: assets, site: cfg, cv: content}
	r.funcs = template.FuncMap{
		// asset returns a content-hashed URL for a file under static/.
		"asset": assets.URL,
		// absURL turns a site-relative path into an absolute one.
		"absURL": cfg.URL,
		// inlineCSS embeds a minified stylesheet directly in <style>.
		"inlineCSS": func(rel string) (template.CSS, error) {
			text, ok := assets.Text(rel)
			if !ok {
				return "", fmt.Errorf("inlineCSS: unknown asset %q", rel)
			}
			return template.CSS(text), nil
		},
		// formatDate turns "2006-01-02" into "Jan 2006"; passes through on error.
		"formatDate": func(s string) string {
			if s == "" {
				return "Present"
			}
			t, err := time.Parse("2006-01-02", s)
			if err != nil {
				return s
			}
			return t.Format("Jan 2006")
		},
		// initials returns up to two uppercase initials from a full name.
		"initials": func(name string) string {
			var b strings.Builder
			for p := range strings.FieldsSeq(name) {
				b.WriteString(strings.ToUpper(p[:1]))
				if b.Len() >= 2 {
					break
				}
			}
			return b.String()
		},
		// join concatenates a string slice with ", ".
		"join": func(items []string) string { return strings.Join(items, ", ") },
		// webpOf swaps a raster path for its .webp sibling (for <picture>).
		"webpOf": func(p string) string {
			for _, ext := range []string{".jpg", ".jpeg", ".png", ".JPG", ".JPEG", ".PNG"} {
				if before, ok := strings.CutSuffix(p, ext); ok {
					return before + ".webp"
				}
			}
			return p
		},
	}
	return r
}

// specs are the HTML pages this site serves.
func (r *Renderer) specs() []pageSpec {
	return []pageSpec{
		{
			path:        "/",
			file:        "home.html",
			title:       r.cv.Name + " — Software Engineer",
			description: r.cv.Name + " — software engineer with 8+ years in product and infrastructure teams. Based in " + r.cv.Location + ". " + r.cv.Availability + ".",
			ogType:      "website",
			isHome:      true,
			inSitemap:   true,
		},
		{
			path:        "/privacy",
			file:        "privacy.html",
			title:       "Privacy & Cookies — " + r.cv.Name,
			description: "How " + r.cv.Name + "'s site uses cookies and Google Analytics, and how to control it.",
			// noindex,follow: a thin legal page shouldn't be indexed, but crawlers
			// should still follow its links back into real content.
			robots:    "noindex,follow",
			ogType:    "website",
			inSitemap: false,
		},
		{
			path:        NotFoundPath,
			file:        "notfound.html",
			title:       "Page not found — " + r.cv.Name,
			description: "That page isn't here.",
			robots:      "noindex,follow",
			ogType:      "website",
			inSitemap:   false,
		},
	}
}

// NotFoundPath is the internal key for the 404 body; it is never routed.
const NotFoundPath = "/__404"

// RenderAll parses each page against the shared layout and renders it to a
// precompressed blob. Parsing and rendering both happen once, at boot, so a
// broken template is a startup failure and no request ever executes a template.
func (r *Renderer) RenderAll(fsys fs.FS, buildTime time.Time) (map[string]*Blob, []string, error) {
	pages := make(map[string]*Blob)
	var sitemap []string

	for _, spec := range r.specs() {
		// One template set per page: layout + partials + that page's content,
		// so two pages defining "content" can never collide.
		tmpl, err := template.New("layout").Funcs(r.funcs).ParseFS(fsys,
			"templates/layout.html",
			"templates/partials/*.html",
			"templates/pages/"+spec.file,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("parse page %s: %w", spec.path, err)
		}

		canonicalPath := spec.path
		if spec.path == NotFoundPath {
			canonicalPath = "/"
		}

		data := PageData{
			CV:          r.cv,
			Site:        r.site,
			Canonical:   r.site.URL(canonicalPath),
			Title:       spec.title,
			Description: spec.description,
			Robots:      spec.robots,
			OGType:      spec.ogType,
			IsHome:      spec.isHome,
		}

		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, "layout", data); err != nil {
			return nil, nil, fmt.Errorf("render page %s: %w", spec.path, err)
		}

		pages[spec.path] = NewBlob("text/html; charset=utf-8", buf.Bytes(), htmlCacheControl, buildTime)
		if spec.inSitemap {
			sitemap = append(sitemap, spec.path)
		}
	}
	return pages, sitemap, nil
}
