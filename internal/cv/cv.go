// Package cv is the content layer: the typed shape of cv.yaml and its loader.
// Content is data, not code — adding a role or a link is a YAML edit, never a
// template or handler change.
package cv

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

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
	return c, nil
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
