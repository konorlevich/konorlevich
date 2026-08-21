# Design Brief — Project Pages

_Feature slug: `project-pages` · Created 2026-07-21_

## Intent

Today the site shows projects as a single short list on the home page. This
feature turns each project into a real, linkable page — a `/projects` index plus
one `/projects/<slug>` detail page per project — so each piece of work can be
described properly and can carry its own honest achievements.

## Goals

1. **A reusable template system** for project pages (index + detail), consistent
   with the existing Warm Minimalism system — no new visual language.
2. **Rich, honest per-project content**: overview, role, what it is, and
   achievements grounded in real data.
3. **Achievements from `.results`** (Google Search Console exports) presented
   truthfully: rankings, indexed-page depth, and country reach — not inflated
   traffic claims.

## Non-goals

- No redesign of the home page beyond linking the project list into the new pages.
- No dark-mode work (out of scope per the parent brief; tokens already ship a
  future palette).
- No analytics dashboards or live data — the `.results` numbers are a point-in-time
  snapshot, rendered as static content.

## Key decisions (locked)

| Decision | Choice |
|----------|--------|
| Structure | **Index + detail** — `/projects` grid, `/projects/<slug>` detail |
| Data source | **One YAML file per project** under `projects/` — single source of truth |
| CreateInvoice | **Included** as a 5th project (own product) |
| Achievements framing | **Rankings & reach, honest** — real numbers, no inflation |
| Home short list | Now reads from the same `projects/*.yaml` and links into detail pages |

## Single source of truth

`projects/*.yaml` becomes the one place project data lives. The `projects:` block
is removed from `cv.yaml`; the home Work section, the PDF export, and the Markdown
export all read from the loader. This removes the duplication of maintaining a
project in two files.

## Aesthetic direction

Inherits **Warm Minimalism** from `.design/personal-webpage/DESIGN_BRIEF.md`:
warm paper background, one terracotta accent, Fraunces display / Inter body /
JetBrains Mono for technical detail, generous space. Detail pages read like a
short editorial case study, not a dashboard.

## Honesty constraint

The `.results` data covers roughly the first four weeks each site was live
(Jun–Jul 2026). Absolute clicks are small. The page therefore leads with what is
genuinely creditable — **search rankings, indexed content depth, and reach across
countries** — and every metrics block carries a dated source note. No "thousands
of visitors" language. Tech stack, timeline, and any unknown specifics are left
empty (the template hides empty sections) rather than fabricated.

## Open questions for the owner

- CreateInvoice: production URL and a confirmed one-line description.
- Per-project `stack`, `timeline`, and longer `overview` prose — pre-seeded where
  verifiable, otherwise left blank to fill in.
