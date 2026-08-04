// Package render turns the CV content into downloadable documents. Both
// renderers are pure functions of cv.CV, so the service builds them once at
// boot and never executes a renderer on the request path.
package render

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"

	"github.com/konorlevich/konorlevich/internal/cv"
)

// Brand colours reused from the design tokens, so the document matches the site.
var (
	textSecondary = [3]int{107, 99, 87} // warm secondary
	accent        = [3]int{168, 72, 42} // terracotta
	black         = [3]int{0, 0, 0}     //
)

// PDF renders the CV as an A4 PDF document.
func PDF(c cv.CV) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	// Translate UTF-8 content into the core font's cp1252 encoding so characters
	// like em dashes and × render correctly rather than as mojibake.
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	setColor := func(c [3]int) { pdf.SetTextColor(c[0], c[1], c[2]) }

	// Name
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, tr(c.Name))
	pdf.Ln(9)

	// Location · Availability · Languages — one quiet meta line
	if meta := metaLine(c); meta != "" {
		pdf.SetFont("Arial", "", 11)
		setColor(textSecondary)
		pdf.MultiCell(190, 6, tr(meta), "", "L", false)
		setColor(black)
	}

	// Summary paragraph
	if c.Summary != "" {
		pdf.Ln(1)
		pdf.SetFont("Arial", "", 11)
		pdf.MultiCell(190, 6, tr(c.Summary), "", "L", false)
	}
	pdf.Ln(4)

	// Links — rendered as clickable hyperlinks
	if len(c.Links) > 0 {
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(40, 10, "Links")
		pdf.Ln(8)
		pdf.SetFont("Arial", "", 12)
		for _, link := range c.Links {
			label := tr(link.Name + ": ")
			pdf.CellFormat(pdf.GetStringWidth(label)+1, 8, label, "", 0, "L", false, 0, "")
			// Display without the mailto: scheme, but keep it as the link target.
			display := strings.TrimPrefix(link.URL, "mailto:")
			setColor(accent)
			pdf.SetFont("Arial", "U", 12)
			pdf.CellFormat(0, 8, tr(display), "", 1, "L", false, 0, link.URL)
			pdf.SetFont("Arial", "", 12)
			setColor(black)
		}
	}

	// Work experience
	if len(c.WorkExperience) > 0 {
		pdf.Ln(2)
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(40, 10, "Work Experience")
		pdf.Ln(10)

		for _, exp := range c.WorkExperience {
			pdf.SetFont("Arial", "B", 12)
			pdf.Cell(40, 8, tr(fmt.Sprintf("%s - %s", exp.Company, exp.Role)))
			pdf.Ln(6)
			pdf.SetFont("Arial", "", 12)
			pdf.Cell(40, 8, tr(fmt.Sprintf("From: %s to %s", exp.From, until(exp.To))))
			pdf.Ln(8)

			if len(exp.Skills) > 0 {
				pdf.Cell(40, 8, tr(fmt.Sprintf("Skills: %s", strings.Join(exp.Skills, ", "))))
				pdf.Ln(8)
			}

			if len(exp.Achievements) > 0 {
				pdf.Cell(40, 8, "Achievements:")
				pdf.Ln(6)
				for _, achievement := range exp.Achievements {
					pdf.MultiCell(190, 6, tr(fmt.Sprintf("- %s", achievement)), "", "", false)
				}
			}
			pdf.Ln(8)
		}
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("render pdf: %w", err)
	}
	return buf.Bytes(), nil
}

// Filename returns a download filename for the CV in the given extension.
func Filename(c cv.CV, ext string) string {
	return strings.ReplaceAll(c.Name, " ", "-") + "-CV." + ext
}

// metaLine joins location, availability and languages into one line.
func metaLine(c cv.CV) string {
	meta := make([]string, 0, 3)
	if c.Location != "" {
		meta = append(meta, c.Location)
	}
	if c.Availability != "" {
		meta = append(meta, c.Availability)
	}
	if len(c.Languages) > 0 {
		meta = append(meta, strings.Join(c.Languages, ", "))
	}
	return strings.Join(meta, "  ·  ")
}

func until(to string) string {
	if to == "" {
		return "Present"
	}
	return to
}
