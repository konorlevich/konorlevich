# Project Checklist — Standing Conventions

These are the decisions we keep making the same way. 
Use it as the default; deviate only on purpose and write down *why*.

---

## 0. The one-line summary of "our stack"

> **Go + `net/http` + `html/template` + `embed.FS`, server-rendering static
> HTML from typed content (Go structs / YAML / JSON), zero JS framework,
> vanilla progressive-enhancement JS, CSS custom-property tokens, self-hosted
> fonts, locale-as-URL-prefix i18n, deployed as a single self-contained binary
> on Railway.**

Everything below is the detail behind that sentence.

---

## 1. Tech stack & framework

- [ ] **Backend: Go**, module path `github.com/konorlevich/<project>`. Pin a Go version (recent projects: 1.26).
- [ ] **Layout: code properly spread across `./internal/*` packages** by responsibility (e.g. `internal/site`,
  `internal/content`, `internal/render`, `internal/handler`) — not one flat `main` package. `main.go` (or `cmd/*`) stays
  a thin wiring layer: parse config, construct dependencies, start the server.
- [ ] **Routing: stdlib `net/http.ServeMux`** with Go 1.22+ method+pattern routes (`GET /{$}`,
  `GET /{lang}/services/{slug}`). No router library unless justified.
- [ ] **Rendering: `html/template`**, parsed **once at startup** (fail-fast on boot).
- [ ] **No SPA / no JS framework** for content. If a prior project used Vite/Tailwind-SPA, the standing decision was to
  *rip it out* and pre-render.
- [ ] **Client JS: as little as possible — none is best.** Vanilla, IIFE-wrapped, progressive-enhancement only. Prefer
  no JS at all; when unavoidable, prefer **inlining it** over a separate file. Keep a JS budget.
- [ ] **CSS: inline where possible** (critical CSS in `<head>`); how CSS/JS code is *structured* doesn't matter — what
  matters is what ships. If Tailwind is used, treat it explicitly as *a CSS build tool, not a JS framework* (adds a Node
  build stage), wired onto the same `tokens.css` via v4 `@theme`.
- [ ] **Served JS and CSS are always minified.** Run the builder/minifier on a **pre-commit hook** so minified output is
  never stale.
- [ ] **Non-blocking asset loading:** external CSS/JS `<link>`/`<script>` must never block render — `defer`/`async`,
  `rel="preload"` + swap, or `media`-trick for non-critical CSS; inline the critical path instead.
- [ ] **Self-host all fonts** (WOFF2, `font-display: swap`, subset, preload the 2 above-the-fold faces). Zero
  third-party font hosts / no Google Fonts `<link>` (CWV + privacy).
- [ ] **Databases only when needed:** Postgres in prod, and if SQLite is also a target, write **dual-dialect
  migrations** (goose, `ON DELETE CASCADE`). GORM where an ORM helps.
- [ ] **Third-party Go SDKs only where required, and always the canonical ones:** email → official
  `github.com/resend/resend-go/v3`; Telegram Bot API → `github.com/go-telegram/bot`; plus `fpdf`, `bluemonday` (HTML
  sanitize), `markdown` renderer as needed.

## 2. Static building / rendering approach

- [ ] **Bundle everything into the binary** via `//go:embed` (templates, static, content, locale JSON). Startup fails
  fast if any asset/template/locale is missing.
- [ ] **Server-render primary content — never JS-inject SEO-critical content.** "Correct on first byte over correct
  after hydration."
- [ ] **Pre-render static pages to bytes at boot** (landing, project/service pages, 404, sitemap, robots, llms.txt) and
  serve from an in-memory map keyed by URL path. Only genuinely dynamic responses (form POST results) execute a template
  per request.
- [ ] **Precompress once at boot** (Brotli + gzip), compute `ETag`/`Last-Modified`; hot path = buffer write +
  conditional-request (304) check. Negotiate via `Accept-Encoding` + `Vary: Accept-Encoding`.
- [ ] **Composition over copy-paste:** shared `head` / `header` / `footer` / `langswitch` / `cookiebanner` partials;
  reusable `{{define}}`/`{{template}}` blocks. One template driven by a content flag can serve multiple similar
  surfaces.
