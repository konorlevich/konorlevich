# Design Review: Checklist Compliance Pass

Reviewed against: [`DESIGN_BRIEF.md`](DESIGN_BRIEF.md)
Philosophy: **Warm Minimalism** (inherited — this pass sets no new direction)
Date: 2026-08-21
Method: Playwright 1.61.1 driving headless Chromium against the real binary
(`GA_ID=G-RH8KWHKMPZ`, `localhost:8099`). No Playwright MCP was available, so the
browser was driven directly; every number below is measured, not estimated.

> The earlier review of this folder is preserved at
> [`DESIGN_REVIEW-2026-08-03.md`](DESIGN_REVIEW-2026-08-03.md). Its screenshots
> (`review-*.png`) are untouched — this pass writes `compliance-*.png` so both sets
> of evidence survive.

## Screenshots Captured

| Screenshot | Breakpoint | Description |
| --- | --- | --- |
| `screenshots/compliance-home-desktop-1280.png` | Desktop (1280×800) | Home, light, full page |
| `screenshots/compliance-home-tablet-768.png` | Tablet (768×1024) | Home, light, full page |
| `screenshots/compliance-home-mobile-375.png` | Mobile (375×812) | Home, light, full page |
| `screenshots/compliance-home-dark-desktop-1280.png` | Desktop (1280×800) | Home, dark |
| `screenshots/compliance-home-dark-mobile-375.png` | Mobile (375×812) | Home, dark |
| `screenshots/compliance-privacy-desktop-1280.png` | Desktop | Privacy, incl. the new external-link cue |
| `screenshots/compliance-privacy-tablet-768.png` | Tablet | Privacy |
| `screenshots/compliance-privacy-mobile-375.png` | Mobile | Privacy |
| `screenshots/compliance-404-desktop-1280.png` | Desktop | **404 — the footer-anchoring evidence** |
| `screenshots/compliance-404-tablet-768.png` | Tablet | 404 |
| `screenshots/compliance-404-mobile-375.png` | Mobile | 404, footer link row wrap |
| `screenshots/compliance-404-dark-desktop-1280.png` | Desktop | 404, dark |
| `screenshots/compliance-404-dark-mobile-375.png` | Mobile | 404, dark |
| `screenshots/compliance-404-consent-open-desktop-1280.png` | Desktop | **Consent bar open over the anchored footer (post-fix)** |
| `screenshots/compliance-404-consent-open-mobile-375.png` | Mobile | Consent bar open, 187px tall |
| `screenshots/compliance-consent-focus-accept.png` | Desktop | Keyboard focus on Accept |
| `screenshots/compliance-extlink-focus.png` | Desktop | Creator-credit link focused |
| `screenshots/compliance-theme-toggle.png` | Desktop | Theme control after cycling |
| `screenshots/compliance-skiplink-focus.png` | Desktop | Skip link revealed on first Tab |

> All screenshots are in `.design/checklist-compliance/screenshots/`.
>
> **Reading note:** in `fullPage` captures the `position: fixed` consent bar is
> painted at its *viewport* offset, so it appears mid-document rather than pinned
> to the bottom. That is a capture artifact, not a layout bug — the
> `-consent-open-` screenshots show its true position.

## Summary

Nine of the ten build tasks landed clean and the pass achieved what the brief asked
for: the site is measurably more compliant with no visual redesign. **The visual
review earned its place** — it caught one genuine regression that the code review,
the unit tests and the green build all missed: anchoring the footer to the bottom
of the viewport put it directly underneath the fixed consent bar, hiding the
colophon and the Soarline Studio credit. That credit is asserted by the `WebSite`
JSON-LD, so hiding it broke the "visible truth" rule while every automated check
stayed green. It was found in `compliance-404-desktop-1280.png`, fixed during the
review, and re-verified at both breakpoints and both scroll extremes.

The one finding that remains open is **pre-existing and was previously reported as
passing**: touch targets. The 2026-08-03 review states "touch targets are ≥44px";
measurement at 375px contradicts that for eleven controls.

## Must Fix

**None outstanding.** One must-fix was found and closed inside this review:

1. ~~**Consent bar covered the footer credit**~~ — **FIXED.** With the footer
   anchored to the viewport bottom (task 6) the `position: fixed` consent bar
   overlaid it, hiding `.colophon` and the `Website created by Soarline Studio`
   credit. Visible in the first capture of `compliance-404-desktop-1280.png`.
   _Fix applied:_ the consent controller now measures the bar and publishes
   `--consent-bar-height` on `:root` while it is open; `body` carries
   `padding-bottom: var(--consent-bar-height, 0px)`, released the moment a choice
   is recorded. Measured rather than hardcoded because the bar wraps from 106px
   (desktop) to 187px (375px). Re-verified: credit not covered at 1280 or 375, on
   both a short page (404) and a long one (home), at scroll-top and scrolled
   fully to the bottom.

## Should Fix

