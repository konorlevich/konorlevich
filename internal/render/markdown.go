package render

import (
	"fmt"
	"strings"

	"github.com/konorlevich/konorlevich/internal/cv"
)

// Markdown renders the CV as a Markdown document.
func Markdown(c cv.CV) []byte {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s\n\n", c.Name)

	if c.Tagline != "" {
		fmt.Fprintf(&b, "_%s_\n\n", c.Tagline)
	}

	if meta := strings.ReplaceAll(metaLine(c), "  ·  ", " · "); meta != "" {
		fmt.Fprintf(&b, "%s\n\n", meta)
	}

	if c.Summary != "" {
		fmt.Fprintf(&b, "%s\n\n", c.Summary)
	}

	if len(c.Links) > 0 {
		b.WriteString("## Links\n\n")
		for _, link := range c.Links {
			display := strings.TrimPrefix(link.URL, "mailto:")
			fmt.Fprintf(&b, "- **%s:** [%s](%s)\n", link.Name, display, link.URL)
		}
		b.WriteString("\n")
	}

	if len(c.Skills) > 0 {
		b.WriteString("## Skills\n\n")
		for _, s := range c.Skills {
			fmt.Fprintf(&b, "- **%s:** %s\n", s.Category, strings.Join(s.Items, ", "))
		}
		b.WriteString("\n")
	}

	if len(c.WorkExperience) > 0 {
		b.WriteString("## Work Experience\n\n")
		for _, exp := range c.WorkExperience {
			fmt.Fprintf(&b, "### %s — %s\n\n", exp.Company, exp.Role)
			fmt.Fprintf(&b, "_%s – %s_\n\n", exp.From, until(exp.To))
			if len(exp.Skills) > 0 {
				fmt.Fprintf(&b, "**Skills:** %s\n\n", strings.Join(exp.Skills, ", "))
			}
			if len(exp.Achievements) > 0 {
				b.WriteString("**Achievements:**\n\n")
				for _, achievement := range exp.Achievements {
					fmt.Fprintf(&b, "- %s\n", achievement)
				}
				b.WriteString("\n")
			}
		}
	}

	return []byte(strings.TrimRight(b.String(), "\n") + "\n")
}
