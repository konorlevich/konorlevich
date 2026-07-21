package cv

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
