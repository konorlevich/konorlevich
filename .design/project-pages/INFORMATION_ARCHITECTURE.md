# Information Architecture — Project Pages

_Feature slug: `project-pages` · Created 2026-07-21_

## URL map

| URL | Page | Template |
|-----|------|----------|
| `/` | Home (Work section links into projects) | `cv_template.html` |
| `/projects` | Projects index — grid of all projects | `projects_template.html` |
| `/projects/{slug}` | Project detail — one case study | `project_template.html` |

Slugs are stable and human-readable:

| Project | Slug | `.results` folder |
|---------|------|-------------------|
| AI News | `ai-news` | `ai-news` |
| CreateInvoice | `create-invoice` | `createinvoice` |
| Chat Structure Helper | `chat-structure-helper` | `tgchathelperbot` |
| Kids Space | `kids-space` | `kidsspace-3` |
| Speakadoo | `speakadoo` | `speakadoo-2` |

Unknown slug → `404`.

## Navigation

- **Top bar**: add a `Projects` link → `/projects` (present on every page).
- **Home Work section**: each card links to its `/projects/{slug}` detail page
  ("View project →"); the section footer links to `/projects` ("See all projects →").
- **Index page**: each card links to its detail page.
- **Detail page**: breadcrumb / back link to `/projects`; primary CTA "Visit site"
  (external) when a URL exists; contact CTA to the owner's email.

## Detail page structure (top → bottom)

1. **Header** — back link, project name, badge, tagline, role · timeline meta,
   "Visit site" button.
2. **Screenshot** — hero image (or typographic placeholder when none).
3. **Overview** — 1–2 short paragraphs: what it is, who it's for.
4. **Highlights** — achievement bullets (build facts + search reach).
5. **By the numbers** — stat tiles (pages indexed, countries reached, best
   ranking, etc.) with a dated Search Console source note. Hidden if no metrics.
6. **Stack** — technology tags. Hidden if empty.
7. **Contact CTA** — "Want something like this?" → email.

Empty sections (`overview`, `stack`, `metrics`) are omitted, never shown blank.

## Index page structure

- Section head: "Projects" eyebrow + short intro.
- Responsive card grid (1 → 2 → 3 columns) reusing the home `.project-card`
  pattern, each card linking to its detail page.
- Contact nudge at the foot, mirroring the home Work section.

## SEO

- Detail: `<title>` = `{Project} — {Owner}`, meta description = project summary,
  canonical `https://konorlevich.tech/projects/{slug}`, Open Graph with the
  project screenshot.
- Index: `<title>` = `Projects — {Owner}`, canonical `/projects`.
