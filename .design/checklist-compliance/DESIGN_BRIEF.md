# Design Brief: Checklist Compliance Pass

_Feature slug: `checklist-compliance`_ · Written 2026-08-21 ·
Source of truth: [`checklist.md`](../../checklist.md) (142 standing conventions)

> This is a **compliance and correctness pass over an existing, shipped site**, not a
> new feature. Its success condition is unusual: a visitor who knows the site today
> should notice almost nothing changed. Everything here is either invisible
> (analytics, sitemap, tests) or a small correction of something already slightly wrong
> (footer on short pages, external-link affordance).

---

## Problem

Two different people are underserved by the site as it stands today.

**The visitor** hits small, unglamorous friction. On the 404 page the footer floats in
the middle of the screen with dead space beneath it, which reads as a broken page at
exactly the moment the visitor is already unsure whether something is broken. Links in
the footer silently throw them into a new tab with no warning — fine for a sighted
mouse user who notices, disorienting for a screen-reader or keyboard user who does not.
The portrait can be dragged onto a desktop without a thought.

**The site's owner** is flying blind. The site has a complete, careful consent
apparatus — Consent Mode v2 default-denied, a non-modal bar, a 6-month re-prompt — and
a Google Analytics property (`G-RH8KWHKMPZ`) sitting in the account. But the tag id was
never set in the Railway environment, so `NeedsConsent()` returns false and **the entire
apparatus is omitted from every page in production**. No banner, no tag, no data. And
even once the id is set, the site would only ever report pageviews: there is no event
layer, so the one question actually worth asking — *does anyone download the CV, and
from which link?* — has no answer.

Underneath both, the site tells crawlers a small lie. Every sitemap entry carries a
`lastmod` of the build time, so a redeploy that changes nothing announces fresh content.
Google learns to distrust the signal and drops it from scheduling.

## Solution

Close the gap between what `checklist.md` says this project does and what it actually
does — without touching the design.

The audit found the site **substantially compliant already**: embed-everything,
precompress + ETag, fail-fast boot, graceful shutdown, logrus JSON, `@graph` JSON-LD
with the creator credit, three-mode theme toggle with pre-paint, skip link, `noindex`
404 with a real 404 status, and a clean `gofmt`/`vet`/`build`/`test` gate. That
narrows this pass to eight real violations and three documented deviations.

The centre of gravity is analytics. One typed `site.Tag{Kind, ID}` resolved once at
boot replaces two independent id strings that could both emit; a single `track()`
helper and one delegated `data-track` listener give every meaningful action an event;
the tag id finally gets set in the environment so any of it runs at all.

## Experience Principles

1. **Invisible by default, honest where visible** — The pass must not alter the
   established look. The only changes a visitor can perceive are ones that correct a
   wrong impression: a footer that sits where a footer belongs, and a link that admits
   it is about to leave the page.

2. **Instrument the outcome, not the intent** — Events fire when something actually
   happened (the file was served, the choice was recorded), never on a hopeful click.
   One canonical name per action, defined once. A measurement that flatters is worse
   than no measurement.

3. **Fail on the deploy, never in silence** — A typo'd tag prefix is a fatal boot
   error. A `UA-` id is rejected outright. The failure modes this pass exists to fix
   were all silent ones; none of the replacements may be.

## Aesthetic Direction

- **Philosophy**: **Warm Minimalism** — inherited unchanged from
  [`.design/personal-webpage/DESIGN_BRIEF.md`](../personal-webpage/DESIGN_BRIEF.md).
  No new visual direction is being set here.
- **Tone**: Unchanged. Warm, confident, plainspoken, first person.
- **Reference points**: The existing site. This pass is measured against its own
  current appearance — a screenshot diff should be near-empty apart from the footer.
- **Anti-references**: Any change that "improves" the design while passing through.
  Consent-banner dark patterns, a cookie wall, a granular-toggles CMP. An
  analytics build that collects more than it can honestly use.

## Existing Patterns

Everything below already exists and constrains this work. The checklist's hardest rule
— **"extend, don't replace / no new tokens"** — applies in full: this pass adds **zero**
new design tokens and **zero** new colors.

- **Typography**: Fraunces (display, variable) + Inter (body, variable) + JetBrains Mono,
  all self-hosted WOFF2, the two above-the-fold faces preloaded. Fluid `clamp()` ramp.