- [ ] **Template-per-page map** (`map[string]*template.Template`, one entry = layout + partials + that page's content)
  to avoid `{{define "content"}}` collisions.
- [ ] **Content is data, not code, and there is no CMS.** Content lives as typed Go structs / YAML front-matter /
  per-slug JSON, loaded into in-memory maps at boot. **Adding a page = drop a data file + regenerate sitemap, no routing
  change.**
- [ ] **Cache headers:** content-hashed assets `Cache-Control: public, max-age=31536000, immutable` (+ precompressed
  `.br`/`.gz`); HTML `no-cache`. Register `font/woff2` MIME.

## 3. Webhooks / integrations / APIs

- [ ] **Default to zero integrations.** Prefer client-side-only (invoice/PDF/XML generated in-browser, no upload),
  `mailto:`, or deep links (`wa.me/…?text=`, `t.me/…`, Instagram) over building backends. "No network spinner" is a
  feature.
- [ ] **Lead delivery: Resend** (`resend-go`). Send in a **background goroutine** so the user gets instant success; log
  as fallback; show a `mailto:` fallback if send fails so a lead is never lost. Lead email stays in English even on
  localized sites.
- [ ] **Resend webhooks only if we serve inbound emails — and ask first.** Not part of the default lead-send setup.
  When they are added (`POST /webhooks/resend`):
    - [ ] **Verify the signature before processing** (`client.Webhooks.Verify`, HMAC-SHA256 over Svix headers) — reject
      unverified/replayed.
    - [ ] **Respond 2xx fast.**
    - [ ] **Process idempotently** (dedupe by event id).
- [ ] **Admin-API pattern (content sites like ai-news):** bilingual content created/edited via an admin form →
  `POST /admin/...`, parsed from `r.PostForm` with indexed fields (`faq[0][ka_question]`). Automated content skills post
  through this same API.
- [ ] **Spam protection:** Cloudflare Turnstile (`data-appearance="interaction-only"`, verified server-side) and/or a
  honeypot field.
- [ ] **Ops routes on every service:** `/healthz`, `/sitemap.xml`, `/robots.txt`, `/site.webmanifest`, `/llms.txt`.
- [ ] **Config from env vars, never hardcoded:** GA id, `GTM_ID`, `GA4_MEASUREMENT_ID` + `GA4_API_SECRET` (server-side
  events), `RESEND_API_KEY`, `RESEND_WEBHOOK_SECRET`, `LEAD_TO`/`LEAD_FROM`, `BASE_URL`, `LOCALES`, business data
  (phone/hours/address). Central `site.Config` struct with placeholder
  defaults, injected into every page's template data.

## 4. Deployment & hosting

- [ ] **Target: Railway.** Deploy unit = **single self-contained Go binary** with embedded assets. No Dockerfile or railway.toml needed.
- [ ] **Multi-stage build:** (build CSS via Node/Tailwind if used) → build Go binary → serve. For DB-backed services:
  migrator one-shot (`cmd/migrator`, goose) runs before the app (`depends_on`).
- [ ] **`/healthz`** liveness endpoint wired to the platform health check.
- [ ] **Graceful shutdown:** trap `SIGTERM`/`SIGINT`, call `server.Shutdown(ctx)` with a bounded timeout so in-flight
  requests drain before exit; close DB pools / flush the logger / stop background goroutines (e.g. pending Resend sends)
  on the way out. Root the whole app on a cancellable `context.Context`. Clean exit (0) — Railway sends `SIGTERM` on
  redeploy, so this must be correct for zero-downtime deploys.
- [ ] **The server produces valid HTTP access logs to stdout** — strictly one parseable line per request in a standard
  format (valid Combined Log Format, or one well-formed JSON object per line), so any log aggregator can consume them
  unmodified: timestamp, method, path, status, bytes, duration, client IP, referer, UA, request id. **Never log PII**
  from form bodies/cookies. Errors logged separately with request id.
- [ ] **Logging library: `logrus`** (`github.com/sirupsen/logrus`) — configure `JSONFormatter` for structured
  one-object-per-line output, log via fields (`log.WithFields`) not string interpolation, to stdout. One shared
  configured logger, injected — no ad-hoc `fmt.Println`/stdlib `log` in handlers.
- [ ] **Correct status codes throughout:** 200 / 301 / 302 / 400 / custom-404 / 405 / 5xx.
- [ ] **Infra checklist:** HTTPS enforced; www↔apex 301 to one canonical host; custom 404; SSL.
- [ ] **Runnable locally with `go run .`** — dev target decoupled from deploy.

## 5. i18n / localization

- [ ] **Locale = first URL path segment** (`/en/…`, `/ka/…`), ISO-639-1, from a whitelist. Never a query param. Every
  page exists at its own indexable URL per language.
- [ ] **Slugs are locale-invariant** (only the prefix differs); proper nouns / standard names stay un-translated
  (`/xrechnung`, `/facturae`, IVA, IRPF, Leitweg-ID).
- [ ] **Root `/` and legacy unprefixed paths → 302 redirect** to a detected locale. **Detection order: cookie →
  `Accept-Language` → default.** Redirect never indexed.
- [ ] **`Vary: Cookie, Accept-Language`** on redirects and prefixed pages so caches never serve the wrong locale.
- [ ] **Locale cookie** (first-party functional, `Path=/`, 1y, `SameSite=Lax`, `Secure` in prod), (re)written on every
  locale page → language switcher can be a plain `<a>` link, no JS.
- [ ] **Translations are content, not code:** move all copy into `content/i18n/<locale>.json` / per-locale Go bundles /
  YAML, decoded into a typed `SiteCopy` struct. Templates read `.Copy.Hero.Headline` — never branch on locale in the
  template. Organize as **per-page namespaces + a small `common.*`** (watch for key collisions across pages).
- [ ] **Boot-time completeness gate:** reflect-walk every enabled locale; any missing/empty string is a **fatal startup
  error**. A half-translated page can never ship. English is the reference locale.
- [ ] **Every localization set is reviewed by a native-language agent** before its locale ships: spawn a dedicated
  review agent per locale, prompted as a native speaker, to sweep the full `content/i18n/<locale>` set against the
  English reference for grammar, naturalness, register/tone, terminology consistency, and locale traps (plural forms,
  formality level, untranslatable proper nouns). Findings fixed before the locale is enabled via `LOCALES`.
- [ ] **Partial-rollout flag** (`LOCALES=en` vs `LOCALES=en,ka`) ships English now, turns on a locale page-by-page as
  translation clears native review. Where a per-feature content-presence gate exists (`HasGuide(lang)`), an untranslated
  route serves a localized 404 — never thin/duplicate content.
- [ ] **Language switcher UX:** real `<a href>` to the exact counterpart URL (reuse the hreflang `Alternates` list —
  never dump to home); no flags, no dropdown for 2 langs; each language named **in its own script** ("English" /
  "ქართული"); active locale marked with `aria-current` **plus a non-color cue** (weight/underline).
- [ ] **Script-aware typography:**
    - [ ] Emit Georgian/Cyrillic `@font-face` **only on the pages that need it** (scope tokens under
      `:root[lang="ka"]`).
    - [ ] **Never `text-transform: uppercase`** on Georgian (Mkhedruli has no caps) or Cyrillic — substitute
      letter-spacing + weight + color.
    - [ ] Tune `overflow-wrap`/`hyphens` for long German compounds & Georgian words at 375px.
    - [ ] Mind grammar: Russian plurals ("1 отзыв" vs "N отзывов").
- [ ] **Locale-aware formatting** (currency, numbers, dates) per locale — but pricing/currency conversion & RTL are
  explicitly **out of scope** by default.

## 6. SEO

- [ ] **Per-page, server-rendered head suite:** unique `<title>` (≤62 chars), meta description (≤155 chars, keyword
  front-loaded), self-referential absolute `<link rel=canonical>` (UTM stripped), OpenGraph + Twitter card
  (`summary_large_image`), `og:locale`, favicon.
- [ ] **Favicon in the most broadly supported format:** ship a classic **`/favicon.ico`** (multi-size ICO: 16/32/48) at
  the site root as the universal fallback every browser and crawler finds, plus a modern `icon` PNG/SVG and an
  `apple-touch-icon` (180×180 PNG) referenced from `<head>`, and the manifest icons. ICO-at-root is the non-negotiable
  baseline.
- [ ] **Correct `<html lang="{lang}">`** per page.
- [ ] **Full reciprocal hreflang cluster** for all locales **+ `x-default` → the `en` URL** (Google needs bidirectional
  confirmation). `og:locale:alternate`, JSON-LD `inLanguage` per page.
- [ ] **Single `<h1>` per page**, logical heading order (the heading map doubles as SEO structure).
- [ ] **`sitemap.xml`** (all localized pages with reciprocal `<xhtml:link hreflang>`, generated from `content.Order()` +
  a suffix loop) and **`robots.txt`** (references sitemap + llms.txt).
- [ ] **Real `<lastmod>` per entry — the date that page's *content* actually changed,** not build/deploy time and not
  `time.Now()`. Every entry sharing one timestamp that moves on each deploy is a lie crawlers learn to ignore; Google
  drops `lastmod` from scheduling once it stops correlating with real changes. Source it from a per-page `updated`
  field in the content data file (fall back to `published`), carry it through `content.Order()`, and take a section
  hub's `lastmod` as the max over its children. Format as W3C datetime (`2026-08-07` or full RFC 3339 with offset).
  Touching CSS, a footer link, or a translation string is **not** a content change — don't bump it. Verify: two
  consecutive deploys with no content edits produce a byte-identical sitemap.
- [ ] **`/llms.txt`** (Markdown at root, per-locale) mirroring on-page facts for AI discoverability.
- [ ] **JSON-LD structured data as an `@graph`,** matched to page type: `Organization`, `WebSite`, `Service`/`Offer`,
  `ProfessionalService`/`AutoRepair`, `NewsArticle`/`Article`/`TechArticle`, `BreadcrumbList`, `FAQPage`, `HowTo`. **No
  fake `AggregateRating`.**
- [ ] **Creator credit in the `WebSite` JSON-LD on every site:** `creator` → Soarline Studio
  (`https://soarline.studio/`) + `creditText`. Substitute the site's own `url`/`name`; the `creator` block stays
  verbatim:

  ```html
  <script type="application/ld+json">
  {
    "@context": "https://schema.org",
    "@type": "WebSite",
    "@id": "https://example.com/#website",
    "url": "https://example.com/",
    "name": "Example Website",
    "creator": {
      "@type": "Organization",
      "@id": "https://soarline.studio/#organization",
      "name": "Soarline Studio",
      "url": "https://soarline.studio/"
    },
    "creditText": "Website created by Soarline Studio"
  }
  </script>
  ```
- [ ] **Matching visible credit link in the footer on every site** (keeps the schema honest per the "visible truth"
  rule — the credit exists in the DOM, not just in markup):

  ```html
  <footer>
    Website created by
    <a href="https://soarline.studio/">Soarline Studio</a>
  </footer>
  ```
- [ ] **"Visible truth" rule:** never emit schema for content not rendered on the page; FAQ schema text must equal
  visible answer HTML (avoids content/markup-mismatch penalties). Content stays in the DOM regardless of `<details>`
  open/closed. Validate with Rich Results Test.
- [ ] **1200×630 OG image** in the site's own brand, absolute URL, with `og:image:alt`. (Watch for OG images pointing at
  a dead domain — a real past bug.)
