# Design Brief: Personal Webpage — Petr Travkin

> Source of truth for the design flow. Owner: Petr Travkin (hello@konorlevich.tech), Tbilisi, Georgia.
> Stack: Go `html/template` + `cv.yaml`, extending the existing `cv_page` project.

## Problem

Petr is found by two very different kinds of people, and today they land on nothing (or a raw, unstyled CV page) that speaks to neither:

- **Recruiters** click through from GitHub / LinkedIn. In ~20 seconds they need to decide "is this a real senior engineer worth contacting?" — and they are usually **non-technical**, so a wall of Kubernetes achievements doesn't land.
- **"Created by" referrals** arrive from a small signature on a site Petr built for someone. They already *like the work*; they want to answer one question — "can this person build **my** thing, and how do I reach them?"

Both groups buy on **trust and clarity**, not on technical jargon. Right now there is no page that earns that trust in the first few seconds or makes the next step obvious.

## Solution

A single, warm, human personal page that leads with the person and then splits cleanly into two doors:

1. A hero that introduces Petr as a real, approachable, reliable engineer — name, one-line positioning, and (optionally) a photo — with the primary action always in reach.
2. **My experience** — a scannable, non-jargon summary of 8+ years of international engineering (Bumble, Badoo, backend/release/infra), sized for a recruiter's quick read, with the full detail available as a downloadable CV/PDF.
3. **Work with me** — featured proof of shipped work (one client site + two of his own products), framed as "here's what I can build for you."

Every path funnels to one action: **email Petr directly.** A separate, printable/downloadable CV serves anyone who wants the complete document.

## Experience Principles

1. **Trust before tech** — Every section must read as human and credible to a non-technical visitor first; technical depth is available on a second look (GitHub, CV), never the price of entry.
2. **One warm entrance, two clear doors** — The top unifies both audiences around the person; below the fold, "experience" (recruiters) and "work with me" (freelance) are visually distinct, never blended into a confused hybrid.
3. **One action, always reachable** — Emailing Petr is the single primary CTA. It is obvious in the hero and repeated at every natural decision point; competing actions stay visually quieter.

## Aesthetic Direction

- **Philosophy**: Warm minimalism. Calm, spacious, typographic, and fast — the restraint of a well-run system, with warmth in color, copy, and (optional) photography so it reads as a *person*, not a corporate CV or a hacker terminal.
- **Tone**: Warm, confident, plainspoken. First person. Approachable and human — the register of a trusted professional you'd feel comfortable emailing, not a jargon-heavy résumé.
- **Reference points**: Clean personal sites of senior engineers who lead with clarity (e.g. the calm, typographic feel of brianlovin.com / leerob.com); the trustworthy simplicity of his own kidsspace.ge and tgchathelperbot.tech.
- **Anti-references**: Terminal/hacker green-on-black; dense corporate CV templates; over-designed SaaS landing pages with heavy gradients, animation, and "trying too hard" polish; anything that requires technical literacy to feel impressed.

## Existing Patterns

The `cv_page` project provides the **data model and render pipeline**, but the visual layer is a clean greenfield — the current template (`cv_template.html`) is raw, unstyled semantic HTML with no CSS, fonts, or tokens.

- **Typography**: None yet. Browser defaults only. → New, to be defined in Design Tokens.
- **Colors**: None yet. → New, to be defined in Design Tokens.
- **Spacing**: None yet. → New scale to be defined in Design Tokens.
- **Data model (reuse)**: `cv.yaml` → `Name`, `Summary`, `Links` (Email/LinkedIn/GitHub), `Skills` (category → items), `WorkExperience` (company, role, from/to, skills, achievements). Extend with fields for: location, languages, availability, and a `Projects` list (name, url, description, optional screenshot).
- **Render (reuse)**: Go `html/template` ranges over the YAML data; `main.go` + `internal/cv`, `internal/config`. Keep this pipeline; add styling and new sections.
- **Semantic structure (respect)**: Existing markup is clean and semantic (`h1/h2/h3`, lists) — preserve semantic HTML, layer CSS on top.

## Component Inventory