- **Colors**: 183 custom properties in `static/css/tokens.css`. Warm-paper
  backgrounds (`--color-bg-primary: #faf7f2`), warm near-black text
  (`--color-text-primary: #2a2620`), one terracotta accent
  (`--color-accent-primary: #a8482a`). Designed warm-charcoal dark palette, never an
  inversion.
- **Spacing**: 8px base scale, `--space-0` … `--space-12`, generous high steps for
  section rhythm.
- **Focus**: `--shadow-focus` applied via `:focus-visible`, reused verbatim everywhere.
- **Components reused**: `.footer` / `.footer-links` / `.footer-linklike` (chrome.html),
  `.cookie-bar` + `consent.js`, `.theme-toggle` + `theme.js`, `.btn` family,
  `.skip-link`. CSS is inlined into `<head>` at boot; JS is two small IIFE files
  loaded `defer`.
- **Rendering**: every page rendered to bytes once at boot into `map[string]*Blob`,
  precompressed Brotli + gzip with an ETag. No request executes a template. Any
  addition must survive that model — see *Key Interactions* on the 404 event.

## Component Inventory

| Component | Status | Notes |
| --- | --- | --- |
| `site.Tag{Kind, ID}` | **New** | Typed tag resolved once at boot. `KindGA4` / `KindGTM`; `UA-` rejected; unrecognized set prefix = fatal. Replaces the `GAID` + `GTMID` string pair. |
| `static/js/track.js` | **New** | The single `track(name, params)` helper + one delegated `document` listener reading `data-track` / `data-track-*`. Branches on `window.__tag.kind`, never on `typeof gtag`. No tag → silent no-op. |
| `{{define "extlink"}}` | **New** | Shared partial for every off-domain link: `target="_blank"`, `rel="noopener noreferrer"`, an `aria-hidden` external icon, and a visually-hidden "(opens in new tab)". Enforced in one place, not per-template. |
| `.visually-hidden` | **New (utility)** | Standard clip-based SR-only utility. Needed by `extlink`. No new tokens. |
| `consent.html` | **Modify** | Emit exactly one tag path from the resolved `Kind`. Pass `window.__tag` to the client. Default-denied block stays shared across both branches. |
| `consent.js` | **Modify** | Esc now records `denied` instead of dismissing without a choice. Emit `consent_update`. |
| `chrome.html` | **Modify** | Footer + social links routed through `extlink`. `data-track` attributes on CV, email and theme controls. |
| `theme.js` | **Modify** | Emit `theme_toggle` on change. |
| `ops.go` | **Modify** | `lastmod` from content `updated:`, not `buildTime`. |
| `cv.yaml` / `internal/cv` | **Modify** | New `updated:` date field, validated at boot. |
| `styles.css` | **Modify** | Footer sticky-to-bottom; portrait no-drag CSS; `.visually-hidden`. |
| `home.html` | **Modify** | `draggable="false"` on the portrait. |
| `.cookie-bar`, `.theme-toggle`, `.btn`, tokens | **Unchanged** | Reused verbatim. |

## Key Interactions

**Tag resolution (boot).** One env read resolves to one typed value. `GTM_ID` is
canonical; `GA_ID` and `GA4_ID` are accepted aliases resolving into the same field with
documented precedence, so the id works regardless of which var holds it. `G-…` →
`KindGA4`. `GTM-…` → `KindGTM`. `UA-…` → rejected. Set-but-unrecognized → the process
does not start. Empty → clean no-op, and local dev stays silent.

**Event dispatch (client).** The server renders the already-resolved tag as
`window.__tag = {kind: 'ga4'}`. `track()` branches on it: `KindGA4` → `gtag('event', …)`,
`KindGTM` → `dataLayer.push({event, …})`, absent → return. Wiring is declarative —
elements carry `data-track="file_download"` plus `data-track-*` params; one listener on
`document` reads them. No inline `onclick`, no per-component listeners.

**The 404 event.** The 404 body is a *single pre-rendered blob* served for every missed
URL, so it cannot carry the missed path in its template data. `page_not_found` therefore
reads `location.pathname` at fire time. This is the one event that is not declarative,
and the reason is worth remembering.

