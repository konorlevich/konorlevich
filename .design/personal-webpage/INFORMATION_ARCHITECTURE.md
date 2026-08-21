# Information Architecture: Personal Webpage — Petr Travkin

> Extends `DESIGN_BRIEF.md` (feature: `personal-webpage`). Single scrollable page, Go `html/template` + `cv.yaml`.
> Structural principle from the brief: **one warm entrance, two clear doors, one action (email).**

## Site Map

Deliberately flat — this is a one-page site. "Pages" below the root are anchor sections on the same document, not routes.

- Home `/`
  - Hero `#top` (default view on load)
  - Facts row `/#facts` (location · availability · languages) — visually part of hero
  - Experience `/#experience` — recruiter door
  - Work with me `/#work` — freelance / referral door
  - Contact `/#contact` — footer, repeats email CTA + profile links
- Downloadable CV `/cv.pdf` — separate static asset (generated from `cv.yaml`, see brief open question)

No other routes. No About/Contact/Projects pages. (See brief "Out of Scope": no multi-page site.)

## Navigation Model

- **Primary navigation**: A minimal top bar with the name/mark on the left and **max 4** jump links + the email CTA on the right: `Experience`, `Work`, `CV`, and a visually distinct **Email me** button. Anchor links scroll to the matching section.
- **Secondary navigation**: None. Within sections, movement is by scroll. The repeated email CTA (hero + contact, optionally sticky) is the only in-page "navigation" that matters.
- **Utility navigation**: Profile links (Email, LinkedIn, GitHub) live in the footer/contact section, not the top bar, to keep the header focused on the primary action. `CV` download appears in both header and contact.
- **Mobile navigation**: The top bar collapses to name/mark + a single persistent **Email me** button (the jump links are lower-value on a short page and can drop, or collapse into a light menu). No hamburger required for 4 anchors — prefer a persistent CTA over hidden nav so the primary action is never buried.

## Content Hierarchy

### Hero (`#top`)
1. **Name + one-line positioning** — Who he is in one warm, non-jargon sentence. First thing both audiences must grasp.
2. **Primary email CTA** — The single action; visible without scrolling.
3. **Short warm intro (1–2 sentences)** — Human framing ("I build reliable software / I help people ship").
4. **Optional photo / avatar fallback** — Reinforces "warm & human"; hero works with or without it.