- [ ] **Robots granularity:** distinguish `noindex,nofollow` from `noindex,follow`; thin/workbench/search pages use
  `noindex,follow` so crawlers still reach real content.
- [ ] **Index gating by content threshold** for combinatorial pages (e.g. a tag-intersection hub is `index` only with ≥3
  shared items; below → `noindex,follow`, excluded from sitemap). "Crawl budget tracks real content, not URL count."
- [ ] **Canonicalization rules for generated URLs:** alphabetize combo slugs (non-canonical order → 301), identical
  slugs → 301 to single hub, unknown/too-deep → 404.
- [ ] **Custom 404 returns real HTTP 404 + `noindex`, excluded from sitemap.**
- [ ] **Hub-and-spoke internal linking** + informational-vs-transactional split (guide pages own question keywords;
  generator/tool pages stay transactional; cross-link them).
- [ ] **First-touch UTM attribution:** read `utm_*` + referrer on landing, store first-touch in a first-party cookie,
  attach to the lead email (the inbox is the register). Keep a consistent outbound UTM tagging convention; strip UTM
  from canonical; no UTM on internal links.
- [ ] Run the **`seo-audit` skill** before launch.

## 7. Design system / tokens / theming

- [ ] **One `tokens.css` = single source of truth,** CSS custom properties: `--color-*`, `--space-*` (8px base scale),
  type ramp (`--font-size-*` / fluid `clamp()`), `--border-radius-*`, `--shadow-*`, `--shadow-focus`, `--easing-*`
  /motion.