**Consent.** Unchanged in shape: non-modal `<aside>`, no focus trap, Accept and Decline
of equal weight and size. Esc now records a real `denied` choice rather than dismissing
silently — both are privacy-safe, but the current behaviour re-prompts a keyboard user
on every visit forever. Every recorded choice emits `consent_update`.

**Footer on short pages.** `body { min-height: 100dvh; display: flex; flex-direction: column }`
with `main { flex: 1 }`. Pure CSS, no JS, not `position: fixed`. Must be verified
against the existing `background-attachment: fixed` glow and `overflow-x: clip` guard,
which is where this change is most likely to go wrong.

### Event catalog

Derived from what this site actually contains. There is **no form**, so `form_start`,
`generate_lead` and `form_error` have nothing to fire on; i18n is out of scope, so
there is no `language_switch`.

| Event | Fires when | Params |
| --- | --- | --- |
| `file_download` | CV PDF or Markdown is requested | `file_name`, `file_extension`, `location` |
| `contact_click` | a `mailto:` link is used | `method: mailto`, `location: header \| hero \| footer` |
| `select_content` | a primary CTA is used | `cta_label`, `location` |
| `theme_toggle` | theme mode changes | `to: dark \| light \| auto` |
| `consent_update` | a consent choice is recorded | `status: granted \| denied` |
| `page_not_found` | the 404 renders | `path` (from `location.pathname`) |

No PII in any param — no email address, ever. All values are low-cardinality labels.

## Responsive Behavior

No layout changes. The one responsive-relevant fix is the footer, which must be
verified at **375 / 768 / 1280** on the **404 page specifically** (the shortest page),
plus the existing hard pass criterion of **zero horizontal overflow at 375**, measured
programmatically (`scrollWidth === clientWidth === 375`), not eyeballed.

The external-link icon must not cause the footer link row to wrap awkwardly at 375.

## Accessibility Requirements

- The external-link cue must be **both** visual (an `aria-hidden` icon) **and**
  announced (visually-hidden "(opens in new tab)" inside the accessible name).
  Never color alone.
- `.visually-hidden` must clip, not `display: none` — the text has to reach the
  accessibility tree.
- The consent bar stays a non-modal `<aside>` complementary landmark. No focus trap,
  no focus steal on first render. Esc handling must not break that.
- The no-drag CSS must not use `pointer-events: none`, must not disable right-click or
  selection document-wide, and must leave `alt` text and `og:image` intact.
- Contrast unchanged — no new colors are introduced, so the existing AA measurements
  hold. Re-verified in the review phase in both themes anyway.
- Existing skip link, single `<h1>` per page, and `--shadow-focus` ring convention are
  preserved verbatim.

## Out of Scope

- **i18n (`checklist.md` §5) — deliberate deviation.** §5 presupposes a multi-locale
  site. This one has a single language and no translated CV. Adding an `/en/` prefix
  alone buys a redirect, a `Vary: Cookie, Accept-Language` header and a language
  switcher with nothing to switch to — real cost, zero reader benefit. Revisit when a
  second locale is genuinely planned; the hreflang cluster, locale cookie, boot-time
  completeness gate and native-speaker review all come as a set at that point.
- **Server-side Measurement Protocol — deferred.** No lead form exists, so the only
  conversion is the CV download, and the client layer covers it. Skipping this also
  keeps `GA4_API_SECRET` out of the deployment entirely.
- **Pre-commit minifier — satisfied by other means.** Minification runs in-process at
  boot, so served CSS/JS is minified and structurally *cannot* be stale — which is the
  outcome the pre-commit rule exists to guarantee. No code change.
- **Ticking `checklist.md`'s boxes.** It stays a shared conventions doc, not this
  repo's state file.
- Visual redesign of any kind. New pages. Reviving `project-pages`. Content edits to
  `cv.yaml` beyond the new `updated:` field. Touching the mail-forwarding path.

## Deploy steps (cannot be done from the codebase)

1. Set the tag id in the Railway `production` environment — currently unset, which is
   why nothing is collected today.
2. Turn **off** GA4 Enhanced Measurement → "File downloads", since this pass sends its
   own `file_download`. Leaving both on returns every download doubled.
3. Verify each event once in GA4 DebugView: exactly one hit per action, expected
   params, no duplicates.

Steps 1–3 are **blocked pre-launch items**, not claims. Until they are done, the event
layer is built and unit-tested but unverified against a live property.
