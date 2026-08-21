# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

**Primary and only target audience: hiring readers.** Recruiters, in-house talent
partners, and engineering managers who arrive from LinkedIn or GitHub — often on
a phone, usually non-technical, mid-sourcing-session. Their job is a fast
credibility check: *is this a real senior engineer worth contacting, and how do I
reach them?* They typically give the page 20–30 seconds before deciding.

Confirmed 2026-08-04: the freelance / "can you build my thing" audience is **no
longer this site's job**. Client acquisition and project showcasing moved to
Soarline Studio. Earlier briefs under `.design/` that describe a second
"Work with me" door are historical, not current product truth.

Incidental readers to keep working, but not to design around: anyone who searches
the name, and AI crawlers (the site publishes `/llms.txt` and `/robots.txt` for
them).

## Product Purpose

`konorlevich.tech` is Petr Travkin's personal site: one page that earns a hiring
reader's trust quickly, states 10 years of engineering experience in language a
non-technical reader can judge, and makes the next step obvious.

Success is a single outcome: **the reader emails Petr** (or downloads the CV and
emails later). There is no funnel beyond that.

## Positioning

Two things a neighboring "developer portfolio" could not truthfully copy:

- **The site is the work sample.** It is a single self-contained Go binary:
  templates, static assets and content are compiled in with `//go:embed`; every
  page, the PDF, the Markdown CV and all ops files are rendered to bytes at boot,
  then Brotli/gzip precompressed with an `ETag`. A request never executes a
  template or touches disk. A broken template or missing asset fails the boot,
  never a visitor's request. The engineering claims in the CV are demonstrated by
  the artifact making them.
- **The site operates its own contact channel.** `hello@konorlevich.tech` is a
  real inbox served by this same binary: Resend receives inbound mail, POSTs an
  `email.received` webhook to `/webhooks/resend/inbound`, and the service verifies
  the signature and forwards the message with `Reply-To` set to the original
  sender.

## Operating Context

- **Arrival:** LinkedIn / GitHub profile links and name searches. Mobile-heavy.
- **Reading scene:** a sourcing pass across many candidates; skimmed, not read.
- **Content updates:** edit `cv.yaml` and redeploy. No CMS, no admin UI. Content
  was last synced from the LinkedIn profile on 2026-07-04.
- **Deployment:** single binary on Railway (`PORT` injected; SIGTERM drains
  in-flight requests within 10s for zero-downtime deploys). `/healthz` is the
  platform liveness probe.
- **Analytics:** consent-gated. Consent Mode v2 is set to `denied` by an inline
  `<head>` script before `gtag/js` loads; the choice is stored in `localStorage`
  for ~6 months. **`GA_ID` unset means no analytics, no consent banner, and no
  JavaScript at all.**
- **Contact:** `mailto:` only — no form, no submission backend.

## Capabilities and Constraints

**Shipped surfaces**

| Route | Purpose |
| --- | --- |
| `GET /` | The page: hero → experience → contact |
| `GET /privacy` | Privacy & cookies (`noindex,follow`) |
| `GET /cv` | `301` → `/` (one indexable URL for the content) |
| `GET /cv/download` · `/cv/download.md` | Generated PDF and Markdown CV |
| `GET /robots.txt` · `/sitemap.xml` · `/llms.txt` · `/site.webmanifest` | Ops files |
| `GET /healthz` | Liveness probe |
| `POST /webhooks/resend/inbound` | Inbound mail forwarding (registered only when its env vars are set) |
| any other path | Real `404` with the designed 404 page |

**Content model** (`cv.yaml` → `internal/cv`): name, tagline, intro, summary,
location, availability, languages, email, photo, links (Email/LinkedIn/GitHub),
skills (category → items), work experience (company, role, from/to, skills,
achievements). There is no `projects:` block, by decision.

**Durable engineering constraints** — the standing conventions in `checklist.md`
are binding for this project, not suggestions. In short: Go + `net/http` +
`html/template` + `embed.FS`, server-rendered from typed content; no SPA and no JS
framework; client JS is vanilla, progressive-enhancement, and kept to a budget
(none is best); CSS custom-property tokens; self-hosted subset WOFF2 fonts with
zero third-party font hosts; all served CSS/JS minified and content-hashed;
non-blocking asset loading; `internal/*` packages by responsibility with `main.go`
as thin wiring; single self-contained binary on Railway.

