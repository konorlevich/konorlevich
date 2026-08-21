# Build Tasks: Personal Webpage — Petr Travkin

Generated from: `.design/personal-webpage/DESIGN_BRIEF.md` (+ `INFORMATION_ARCHITECTURE.md`, `DESIGN_TOKENS.css`)
Date: 2026-07-04
Stack: Go `html/template` + `cv.yaml`, extending the existing `cv_page` project.

> Each task is a vertical slice (structure + styling + interaction together), independently buildable and verifiable against the brief. Work top-to-bottom; check off after confirming with the designer.

## Context from codebase scan (read before starting)
- **Routes already exist** in `cv_page/main.go`: `GET /` & `GET /cv` (HTML), `GET /cv/download` (PDF via `fpdf`), `POST /contact` (stub).
- **CV data model**: `internal/cv/cv.go` → `Name, Summary, Links, Skills, WorkExperience`. **Missing**: `Projects`, `Location`, `Languages`, `Availability`.
- **Template**: `cv_template.html` — unstyled semantic HTML, re-parsed on every request. No static-asset serving.
- **CV/PDF resolved**: the downloadable CV = `GET /cv/download`, generated from `cv.yaml`. Page & PDF share one source. → "Download CV" links to `/cv/download`.
- **Contact form**: `POST /contact` stub is **OUT OF SCOPE for v1** — primary CTA is `mailto:` (brief). Leave the stub untouched/unused.

---

## Foundation

- [x] **Extend CV data model + `cv.yaml` content**: Add to `internal/cv/cv.go` a `Projects []Project` (`Name, URL, Description, Screenshot, Tag`) and top-level `Location`, `Languages []string`, `Availability` fields; populate `cv.yaml` with real content (Tbilisi; Remote/relocation; EN/RU; the 3 projects; a humanized first-person summary). "Done" = `go build` passes and the new fields render in a quick template dump. _Modifies `internal/cv/cv.go` + `cv.yaml`._

- [x] **Static-asset pipeline + base stylesheet (establishes WARM MINIMALISM)**: Serve a `/static/` path (CSS, fonts, images) from `main.go`; add `tokens.css` (copy of `DESIGN_TOKENS.css`) + a `base.css` with reset, base element styles, page background (`--color-bg-primary`), body type (Inter, `--font-size-base`), and the `--max-width-page` container. Load Fraunces / Inter / JetBrains Mono — **self-host preferred** (privacy + no external dependency; render-blocking-safe with `font-display: swap`); Google Fonts link acceptable fallback. "Done" = the page renders on warm paper with the correct fonts and one visible token-driven element. _New; consumes `DESIGN_TOKENS.css`. This task locks the visual direction — validate the aesthetic here before building further._

- [x] **Parse templates once at startup**: Move `template.ParseFiles` out of the per-request handler into a package-level parse (fail fast on boot). Small but prevents a foot-gun as the template grows. _Modifies `main.go`._

## Core UI

- [x] **Hero + primary Email CTA (RISK-FIRST + top visual priority)**: Build the hero — name (Fraunces, fluid `clamp()`), one-line positioning, 1–2 sentence warm intro, and the prominent **Email me** `mailto:` CTA. Design the **optional photo** with a strong non-photo fallback (initials mark / warm typographic hero) so it works either way — this resolves the open photo question. "Done" = hero reads warm + human, CTA is obvious above the fold, and it looks intentional with photo removed. _New. Build first among sections so the aesthetic + the hardest layout call (photo-or-not) surface early._

- [x] **Top bar / nav**: Minimal header — name/mark left; `Experience`, `Work`, `CV` jump links + distinct **Email me** button right. Anchor links scroll to sections. Mobile: collapse to name/mark + persistent Email CTA (no hamburger). _New; shared component. Depends on: hero (for anchor targets) — can be built alongside._

- [x] **Facts row**: Quiet chip row under the hero — Tbilisi, Georgia · Remote or relocation · English & Russian (fluent). Uses `--chip-*` tokens; availability chip may use the muted teal secondary accent. _New; depends on data model._

- [x] **Experience section (recruiter door)**: Restyle `WorkExperience` into a scannable, non-jargon timeline — company + role + dates legible at a glance (big names first), 1–2 humanized achievement lines, mono dates, skill tags as quiet secondary. Include a **Download CV** link → `/cv/download`. "Done" = a non-technical reader grasps seniority in ~20s. _Modifies `cv_template.html` markup for this data._

- [x] **Work with me / Projects section (freelance door)**: Section intent line + 3 project cards (title, one-liner, **live link opens new tab**, optional screenshot) for kidsspace.ge, ai-news.ge, tgchathelperbot.tech; closes with a soft freelance nudge + Email CTA. Cards use `--card-*` tokens. _New; depends on data model `Projects`._

- [x] **Contact / footer**: Repeated prominent **Email me** CTA, profile links (Email, LinkedIn, GitHub) from `Links`, `Download CV`, and a "Created by Petr Travkin" line (consistency anchor for referral traffic). Semantic `<footer>`. _New._

## Interactions & States

- [x] **Interactive states for CTA, links & cards**: Hover / focus / active for the Email CTA (all instances), nav links, and project cards. Visible focus ring via `--shadow-focus`. Optional `mailto:` prefilled subject (e.g. `?subject=Hi Petr`). Covers: hover, focus-visible, active, keyboard operation. _Applies across components._

- [x] **Edge / empty states in template**: Ensure Go template conditionals are robust — hero without a photo, project without a screenshot, role with no end date (→ "Present"), long summary wrapping, missing optional field. Covers: empty, optional, overflow. _Modifies `cv_template.html`._

## Responsive & Polish

- [x] **Responsive pass**: Hero stacks (mobile) → side-by-side (≥`md` 768px); facts chips wrap; projects 1-up → 2/3-up on wider screens; top bar collapses to name + Email CTA on mobile; fluid hero name. Breakpoints: sm 375 / md 768 / lg 1024. _Cross-component._

- [x] **Accessibility pass**: WCAG **AA** contrast verify (especially terracotta CTA text + link color on paper); semantic landmarks (`header`/`main`/`footer`), correct heading order; descriptive link text (no "click here"); meaningful/empty `alt` on photo; visible focus everywhere; honor `prefers-reduced-motion`; touch targets ≥ 44px. Checks pulled from brief → Accessibility Requirements. _Cross-component._

- [x] **Meta / shareability / perf polish**: Page `<title>` + meta description; favicon; **Open Graph + Twitter card** tags (recruiters & referrals share the link — the preview must look intentional); preload self-hosted fonts; verify fast first paint. _Modifies `cv_template.html` head + `main.go` static serving._

- [ ] **(Optional) Align PDF with new fields**: The `/cv/download` `fpdf` output is currently plain and omits `Projects`/`Location`/`Languages`. Optionally extend it so the PDF matches the page's content. _Modifies `main.go` `servePDF`. Low priority — page is the primary artifact._

## Review

- [ ] **Design review**: Run `/design-review` against the brief once the page is built and running (captures responsive + interactive-state screenshots into `.design/personal-webpage/screenshots/`).
