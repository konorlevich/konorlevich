package web

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/konorlevich/konorlevich/internal/cv"
	"github.com/konorlevich/konorlevich/internal/site"
)

// opsCacheControl lets these small, rarely-changing files be cached briefly
// while still picking up a redeploy quickly.
const opsCacheControl = "public, max-age=3600"

// buildOps renders robots.txt, sitemap.xml, llms.txt and site.webmanifest.
// They are plain data derived from the same content the pages use, so they can
// never drift from what is actually served.
func buildOps(cfg site.Config, content cv.CV, sitemapEntries []SitemapEntry, assets *Assets, buildTime time.Time) (map[string]*Blob, error) {
	out := make(map[string]*Blob, 4)

	out["/robots.txt"] = NewBlob("text/plain; charset=utf-8",
		[]byte(robotsTxt(cfg)), opsCacheControl, buildTime)

	sitemap, err := sitemapXML(sitemapEntries)
	if err != nil {
		return nil, err
	}
	out["/sitemap.xml"] = NewBlob("application/xml; charset=utf-8", sitemap, opsCacheControl, buildTime)

	out["/llms.txt"] = NewBlob("text/plain; charset=utf-8",
		[]byte(llmsTxt(cfg, content)), opsCacheControl, buildTime)

	manifest, err := webManifest(content, assets)
	if err != nil {
		return nil, err
	}
	out["/site.webmanifest"] = NewBlob("application/manifest+json", manifest, opsCacheControl, buildTime)

	return out, nil
}

func robotsTxt(cfg site.Config) string {
	var b strings.Builder
	b.WriteString("User-agent: *\n")
	b.WriteString("Allow: /\n\n")
	b.WriteString("Sitemap: " + cfg.URL("/sitemap.xml") + "\n")
	// llms.txt mirrors the page facts in Markdown for AI crawlers.
	b.WriteString("# LLM-readable summary: " + cfg.URL("/llms.txt") + "\n")
	return b.String()
}

// urlEntry / urlSet mirror the sitemaps.org schema.
type urlEntry struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

type urlSet struct {
	XMLName xml.Name   `xml:"urlset"`
	NS      string     `xml:"xmlns,attr"`
	URLs    []urlEntry `xml:"url"`
}

// sitemapXML renders the sitemap from content dates only. It deliberately takes
// no build time: <lastmod> must reflect when each page's content changed, so two
// consecutive deploys with no content edit produce identical bytes. A lastmod
// that moves on every deploy is a signal crawlers learn to ignore.
func sitemapXML(entries []SitemapEntry) ([]byte, error) {
	set := urlSet{NS: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	for _, e := range entries {
		set.URLs = append(set.URLs, urlEntry{Loc: e.Loc, LastMod: e.LastMod})
	}
	body, err := xml.MarshalIndent(set, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal sitemap: %w", err)
	}
	return append([]byte(xml.Header), append(body, '\n')...), nil
}

// llmsTxt mirrors the on-page facts as Markdown for AI discoverability. It is
// generated from the same cv.CV the HTML renders, so the two cannot disagree.
func llmsTxt(cfg site.Config, c cv.CV) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", c.Name)
	fmt.Fprintf(&b, "> %s\n\n", c.Tagline)
	fmt.Fprintf(&b, "Software engineer. %s. %s.\n\n", c.Location, c.Availability)
	if c.Summary != "" {
		fmt.Fprintf(&b, "%s\n\n", c.Summary)
	}

	b.WriteString("## Pages\n\n")
	fmt.Fprintf(&b, "- [Home](%s): full CV — experience, skills, contact.\n", cfg.URL("/"))
	fmt.Fprintf(&b, "- [CV (PDF)](%s): the same CV as a PDF document.\n", cfg.URL("/cv/download"))
	fmt.Fprintf(&b, "- [CV (Markdown)](%s): the same CV as Markdown.\n", cfg.URL("/cv/download.md"))
	fmt.Fprintf(&b, "- [Privacy](%s): cookie and analytics policy.\n\n", cfg.URL("/privacy"))

	if len(c.Skills) > 0 {
		b.WriteString("## Skills\n\n")
		for _, s := range c.Skills {
			fmt.Fprintf(&b, "- **%s:** %s\n", s.Category, strings.Join(s.Items, ", "))
		}
		b.WriteString("\n")
	}

	if len(c.WorkExperience) > 0 {
		b.WriteString("## Experience\n\n")
		for _, e := range c.WorkExperience {
			to := e.To
			if to == "" {
				to = "Present"
			}
			fmt.Fprintf(&b, "- **%s — %s** (%s – %s)\n", e.Company, e.Role, e.From, to)
		}
		b.WriteString("\n")
	}

	b.WriteString("## Contact\n\n")
	fmt.Fprintf(&b, "- Email: %s\n", c.Email)
	for _, l := range c.Links {
		if l.Name != "Email" {
			fmt.Fprintf(&b, "- %s: %s\n", l.Name, l.URL)
		}
	}
	return b.String()
}

type manifestIcon struct {
	Src     string `json:"src"`
	Sizes   string `json:"sizes"`
	Type    string `json:"type"`
	Purpose string `json:"purpose,omitempty"`
}

type manifest struct {
	Name            string         `json:"name"`
	ShortName       string         `json:"short_name"`
	StartURL        string         `json:"start_url"`
	Display         string         `json:"display"`
	BackgroundColor string         `json:"background_color"`
	ThemeColor      string         `json:"theme_color"`
	Icons           []manifestIcon `json:"icons"`
}

func webManifest(c cv.CV, assets *Assets) ([]byte, error) {
	m := manifest{
		Name:            c.Name + " — Software Engineer",
		ShortName:       c.Name,
		StartURL:        "/",
		Display:         "minimal-ui",
		BackgroundColor: "#faf7f2",
		ThemeColor:      "#faf7f2",
		Icons: []manifestIcon{
			{Src: assets.URL("icons/icon-192.png"), Sizes: "192x192", Type: "image/png"},
			{Src: assets.URL("icons/icon-512.png"), Sizes: "512x512", Type: "image/png"},
			{Src: assets.URL("icons/icon-512.png"), Sizes: "512x512", Type: "image/png", Purpose: "maskable"},
		},
	}
	body, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal manifest: %w", err)
	}
	return append(body, '\n'), nil
}