| Component | Status | Notes |
| --------- | ------ | ----- |
| Page shell / layout | New | Single scrollable page; max-width content column; responsive. |
| Header / nav | New | Minimal; name/mark + jump links (Experience, Work, Contact) + email CTA. Optional. |
| Hero | New | Name, one-line positioning, short warm intro, primary email CTA, optional photo. Location + languages + availability shown as light "facts." |
| Primary CTA (email button) | New | Reused everywhere; `mailto:` — the single most important element. |
| Facts row (location / remote / languages) | New | Recruiter reassurance: Tbilisi · Remote or relocation · EN/RU fluent. |
| Experience section | Modify | Restyle the existing work-experience data into a scannable, non-jargon timeline; big-name companies legible at a glance. |
| Experience item | Modify | Company, role, dates, 1–2 humanized achievement lines; skill tags secondary. |
| Skills display | Modify | Existing skills data, restyled as quiet tags — supporting, not headline. |
| Projects / "Work with me" section | New | Featured cards for the 3 projects; framed as freelance proof + soft "let's build yours." |
| Project card | New | Title, one-line description, live link, optional screenshot/preview. |
| CV download | New | Link/button to the printable/downloadable CV (PDF). |
| Contact / footer | New | Repeats email CTA + Links (Email, LinkedIn, GitHub). "Created by" consistency lives here. |
| Photo / avatar | New (optional) | Petr is undecided on a photo — design must work with AND without one (strong non-photo fallback: initials mark or warm typographic hero). |

## Key Interactions

- **Land → contact fast**: The email CTA is visible without scrolling. Clicking it opens the visitor's mail client via `mailto:` (pre-filled subject optional, e.g. "Hi Petr — ..."). No form, no backend.
- **Skim path (recruiter)**: Hero → facts row (location/remote/languages) → Experience (scannable) → CTA / download CV. Should be satisfying to read top-to-bottom in ~20–30s.
- **Proof path (referral)**: Hero → Work with me → project cards → outbound link to a live site (opens in new tab) → back to CTA. Each project link is a credibility payoff.
- **Feedback / states**: Hover states on CTA, links, and project cards give clear affordance; focus states are visible for keyboard users. No heavy motion — at most gentle, optional transitions consistent with "warm minimalism."

## Responsive Behavior

- **Mobile-first.** Single-column throughout; comfortable line length; generous tap targets (email CTA ≥ 44px).
- **Hero**: Photo (if used) stacks above/below text on mobile, side-by-side on ≥ tablet. Non-photo fallback centers cleanly.
- **Facts row**: Wraps to stacked chips on narrow screens.
- **Projects**: Single column on mobile → 2 (or 3) up on wider screens.
- **Experience**: Always single column; dates may move above the role on mobile for readability.
- **CTA**: Consider a persistent/repeated email CTA so it's never far away on long scrolls (esp. mobile).

## Accessibility Requirements

- **Contrast**: Body text and interactive elements meet WCAG AA (≥ 4.5:1 for text, ≥ 3:1 for large text / UI). Warmth must not cost legibility.
- **Keyboard**: All links/buttons reachable and operable by keyboard; visible focus rings; logical tab order.
- **Screen readers**: Semantic HTML preserved (headings in order, lists, landmarks: `header`/`main`/`footer`). Project links have descriptive text (not "click here"). Photo/avatar has meaningful `alt` (or empty alt if purely decorative).
- **Motion**: Respect `prefers-reduced-motion`; keep any animation subtle and optional.
- **Targets**: Interactive targets ≥ 44×44px on touch.

## Out of Scope

- **No contact form / backend for submissions** — email (`mailto:`) is the only contact channel for v1.
- **No blog / CMS / writing section** — evergreen single page; content grows by editing `cv.yaml`, not a publishing system.
- **No multi-page site / routing** — one page + a downloadable CV. No separate About/Contact routes.
- **No auth, no analytics build-out, no i18n** — page is English-only for v1 (audience reads English; RU/EN fluency is *stated*, not a language toggle).
- ~~No dark-mode requirement — single warm light theme for v1.~~ **Updated 2026-07-04 (post-review):** dark mode is now shipped as an intended feature via `prefers-color-scheme` (warm charcoal palette, not an inversion). Light remains the primary/default theme.
- **No CMS-managed project screenshots pipeline** — project images are static assets added manually.
- **Not a redesign of the featured projects themselves** — kidsspace.ge / ai-news.ge / tgchathelperbot.tech are shown as-is via link/screenshot.

---

### Featured Projects (content reference)

| Project | URL | One-liner | Why it's here |
| ------- | --- | --------- | ------------- |
| Kids Space | https://kidsspace.ge/ | Multilingual (GE/EN/RU) site for a Tbilisi kindergarten. | Real **client** work — the "Created by" proof. |
| AI ამბები (AI News) | https://ai-news.ge | Georgian-language AI news platform. | His own **shipped product** — builds & runs end to end. |
| Chat Structure Helper | https://tgchathelperbot.tech | No-code Telegram bot that auto-structures group chats (topics, permissions, templates) with a visual builder; EN/RU/ES. | Own **SaaS product** — product thinking + polish. |

### Open Questions (carry forward)

- **Photo: undecided.** Hero must be designed to work strongly with or without it. Decide at build (Phase 6).
- **CV/PDF source**: confirm whether the downloadable CV is generated from `cv.yaml` or a separate maintained PDF. (Lean: generate from the same data.)