1. **Touch targets below 44×44px at 375 — pre-existing, and previously
   mis-reported as passing.** Measured on the home page at 375:

   | Control | Size | Shortfall |
   | --- | --- | --- |
   | Footer `Email` | 35×22 | height |
   | Footer `LinkedIn ↗` | 67×22 | height |
   | Footer `GitHub ↗` | 57×22 | height |
   | Footer `Privacy` | 49×22 | height |
   | Footer `Cookie settings` | 103×22 | height |
   | `Website created by Soarline Studio ↗` | 113×17 | height |
   | Theme toggle | 30×38 | both |
   | `Email me` (header) | 95×40 | height |
   | `Download my CV →` | 161×27 | height |
   | `PDF document` / `Markdown` menu items | 210×38 | height |
   | `hello@konorlevich.tech` | 185×18 | height |
   | Privacy `What this means` | 110×17 | height |

   `checklist.md` §8 requires ≥44×44px. **This pass did not introduce it** — the
   CSS diff is 85 insertions and 1 deletion, and none of it changes these
   heights; the `↗` glyph only widened two footer links. But the earlier review's
   claim is wrong and should not be inherited as a pass.
   _Fix: add `min-height: 44px` plus vertical padding to `.footer-links a`,
   `.footer-linklike`, `.theme-toggle`, `.link-arrow`, `.dl-menu-list a` and
   `.contact-email a`. Purely additive, no token changes. Deliberately left for a
   separate pass rather than smuggled into this one — it is a real visual change
   to spacing that deserves its own review._
   (The `.skip-link` also measures 1×1, which is correct: it is clipped until
   focused, and expands to a full control on `:focus`. Not a finding.)

2. **The consent bar takes 187px — 23% of the viewport — at 375.** See
   `compliance-404-consent-open-mobile-375.png`. Its body copy wraps to three
   lines. This mattered less when the bar simply overlaid content; now that the
   page correctly reserves space for it, that height is 187px of real layout cost
   on every first visit. _Fix: tighten the copy at narrow widths or reduce
   `.cookie-bar` vertical padding below 480px. Copy is content, so it is a
   deliberate decision rather than a mechanical fix._

## Could Improve

1. **The portrait's cyan background fights the warm palette.** Clearest in
   `compliance-home-dark-desktop-1280.png`, where a saturated cyan disc sits in an
   otherwise warm-charcoal page. Pre-existing content, explicitly out of scope
   here. _Suggestion: a warm-neutral or transparent backdrop would sit inside
   Warm Minimalism instead of against it._
2. **`page_not_found` will count bot traffic.** Every 404 fires it, including
   scanners probing for `/wp-login.php`. That is arguably the useful signal, but
   worth knowing before reading the report. _Suggestion: leave as is; filter in
   GA4 if it becomes noisy._
3. **`internal/web` has tests now, but only for the sitemap path.** The tag
   rendering branches are verified by hand (three boots, documented in the task
   list) rather than by a test. _Suggestion: a golden-file test over the rendered
   `<head>` per tag kind would lock in "exactly one path emits"._

## Verified — measured, not assumed

**Layout**

- **Zero horizontal overflow at 375** on all three pages:
  `scrollWidth === clientWidth === 375`. Hard pass criterion met.
- **Footer anchoring**: 404 at 1280 → `footerBottom=800`, `viewportHeight=800`,
  gap **0**, page does not scroll. At 375 → `footerBottom=812`, gap **0**.
  Long pages (home, privacy) keep the footer in normal flow at both widths.
  Achieved by removing `.footer`'s redundant `margin-top: var(--space-8)` — the
  last `.section` already contributes 64px of `padding-block`, and that extra
  32px was exactly what pushed the 404 eleven pixels past the fold at 375.

**Accessibility**

- One `<h1>` per page; heading order valid on all three (home `[1,2,3,3,3,2]`,
  privacy `[1,2,2,2,2,2,2]`, 404 `[1]`).
- All four landmarks (`header`/`nav`/`main`/`footer`) present on every page.
  Skip link present and revealed on first Tab (`compliance-skiplink-focus.png`).
- Zero images missing `alt`.
- **External links: 3/3 on home, 4/4 on privacy fully marked** — `target="_blank"`
  **and** `rel="noopener noreferrer"` **and** a visible `↗` **and**
  "(opens in new tab)" in the accessible name.
- **Zero internal or `mailto:` links given `target="_blank"`** on any page.
- Contrast: **17/17 token pairs pass WCAG AA** in both themes (measured
  numerically in `DESIGN_TOKENS.md`; no new colors were introduced, so those
  measurements still hold). No axe-core run — the package is not installed
  locally; structural checks above were done by direct DOM measurement instead.
- `prefers-reduced-motion` honored in two places (`scroll-behavior`, cookie-bar
  animation).
- Fraunces and Inter both report loaded via `document.fonts.check` — no FOIT/FOUT.

**Behavior**

- **Esc now records a decline** (task 5): `localStorage` holds
  `{"status":"denied","version":1,"timestamp":…}`, the bar hides, and
  `barReturnsAfterReload = false`. The old behaviour re-prompted forever.