**Deliberate absences** — future work must not quietly add these:
- No contact form or submission backend.
- No blog, CMS, or publishing system.
- No i18n. English only (EN/RU fluency is *stated content*, not a language toggle).
- No auth, no accounts, no database.
- No project showcase or case-study pages in this repo (confirmed 2026-08-04 —
  that work lives on the Soarline Studio site). `.design/project-pages/` is a
  stale brief for this codebase.
- The current employer is intentionally listed as "Payment platform" rather than
  by name; do not resolve or infer it.

**Settled decisions** (earlier briefs list these as open — they are not):
- A portrait photo ships (`static/img/photo.jpg` + `.webp` via `<picture>`), with
  an initials-mark fallback path retained in the template.
- The downloadable CV is generated from `cv.yaml` at boot, not maintained
  separately.
- Dark mode ships via `prefers-color-scheme` plus a JS theme toggle that stays
  hidden until JS reveals it.

## Brand Commitments

- **Name and identity:** Petr Travkin. Domain `konorlevich.tech`; handle
  `konorlevich` (GitHub, LinkedIn). Based in Tbilisi, Georgia.
- **Voice:** warm, confident, plainspoken, first person. Credible to a
  non-technical reader without jargon; never a wall of infrastructure nouns.
- **Binding credit — do not remove or alter:** the footer signature "Website
  created by **Soarline Studio**" (`https://soarline.studio/`) and its matching
  JSON-LD `creator` / `creditText` block on `/`. Soarline Studio is a separate
  studio; the visible credit and the structured-data credit must always agree.
- **Structured data is visible truth:** every value in the JSON-LD `@graph` must
  also be rendered somewhere on the page. This constraint is already documented in
  `templates/pages/home.html` and must survive future edits.

## Evidence on Hand

**Real, verifiable, safe to use** (sourced from the LinkedIn profile, synced
2026-07-04, and stored in `cv.yaml`):
- Kubernetes-centric internal PaaS deploying 100+ services across Go, Kotlin and
  PHP; presented to 150+ colleagues globally (Bumble).
- CI/CD code-quality annotation service saving $100k+ a year (Bumble).
- Pipeline-configuration library adopted across 130+ projects (Bumble).
- Monolith → microservices split cutting API response time 10× (2s → 0.2s), while
  leading a team of three (Linkprofit).
- 20+ partner platforms integrated and maintained (Linkprofit).

**Assets:** portrait (`static/img/photo.jpg` / `.webp`), OG image
(`static/img/og.png`), icon set + favicon, three self-hosted variable fonts.
Review screenshots and prior design reviews live under `.design/*/`.

**Absences future work must not fabricate:** no testimonials, no client
references, no named customers, no traffic or analytics figures, no case studies,
no certifications, no named current employer, no salary or rate information.

## Product Principles

1. **Trust before tech.** Every line must read as credible and human to a
   non-technical hiring reader first; technical depth is available on a second
   look (CV, GitHub), never the price of entry.
2. **One action.** Emailing Petr is the single primary CTA, reachable at every
   natural decision point. Competing actions — CV download, outbound links — stay
   visually quieter.
3. **The artifact is the argument.** Speed, correctness, accessibility and clean
   markup are product features here, not polish. Nothing ships that the CV's own
   claims would be embarrassed by.
4. **Facts live in `cv.yaml`.** Content is data rendered through templates;
   resist hardcoding claims into markup, and never state something the data does
   not support.
5. **Nothing loads that the visitor didn't allow.** Zero third-party requests by
   default, no JavaScript at all when analytics is off, and consent asked plainly
   before anything changes.

## Accessibility & Inclusion

**WCAG 2.1 AA is the floor** (`checklist.md` §8), not an aspiration — the audience
includes recruiters using assistive tech and reading on phones in poor conditions.

- Text ≥ 4.5:1 contrast, large text and UI ≥ 3:1, in **both** light and dark
  themes. Warmth must not cost legibility.
- Full keyboard operability with visible focus rings and logical tab order;
  skip link to main content.
- Semantic landmarks (`header` / `main` / `footer`), headings in order,
  descriptive link text, meaningful or intentionally empty `alt`.
- Touch targets ≥ 44×44px.
- `prefers-reduced-motion` respected; motion stays subtle and optional.
- **The page must be fully usable with JavaScript disabled.** The theme toggle
  stays hidden until JS reveals it; `prefers-color-scheme` covers the no-JS case.
- English only; no RTL requirement.
