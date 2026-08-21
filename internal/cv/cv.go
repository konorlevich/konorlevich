// Package cv is the content layer: the typed shape of cv.yaml and its loader.
// Content is data, not code — adding a role or a link is a YAML edit, never a
// template or handler change.
package cv

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// UpdatedLayout is the W3C date format used by page `updated` fields and
// emitted verbatim as the sitemap's <lastmod>.
const UpdatedLayout = "2006-01-02"

// Parse decodes cv.yaml content and checks that the fields every page depends
// on are present. Callers run this at boot, so a malformed or half-filled CV is
// a startup failure rather than a broken page in production.
func Parse(data []byte) (CV, error) {
	var c CV
	if err := yaml.Unmarshal(data, &c); err != nil {
		return CV{}, fmt.Errorf("parse cv: %w", err)
	}
	required := []struct{ field, value string }{
		{"name", c.Name},
		{"tagline", c.Tagline},
		{"email", c.Email},
		{"location", c.Location},
	}
	missing := make([]string, 0, len(required))
	for _, r := range required {
		if r.value == "" {
			missing = append(missing, r.field)
		}
	}
	if len(missing) > 0 {
		return CV{}, fmt.Errorf("cv: required field(s) empty: %v", missing)
	}
	if len(c.WorkExperience) == 0 {
		return CV{}, fmt.Errorf("cv: work_experience is empty")
	}
	// Page dates feed the sitemap's <lastmod>, which crawlers only keep trusting
	// while it correlates with real content changes. A malformed date is a boot
	// failure rather than a silently wrong signal.
	for path, page := range c.Pages {
		if page.Updated == "" {
			return CV{}, fmt.Errorf("cv: pages[%q].updated is empty", path)
		}
		if _, err := time.Parse(UpdatedLayout, page.Updated); err != nil {
			return CV{}, fmt.Errorf("cv: pages[%q].updated %q is not a %s date: %w",
				path, page.Updated, UpdatedLayout, err)
		}
	}
	return c, nil
}

// Page is the per-page content metadata that lives alongside the CV itself.
//
// Updated is the date that page's *content* actually changed — not the build
// date, and never time.Now(). Touching CSS, a template, a footer link or a
// dependency is not a content change and must not bump it. Every entry sharing
// one timestamp that moves on each deploy is a lie crawlers learn to ignore,
// and Google drops lastmod from scheduling once it stops correlating with real
// edits.
type Page struct {
	Updated string `yaml:"updated"`
}

type CV struct {
	Name           string       `yaml:"name"`
	Tagline        string       `yaml:"tagline"`         // one-line positioning (hero)
	Intro          string       `yaml:"intro"`           // warm 1-2 sentence hero intro
	Summary        string       `yaml:"summary"`         // longer summary (also used for PDF)
	Location       string       `yaml:"location"`        // e.g. "Tbilisi, Georgia"
	Availability   string       `yaml:"availability"`    // e.g. "Remote or relocation"
	Languages      []string     `yaml:"languages"`       // e.g. ["English", "Russian"]
	Email          string       `yaml:"email"`           // primary contact (mailto)
	Photo          string       `yaml:"photo,omitempty"` // path under /static/img, optional
	Links          []Link       `yaml:"links"`
	Skills         []Skill      `yaml:"skills"`
	WorkExperience []Experience `yaml:"work_experience"`

	// Pages maps a request path to that page's content metadata. Keyed by the
	// path the renderer serves it at, so adding an indexable page means adding
	// a data entry, not touching code.
	Pages map[string]Page `yaml:"pages"`
}

type Link struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type Skill struct {
	Category string   `yaml:"category"`
	Items    []string `yaml:"items"`
}

type Experience struct {
	Company      string   `yaml:"company"`
	Role         string   `yaml:"role"`
	From         string   `yaml:"from"`         // In YYYY-MM-DD format
	To           string   `yaml:"to,omitempty"` // End date, optional
	Skills       []string `yaml:"skills"`
	Achievements []string `yaml:"achievements"`
}