- **Theme toggle cycles all three modes** with a correct non-color cue and a
  state-plus-action accessible name at each step:
  `Auto → Dark → Light → Auto`, `data-theme` going `absent → dark → light →
  absent`, `aria-label` reading e.g. "Theme: Dark. Switch to Light."
- **Events fire with real values.** `window.__tag.kind === 'ga4'`,
  `typeof window.track === 'function'`. A click on the header CV link produced
  exactly one `file_download` with
  `{file_name: "Petr-Travkin-CV.pdf", file_extension: "pdf", location: "header"}`
  — the filename came from `render.Filename`, the same function that names the
  actual download, so it cannot drift. Toggling the theme produced one
  `theme_toggle {to: "dark"}`. The 404 declares
  `data-track-load="page_not_found"` with `data-track-path="$path"`.
- **Exactly one tag path emits.** Verified across three boots: `G-…` → gtag.js +
  `gtag('config')`, no container, **no** `<noscript>` iframe. `GTM-…` → container
  + iframe, **no** `gtag('config')`. Unset → nothing, and no cookie bar at all.
  `UA-1-1` and `OOPS-9` both refuse to boot with a message naming the variable.
- **Sitemap is honest and stable.** `<lastmod>2026-08-04</lastmod>` (the real
  date `cv.yaml` last changed, per git), never the build date. Two separate boots
  produce byte-identical XML — verified live and locked in by a test asserting
  both the body and the ETag.
- Every route smoke-tested: `/` 200, `/cv` 301, `/cv/download` 200 (PDF),
  `/cv/download.md` 200, `/privacy` 200, `/healthz` 200, `/robots.txt`,
  `/sitemap.xml`, `/llms.txt`, `/site.webmanifest`, `/favicon.ico`, a hashed
  static asset, and an unknown path → real **404**. No PII in logs.
- `gofmt` clean, `go vet` clean, `go fix` clean, `go test ./...` green (4 packages).

## Aesthetic Fidelity

The pass was required to be visually invisible except for two intended changes, and
it is. Comparing against the pre-existing `review-*.png` set: type scale, palette,
spacing rhythm, focus rings and dark-mode treatment are unchanged. **Zero new
design tokens, zero new colors** — the `↗` glyph inherits `currentColor` and `em`
sizing, and `.visually-hidden` is pure geometry.

The one judgement call worth naming: the external-link marker is a text glyph
(`↗`), not an icon system. That was chosen to match the site's existing vocabulary
— `→` in `.link-arrow`, `▾` in `.dl-caret` — rather than introduce an SVG set for
four links. In `compliance-404-desktop-1280.png` it reads as a quiet typographic
mark, which is the right register for Warm Minimalism.

## Deviations Carried Forward

Recorded decisions, not omissions — each with its rationale in the brief:

1. **i18n (§5) out of scope.** Single-locale site with no translated CV; the
   section presupposes multi-locale.
2. **Server-side Measurement Protocol deferred.** No lead form, so the CV download
   is the only conversion and the client layer covers it. Also keeps
   `GA4_API_SECRET` out of the deployment.
3. **Pre-commit minifier not added.** Minification runs in-process at boot, so
   served assets cannot be stale — the outcome the rule exists to guarantee.
4. **`checklist.md` left unticked.** It stays a shared conventions document.

## Blocked — not done, and not claimable as done

1. **Set the tag id in Railway `production`.** Currently unset: production has no
   `GA_ID`/`GTM_ID`, so the consent bar and analytics are absent from the live
   site today. Everything above was verified against a locally-injected
   `G-RH8KWHKMPZ`.
2. **Turn off GA4 Enhanced Measurement → "File downloads."** This pass sends its
   own `file_download`; leaving both on returns every download doubled.
3. **Verify each event once in GA4 DebugView.** Events were verified firing in the
   browser with the expected params, but never against the live property. Unit
   tests and a headless browser are not a substitute for DebugView.

## What Works Well

- **The typed `site.Tag` is the right shape for the problem.** Making "two tags at
  once" unrepresentable, rather than merely discouraged, is what turns a silent
  double-counting bug into something that cannot compile. The fatal-on-typo rule
  moves the failure to the deploy where someone is watching.
- **The render-to-bytes pipeline held up.** Every fix — content dates, tag
  resolution, event wiring — slotted into boot-time rendering without a single
  per-request template execution. The 404's `$path` substitution is the one place
  the model genuinely constrained the design, and it is documented where a future
  reader will find it.
- **The consent apparatus was already excellent** and needed only the Esc change:
  default-denied before the tag, non-modal `<aside>`, no focus trap, equal-weight
  buttons, versioned and timestamped storage, footer withdrawal path.
- **Truthfulness by construction.** `file_name` comes from the same function that
  names the download; `lastmod` comes from content and is provably stable across
  deploys; the JSON-LD credit is visible in the DOM. Each of these removes a way
  for the site to drift into lying.
- **`checklist.md` proved its worth as a review instrument.** The two most
  valuable findings of this whole pass — the tag id missing from production and
  the sitemap's build-time `lastmod` — were both invisible from inside the running
  site and surfaced only by reading the standing conventions against the code.