- [ ] **"Extend, don't replace / no new tokens"** is a hard rule. New features reuse the existing contract; add only
  minimal *semantic* tokens (e.g. `--on-accent`, `--inv-deduct`) — never a new raw hue.
- [ ] **Name the aesthetic philosophy** per project and hold it consistently (past examples: "Quiet Editorial," "Warm
  Minimalism," "Editorial minimalism + brand accent," "Functional-clean SaaS," "child-friendly paper-cut"). Record
  reference points **and anti-references**.
- [ ] **One confident accent used as information, not decoration.** Never dilute with a second accent. Verify the accent
  passes contrast as text (gold/warm accents often fail → use a darker link variant).
- [ ] **Typography = display face + body face (+ optional mono),** self-hosted, fluid ramp. Derive palette from the
  brand asset/logo.
- [ ] **Semantic tokens over raw values everywhere,** so both light and dark themes are correct by construction.
- [ ] **Dark theme ships by default, unless said otherwise — and it's designed, never an automatic inversion:**
    - [ ] Default: **ship dark** (warm/charcoal designed palette via `prefers-color-scheme: dark`, accent lightened,
      contrast re-verified). Light-only is the explicit opt-out; then dark token *values* may exist but must be
      **gated/inert** — auto-activating dark has been a real regression bug.
    - [ ] **When both dark and light themes exist, ship a visible theme toggle** (in the header/nav — discoverable, not
      buried), an accessible real control with a non-color state cue. The toggle cycles **three modes — `dark` /
      `light` / `auto`** — where `auto` follows `prefers-color-scheme` (the initial default before any choice) and the
      two explicit modes override it. Persist the chosen mode in `localStorage` (`try/catch`-wrapped, store the literal
      `dark|light|auto`), apply **pre-paint in `<head>`** to avoid flash; optional deep-link `?theme=dark|light|auto`.
    - [ ] Document/invoice/print previews deliberately stay light "paper."
- [ ] **Reading widths:** prose measure ~45–75ch / ~68ch; page container ~72–80rem.
- [ ] **Focus ring convention reused verbatim:** `outline: 2px solid var(--accent)` / shared `--shadow-focus`.
- [ ] **Motion is minimal & gated:** signature is a subtle fade-up-on-scroll (~12px, IntersectionObserver/CSS); content
  fully present without JS.

## 8. Accessibility (WCAG 2.1 AA is the floor, everywhere)

- [ ] **Contrast:** body/UI ≥ 4.5:1, large text/UI ≥ 3:1, **measured numerically in both light and dark**
  (warm/gold/earthy palettes are contrast traps — re-measure them). Aim ≥7:1 where feasible. Target 0 axe-core
  violations.
- [ ] **Semantic landmarks** (`header`/`nav`/`main`/`footer`), exactly **one `<h1>`**, logical heading order, working
  **skip link**.
- [ ] **Real controls, not div-clicks:** `<button type="button">` with accessible names; native `<details>`/`<summary>`
  for accordions (keyboard + SR + no-JS free).
- [ ] **Forms:** visible `<label>` per field, errors via `aria-describedby`, `aria-live`/`role="status"` for submit
  status, `role="alert"` error banner, top-level error summary + inline messages, **preserve entered values**, focus
  first invalid field.
- [ ] **CTAs are real `<a>` links** (`tel:`, `wa.me`, `t.me`) — work without JS and for AT.
- [ ] **External links open in a new tab, and say so.** Any `<a>` pointing off-domain (social, `wa.me`/`t.me`, the
  Soarline Studio footer credit, cited sources) gets `target="_blank"` **plus `rel="noopener noreferrer"`** — never
  `target="_blank"` alone (`window.opener` hijack + referrer leak). Announce it: a small external-link icon
  (`aria-hidden`) **and** a visually-hidden "(opens in new tab)" in the accessible name, so it isn't a surprise for
  screen-reader and keyboard users. **Internal links stay in the same tab** — never steal navigation from the user;
  `mailto:`/`tel:` get no `target` at all. Enforce it in the shared link partial/helper, not per-template, and check it
  in the pre-launch sweep (every off-domain `href` has both attributes).
- [ ] **Never color alone** — pair status/required/negative with text or icon.
- [ ] **Touch targets ≥ 44×44px**; inputs ≥16px font (avoid iOS zoom); respect `env(safe-area-inset-bottom)` for pinned
  bars.
- [ ] **`prefers-reduced-motion` honored** (zero durations *and* translate distance); never hide content when JS/motion
  is off.
- [ ] **Images:** meaningful localized `alt`; decorative `alt=""`/`aria-hidden`. Localized `alt`/`aria-*` must not leak
  English onto a non-English page.
- [ ] **Live regions** (`aria-live`) for dynamic recalculation/status; disabled controls expose the reason via
  `aria-describedby`.
- [ ] Consent banner is **non-modal** (`<aside>` complementary landmark, not `role="dialog"`) — must not trap or steal
  focus.

## 9. Responsive / breakpoints

- [ ] **Mobile-first, single-column base.**
- [ ] **Standard review viewports: 375 / 768 / (1024) / 1280** — capture at these exact sizes every review.
- [ ] **Reflow, don't shrink:** grids step 4→2→1 (or 3→2→1); chip rows / toolbars `flex-wrap`; previews/actions stack
  below their breakpoint; media reorders above text on mobile.
- [ ] **Two-pane tool layouts engage at ≥1024px** (form left / live preview right); mobile is a *behavior change* (form
  first, preview → toggle/sheet).
- [ ] **Mobile conversion pattern: sticky full-width bottom CTA bar** that appears after the hero scrolls out
  (Call/WhatsApp/"Book"); hidden on desktop.
- [ ] **Zero horizontal overflow at 375 (hard pass criterion),** verified programmatically
  (`scrollWidth === clientWidth === 375`), not eyeballed. Wide content (tables) scrolls inside its own `overflow-x:auto`
  container.
- [ ] **No hamburger where a menu isn't needed** — collapse to logo + language toggle + CTA.
- [ ] **Footer sticks to the bottom of the page.** On short pages (404, thin locale pages, empty states) the footer sits
  at the bottom of the viewport, never floating mid-screen with dead space under it; on long pages it stays in normal
  flow below the content. Pure CSS, no JS: `body { min-height: 100dvh; display: flex; flex-direction: column }` +
  `main { flex: 1 }` (or a grid equivalent). Not `position: fixed` — the footer must not overlay content or fight the
  sticky mobile CTA bar. Verify on the 404 page at 375 and 1280.


## 11. Analytics, events, cookie consent / GDPR

- [ ] **Google tag code preparation — always, in every project:** templates carry the tag snippet, **gated by an env
  var** (`GTM_ID` on `site.Config`). **Render only if the ID is set**; empty var → nothing emitted. Wiring it is never a
  retrofit task.
- [ ] **`GTM_ID` is a *Google tag id*, not necessarily a GTM container — detect the type and emit the matching
  snippet.** Clients hand over whatever ID their Google account showed them, and a GA4 measurement ID pasted into
  `GTM_ID` rendered as a GTM container silently collects **nothing**. Resolve it **once at boot** into a typed value
  (e.g. `site.Tag{Kind, ID}` with `KindGTM` / `KindGA4`), after `strings.TrimSpace` + upper-casing the prefix, and let
  templates branch on `Kind` — never prefix-match a raw string inside the template.
    - [ ] **`GTM-…` → Tag Manager:** container `<script>` in `<head>` **+ the `<noscript>` iframe** as the first thing in
      `<body>`.
    - [ ] **`G-…` → GA4 direct (gtag.js):** `<script async src="https://www.googletagmanager.com/gtag/js?id=G-…">` +
      the inline `dataLayer`/`gtag('config', id)` bootstrap. **No `<noscript>` iframe** — it does nothing for gtag; don't
      copy it over from the GTM branch.
    - [ ] **Exactly one path emits.** If a GTM container is configured, GA4 is configured *inside the container* — never
      also hardcode `gtag('config','G-…')` in the template. Two tags on one page = double-counted pageviews and wrecked
      bounce/engagement metrics.
    - [ ] **`UA-…` → reject.** Universal Analytics stopped processing in 2023; treat it as a stale value to be replaced,
      not something to render.
    - [ ] **Set-but-unrecognized prefix is a fatal boot error,** consistent with the fail-fast rule — a typo'd ID must
      surface on the deploy, not as three months of missing data. Empty/unset stays a clean no-op (local dev, previews).
    - [ ] Accept `GA_ID`/`GA4_ID` as aliases resolving into the same single field if a project already uses them, with a
      documented precedence — but **one resolved tag id reaches the templates**, never two competing vars.
    - [ ] The resolved `Kind` is not only about which snippet renders — it also decides **how every event is sent**. See
      *Event tracking* below; the same typed value drives both, so there is never a second place that guesses the id's
      type.
- [ ] **Google Consent Mode v2:** a **synchronous inline `<head>` script** sets
  `gtag('consent','default', {analytics_storage:'denied', ad_storage:'denied', ad_user_data:'denied', ad_personalization:'denied'})`
  **before whichever tag the resolution above picked** — before the GTM container script, or before `gtag/js` and
  `gtag('config')`. GA runs cookieless until granted. (This head reorder is the risk-first task.) The default-denied
  block is **identical for both branches** — it must not live inside the GTM-only partial, or the GA4 path ships with no
  consent gating at all.
- [ ] **Binary Accept/Decline only** (no granular toggles for a single tracker), **equal weight/ease** (same
  size/position/one tap), no pre-ticked boxes, no cookie wall, no dark pattern. This deliberately overrides the
  "accent = the one primary CTA" habit.
- [ ] **Non-blocking bottom bar,** `<aside>` complementary landmark (not `role="dialog"`), no focus trap/steal; **Esc =
  Decline** (privacy-safe).
- [ ] **Persist choice in `localStorage`** (versioned + timestamped), **~6-month expiry** then re-prompt; **replay in
  `<head>` before config** on return visits; footer **"Cookie settings"** link reopens it (withdrawal as easy as
  granting).
- [ ] **First-party only** — no CMP (Cookiebot/OneTrust rejected), no server-side consent ledger. Centralize consent
  config (key, version, max-age, GA id) in one shared bootstrap file.
- [ ] **Every meaningful action emits a detailed event — analytics is not just pageviews.** A tag that records only
  pageviews cannot answer the one question the client actually asks ("where do the leads come from?"). The event layer
  ships **in the same pass as the tag**, never as a retrofit.
    - [ ] **One `track(name, params)` helper, branching on the resolved tag kind — not on `typeof gtag`.** The server
      hands the client the already-resolved tag (e.g. `window.__tag = {kind:'gtm'|'ga4'}`, rendered from the same
      `site.Tag` — no second source of truth, no prefix-matching in JS):
        - **`KindGTM` → `dataLayer.push({event: name, ...params})`**, and the container maps it to a GA4 event.
        - **`KindGA4` → `gtag('event', name, params)`.** A bare `dataLayer.push` on this branch records **nothing** —
          gtag.js does not forward arbitrary `dataLayer` events — which is the same silent-no-data failure this whole
          section exists to prevent.
        - **No tag configured → `track()` is a silent no-op.** Local dev and previews must never throw.
    - [ ] **Use GA4's recommended event name when one exists, verbatim:** `generate_lead`, `file_download`,
      `select_content`, `search`, `share`, `sign_up`, `view_item`, `login`. Custom names only where none fits;
      `snake_case`, ≤40 chars, ≤25 params. This is the pick-one rule applied to analytics — one canonical name per
      action, defined once in the shared module, never re-worded per template.
    - [ ] **Fire on the outcome, not the intent.** `generate_lead` fires when the server confirms the lead (2xx /
      thank-you render), never on button click — otherwise validation failures and spam bots count as leads. **Forms
      work with JS off, so the event must too:** the server-rendered success page carries the event in its template data
      and emits it at render time, or every JS-off lead is invisible.
    - [ ] **Default event catalog** (adjust per project, write down deviations): `form_start` (first field focus),
      `generate_lead` (confirmed submit — `form_id`, `locale`), `form_error` (`form_id`, `field`/`reason` — never the
      entered value), `file_download` (`file_name`, `file_extension`), `contact_click` (`method`: `tel` | `whatsapp` |
      `telegram` | `mailto`, plus `location`: `header` | `hero` | `sticky_bar` | `footer`), `select_content` for primary
      CTAs (canonical `cta_label` + `location`), `language_switch` (`from`/`to`), `theme_toggle`, `consent_update`
      (`granted`/`denied`), and `page_not_found` on the 404 route (which path missed).
    - [ ] **Don't double-count against GA4 Enhanced Measurement.** It already auto-collects outbound clicks,
      `file_download` (by extension), `scroll`, site search and video. Pick exactly one source per event — turn the
      Enhanced Measurement toggle off for it, *or* don't send ours — never both, or downloads and outbound clicks come
      back doubled.
    - [ ] **Never put PII in an event.** No email/phone/name/message text as params, no PII in `page_location` query
      strings — the same rule as "no PII in logs", and GA4 has no delete-a-value button afterwards. Params stay
      low-cardinality (ids, canonical labels, locale, section), not free text.
    - [ ] **Wiring is declarative and centralized:** one delegated `document`-level listener reads `data-track` /
      `data-track-*` attributes off the element actually used. No per-component listeners, no inline `onclick`, no
      analytics snippets copy-pasted into templates. It lives in the shared bootstrap next to the consent code and stays
      inside the JS budget.
    - [ ] **Events respect consent mode but still fire** — under default-denied GA runs cookieless. Do not build a
      custom queue that stores actions locally and replays them after consent.
    - [ ] **Verify every event once before "done"** — GA4 DebugView (GA4 branch) or Tag Assistant preview (GTM branch):
      each action produces **exactly one** hit with the expected params, JS-off form submit included.
- [ ] **Mirror the conversions server-side — the browser is not a reliable witness.** Ad blockers, tracking
  protection, JS-off and dead tabs swallow a real share of client hits, and the client cannot know whether a lead was
  actually accepted. The server can: it is the only place that knows the honeypot/Turnstile passed, the handler returned
  2xx and Resend took the message. Server-side events are **not** a replacement for the client layer (they have no
  referrer, campaign or device context) — they are the source of truth for **conversions specifically**.
    - [ ] **Transport: the GA4 Measurement Protocol** — `POST` to
      `https://www.google-analytics.com/mp/collect?measurement_id=G-…&api_secret=…`, with `GA4_API_SECRET` from env
      (GA4 Admin → Data Streams → Measurement Protocol API secrets). It is a **secret**: env var only, never embedded
      in a template or shipped to the client.
    - [ ] **It needs a `G-…` id, which the GTM branch does not have.** A container id cannot receive Measurement
      Protocol hits (that is server-side GTM — a separate tagging server, out of scope for a single Railway binary). So:
      `KindGA4` → the resolved id is reused directly; `KindGTM` → require an explicit **`GA4_MEASUREMENT_ID`** alongside
      it, and if it is absent, **the server layer is cleanly disabled**, not silently half-wired.
    - [ ] **Config is fail-fast and all-or-nothing:** secret set without an id (or an id without the secret) is a boot
      error, same rule as a typo'd tag prefix. Neither set → `analytics.Client` is a no-op and nothing is sent (local
      dev, previews, tests).
    - [ ] **Stitch to the same user via `client_id`, taken from the `_ga` cookie on the request** (`GA1.1.<id>.<ts>` →
      strip the two leading parts) — plus `session_id` from the `_ga_<STREAM>` cookie where session stitching matters.
      Skip this and every server event lands as a brand-new direct-traffic user, which inflates users and destroys
      campaign attribution for exactly the conversions you care most about. **No cookie → no identifier:** send a
      one-off random `client_id` and flag the event (e.g. `stitched: false`) so it can be filtered — **never**
      reconstruct an id from IP + User-Agent, which is fingerprinting and is precisely what the consent banner exists to
      prevent.
    - [ ] **Same consent rules as the client.** Server-side is not a consent loophole: no `_ga` cookie means consent was
      never granted (or the user is cookieless), so send the unstitched, identifier-free form or don't send at all —
      document which, per project.
    - [ ] **Send from the background goroutine that already handles the lead**, with a bounded `context` timeout; the
      user's response never waits on Google. Failures are logged and dropped — an analytics outage must not fail a lead.
    - [ ] **Exactly one conversion per action across both layers.** If the server sends `generate_lead`, the client must
      not send it too — pick the server as the source of truth for conversions and drop or rename the client twin.
      Same trap as two tags on one page, one level up.
    - [ ] **What is worth sending server-side:** `generate_lead` (after spam checks pass and delivery is accepted),
      `file_download` from the handler that actually serves the file (the only truthful signal — count the initial
      request, not `HEAD`/`Range` re-requests), server-side `form_error` (validation rejects the client never sees),
      and, where inbound webhooks are wired, delivery outcomes from `POST /webhooks/resend`
      (`email_delivered` / `email_bounced`) so a lead that never arrived is visible.
    - [ ] **Share one vocabulary with the client layer:** event names and param keys as Go constants in
      `internal/analytics`, matching the JS module's names exactly and documented in the same table. Two spellings of
      `generate_lead` is two half-populated reports.
    - [ ] **Know the protocol's teeth:** `/mp/collect` answers **204 for valid *and* invalid payloads** — a malformed
      event is silent, so validate against **`/debug/mp/collect`** while building and keep a `go test` asserting the
      payload shape against an `httptest` server. `timestamp_micros` must be within **72 hours**, so retries are
      short-lived, not queued for a day. Server hits also carry no reliable geo/device data — never build a report that
      depends on it.
    - [ ] **No PII crosses the wire**, same as the client layer: no email, phone, name or message body as params — not
      even hashed "just for matching". Ids, canonical labels, locale, form id.