### Facts row (`#facts`)
1. **Location** — Tbilisi, Georgia. — Recruiter's first filter.
2. **Availability** — Remote or reasonable relocation. — Removes the biggest disqualifier immediately.
3. **Languages** — English & Russian (fluent). — International reassurance.
   *(Quiet, chip-style; supports the hero, doesn't compete with the CTA.)*

### Experience (`#experience`) — recruiter door
1. **Section intent line** — "8+ years building and shipping software internationally." — Frames the list for a non-technical reader.
2. **Company + role + dates** — Big names (Bumble, Badoo) legible at a glance; seniority obvious.
3. **1–2 humanized achievement lines per role** — Outcome-first, jargon-light.
4. **Skill tags (secondary)** — Quiet, supporting; never the headline.
5. **Download CV link** — For anyone wanting the full document.

### Work with me (`#work`) — freelance / referral door
1. **Section intent line** — "Things I've built for people (and myself)." — Speaks directly to the "Created by" visitor.
2. **Project cards (3)** — Kids Space, AI News, Chat Structure Helper — each: title, one-liner, live link, optional screenshot.
3. **Soft freelance nudge + email CTA** — "Want something like this? Let's talk." → email.

### Contact / footer (`#contact`)
1. **Repeated primary email CTA** — Last, most deliberate chance to act.
2. **Profile links** — Email, LinkedIn, GitHub.
3. **CV download** — Repeated.
4. **"Created by" line / copyright** — Consistency anchor for referral traffic.

## User Flows

### Flow A — Recruiter skim ("is this person worth contacting?")
1. User lands on `/` from GitHub/LinkedIn → sees **Hero** (name, positioning, email CTA).
2. User reads the **Facts row** — location, remote/relocation, languages.
   - If availability/location is a dealbreaker → user leaves (good: qualified out fast, no wasted contact).
   - If it fits → continues.
3. User scans **Experience** — recognizes Bumble/Badoo, senior roles, clear outcomes.
4. User takes action:
   - **Email me** (primary intended outcome), OR
   - **Download CV** for the full document to share internally, OR
   - Clicks GitHub/LinkedIn to verify depth, then returns to email.
5. User arrives at: an open, pre-framed email to Petr.

### Flow B — Referral proof path ("can they build my thing?")
1. User lands on `/` from a "Created by Petr Travkin" signature on a site he built → sees **Hero**.
2. User is already warm; jumps toward **Work with me** (via `Work` nav link or scroll).
3. User sees **project cards**; opens a live site (`kidsspace.ge` / `ai-news.ge` / `tgchathelperbot.tech`) in a new tab.
   - If the live work convinces them → returns to the page.
   - If not → leaves.
4. User hits the **soft freelance nudge + email CTA**.
5. User takes action: **Email me** with a project inquiry.
6. User arrives at: an open email to Petr.

### Flow C — Fast contact (either audience, high intent)
1. User lands on `/`.
2. Email CTA is visible in hero without scrolling.
3. User clicks **Email me** immediately → mail client opens. (No form, no backend — brief constraint.)

## Naming Conventions

| Concept | Label in UI | Notes |
|---------|-------------|-------|
| Employment history | **Experience** | Warmer and shorter than "Work Experience"/"Career"; recruiter-familiar. |
| Freelance / project proof | **Work with me** | Action-framed, speaks to the referral visitor's intent. Section anchor short label: `Work`. |
| Individual featured build | **Project** | Consistent across cards; avoid mixing "case study"/"portfolio item". |
| Primary contact action | **Email me** | One verb-first label used on every CTA. Never "Contact"/"Get in touch"/"Reach out" mixed in. |
| Downloadable résumé | **CV** | Use "CV" everywhere (not "resume"/"résumé") for consistency; matches existing `cv_page`. |
| Availability status | **Remote or relocation** | Plain phrasing; avoid "open to work" badge clichés. |

## Component Reuse Map

| Component | Used on | Behavior differences |
|-----------|---------|----------------------|
| Page shell / container | All sections | Single max-width column; consistent horizontal padding. |
| Top bar (name/mark + nav + CTA) | Global (sticky optional) | Full jump links on desktop; collapses to name + Email CTA on mobile. |
| Primary email CTA | Hero, Work, Contact, Top bar | Same component/style everywhere; may render as prominent button (hero/contact) vs. compact button (top bar). |
| Section wrapper (heading + intent line + body) | Experience, Work, Contact | Same rhythm/spacing; differing inner content. |
| Card | Work (project cards) | Only place cards appear; grid density changes by breakpoint. |
| Fact chip | Facts row | Repeated for location/availability/languages; identical style. |
| Skill tag | Experience (per role), (optional) skills block | Quiet, secondary styling; never primary. |
| Footer/contact block | Contact | Reuses CTA + link list; adds "Created by" line. |

## Content Growth Plan

This is an evergreen page; growth is slow and edited-in, not published via a system.

- **Experience** — grows as roles are added by editing `cv.yaml` `work_experience`. Renders newest-first; no pagination needed at this volume. If it ever grows long, older roles can collapse to a compact list.
- **Work with me / Projects** — grows by adding entries to a new `projects` list in `cv.yaml`. Keep to ~3 *featured*; if it grows, feature the best 3–4 and (optionally) link the rest as a plain list. No filtering/search — curation over accumulation.
- **Skills** — edited in `cv.yaml` `skills`; static.
- **CV** — regenerated from the same `cv.yaml` data (target), so the page and PDF never drift.
- No archive/pagination/search patterns are warranted at this scale. (Brief: no blog/CMS.)

## URL Strategy

- **Pattern**: Single route `/`; in-page destinations are fragment anchors `/#section` (`#top`, `#facts`, `#experience`, `#work`, `#contact`).
- **Dynamic segments**: None. No parameterized routes.
- **Query parameters**: None. (Optional later: `mailto:` may carry a prefilled `?subject=` — that's a mail-link param, not a page route.)
- **Static assets**: `/cv.pdf` for the downloadable CV; project screenshots as static image files under an assets path.
- **Anchors must match nav labels** so `Experience`/`Work`/`CV` links and section IDs stay in sync (see Naming Conventions).