- [ ] Prefer avoiding a banner entirely (functional-only cookie / privacy-respecting analytics) where the law allows.
- [ ] Lightweight, plain-language **privacy page** (legal basis Art. 6 (1)(a), cookie lifetime, what's stored locally);
  may be `noindex`.

## 12. Cross-cutting engineering conventions

- [ ] **Extend, do not replace** — the single most-repeated rule. New features compose existing
  components/tokens/partials; aim for zero new colors/components.
- [ ] **Graceful degradation is mandatory:** every `localStorage` access `try/catch`-wrapped; forms work with JS off;
  empty/thin/error states omit sections rather than render broken/empty headings.
- [ ] **Consent-first / no surprises:** never repopulate a form or swap content without explicit user action (restore
  banners, format-switch warnings).
- [ ] **Backward-compatible migrations:** dual-write old + new fields through a deprecation window; new field
  authoritative when present; never auto-upgrade legacy rows.
- [ ] **Replace-all-in-transaction** for structured child rows (delete removed / upsert kept / set positions in one tx);
  batched loaders to avoid N+1 on read.
- [ ] **Privacy-as-architecture where the data is sensitive:** 100% client-side, no accounts/upload; state moves between
  tools via a shared `localStorage` draft (never query params / URL / referrer / server log). "No network spinner" = the
  proof.
- [ ] **Feature-gated JS:** heavy per-feature modules load only on their page.
- [ ] **Placeholders are a first-class, clearly-marked stage:** business data, photos, testimonials, metrics ship as
  labeled placeholders (`[HOURS TBD]`, `+971 5X XXX XXXX`); filling them is a pre-launch content task, not a design
  defect. Image slots are labeled `role="img"` boxes now, `<img loading="lazy">` (WebP) later.
- [ ] **Images are undraggable / not casually saveable.** Content images (photos, artwork, portfolio shots) ship with
  `draggable="false"` **plus** CSS `-webkit-user-drag: none; user-drag: none; user-select: none;` so they can't be
  dragged onto the desktop or into another tab. Where the client explicitly asks for it, also suppress the context menu
  on the image only (`contextmenu` listener scoped to `img`, never `document`) and lay a transparent overlay over the
  image so long-press/"Save image as" grabs nothing useful.
    - [ ] **Never `pointer-events: none` on an `<img>` inside a link/button** — it kills the click target.
    - [ ] **Never disable right-click, text selection, or keyboard shortcuts document-wide.** Blocking the whole page is
      a dark pattern, breaks copy/paste and AT, and is the fastest way to make a site feel broken.
    - [ ] Keep `alt` text and semantics intact — the overlay/`user-select` trick must not hide the image from screen
      readers or from `og:image`/crawlers.
    - [ ] **Be honest about what this is: friction, not protection.** The file is still one devtools/Network-tab click
      away. When the asset genuinely must be protected, the answer is *what you serve*, not what you disable — ship a
      downscaled/watermarked derivative and keep the full-res original off the public site.
- [ ] **Integrity / no fabrication:** never invent metrics presented as real; publishing client metrics needs explicit
  permission; unknown stack/timeline/URL → hidden section, never faked; ratings-only until real reviews exist.
- [ ] **Naming discipline ("pick-one rule"):** one canonical label per action, verbatim sitewide (never reworded per
  surface).
- [ ] **Single source of truth** enforced by refactor — no duplicated content across files.
- [ ] **Data-driven repeatable content:** reviews/projects/services/testimonials as struct slices / JSON; no
  pagination/search/archive at small scale (document the threshold to revisit, e.g. ~15–20 items).
- [ ] **Performance budget as a product feature:** 95+ (ideally 100) PageSpeed mobile, green Core Web Vitals, sub-1s
  load — backed by precompile + precompress + immutable-cache + subset-font, not hoped for.
- [ ] **Green-build gate before "done":** `go build`, `go vet`, `go fix ./...`, `gofmt` clean (+ `go test ./...` where
  tests exist) **and** a running-server smoke test of every new route + a 404 check for unknown segments/slugs. `go fix`
  runs as part of build sanitation.

---

## Quick pre-launch checklist (the short version)

- [ ] `/healthz`, `/robots.txt`, `/sitemap.xml`, `/llms.txt`, custom 404, `/site.webmanifest` all present & correct
- [ ] sitemap `<lastmod>` reflects real per-page content dates, not build time (redeploy without content edits →
  identical sitemap)
- [ ] canonical + full hreflang cluster (+ x-default) + JSON-LD validated (Rich Results Test); Soarline Studio creator
  credit present in both `WebSite` schema and the visible footer link
- [ ] all locales pass the boot-time completeness gate; each locale reviewed by a native-language agent; no
  untranslated page ships
- [ ] WCAG AA contrast measured in light **and** dark; visible theme toggle present when both themes ship; 0 axe
  violations
- [ ] zero horizontal overflow at 375 (measured); screenshots captured at 375/768/1280 × light/dark × each locale
- [ ] footer sits at the bottom of the viewport on short pages (check the 404 at 375 and 1280)
- [ ] every off-domain link has `target="_blank"` + `rel="noopener noreferrer"` + a "(opens in new tab)" cue; internal
  links stay in the same tab; content images carry `draggable="false"` + the no-drag CSS
- [ ] consent mode v2 default-denied before `gtag/js`; Accept/Decline equal-weight
- [ ] env config set (Resend keys, webhook secret, GA id, GTM_ID, GA4_MEASUREMENT_ID/GA4_API_SECRET, LEAD_TO/FROM,
  BASE_URL, LOCALES); no secrets hardcoded
- [ ] tag snippet present in templates, renders only when the tag id is set, and **matches the id's type** — `GTM-…`
  container (+ `<noscript>` iframe) vs `G-…` gtag.js (no iframe); exactly one tag fires per page (verify in Tag
  Assistant / the Network tab: one `collect` hit, not two)
- [ ] every meaningful action emits its event (lead, download, contact click, CTA, language switch, 404) via the single
  `track()` helper — verified in DebugView / Tag Assistant preview: one hit per action, no duplicate with Enhanced
  Measurement, no PII in params, JS-off submit included
- [ ] conversions mirrored server-side via Measurement Protocol — `client_id` read from the `_ga` cookie (not a fresh
  one per hit), validated once on `/debug/mp/collect`, and **not** also sent from the client (one `generate_lead` per
  lead, checked in Realtime)
- [ ] (only if serving inbound email) Resend webhook verifies signature, responds 2xx, idempotent
- [ ] JS/CSS minified (pre-commit builder), critical CSS inlined, no render-blocking asset links
- [ ] HTTPS enforced, www↔apex 301, immutable cache on hashed assets, HTML no-cache
- [ ] `go build` / `go vet` / `go fix` / `gofmt` clean; every route smoke-tested; no PII in logs
- [ ] placeholders replaced with real business data; no fabricated metrics
