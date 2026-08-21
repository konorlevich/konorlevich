# Build Tasks: Checklist Compliance Pass

Generated from: `.design/checklist-compliance/DESIGN_BRIEF.md`
Also reads: `INFORMATION_ARCHITECTURE.md`, `DESIGN_TOKENS.md`
Date: 2026-08-21

**Aesthetic philosophy: Warm Minimalism — inherited, not re-established.** No task in
this list introduces a token, a color, or a visual idea. Task 6 is the only one a
visitor should be able to see, and only because it corrects something already wrong.

**Ordering rationale.** Risk first, per the skill and per `checklist.md`'s own note that
the head-reorder is the risk-first task. Tasks 1–2 are the ones that fail *silently* if
done wrong (wrong tag emitted, or a sitemap that lies), so they go before anything
cosmetic. Task 6 is the one most likely to visibly break something, so it precedes the
trivial polish rather than trailing it.

**Ten tasks. Each is independently buildable and leaves the tree green** — `go build`,
`go vet`, `gofmt`, `go test ./...` all clean, server boots, every route answers.

---

## Foundation

- [x] **1. One tag, resolved and typed** — Replace `site.Config.GAID` + `GTMID` (two
      independent strings that can both emit) with a single resolved
      `site.Tag{Kind, ID}`. `KindGA4` for `G-…`, `KindGTM` for `GTM-…`, `UA-…`
      **rejected**, set-but-unrecognized prefix = **fatal boot error**, empty = clean
      no-op. `GTM_ID` is canonical; `GA_ID` and `GA4_ID` are accepted aliases resolving
      into the same field with precedence documented in the doc comment. Rewrite
      `consent.html` so **exactly one** tag path emits, branching on `.Kind` — never
      prefix-matching a raw string in the template. The default-denied Consent Mode v2
      block stays **shared by both branches and synchronous, before either tag**; the
      `<noscript>` iframe stays GTM-only. Update `NeedsConsent()` and the boot log field.
      _Modifies: `internal/site/site.go`, `templates/partials/consent.html`, `main.go`._
      _Tests: every prefix, alias precedence, UA rejection, fatal-on-garbage._
      **Done when:** `GA_ID=G-x` emits gtag.js and no iframe; `GTM_ID=GTM-x` emits the
      container and the iframe; both set resolves to one tag, never two; `UA-1` and
      `XX-1` both refuse to boot; unset emits nothing at all.

- [x] **2. Honest sitemap dates** — Add a `pages:` map to `cv.yaml` keyed by request
      path (`/`, `/privacy`, `/__404`), each with an `updated:` date. Add the type to
      `internal/cv`, **validated at boot**: every `pageSpec` must resolve to a non-zero
      date or the process does not start. `pageSpec` gains `updated`; `sitemapXML` reads
      it instead of `buildTime`. `buildTime` keeps `Last-Modified` and the cache
      validators — those are legitimately deploy-time.
      _Modifies: `cv.yaml`, `internal/cv/cv.go`, `internal/web/pages.go`,
      `internal/web/ops.go`._
      _Test (checklist-mandated): two boots with no content edit produce a
      **byte-identical** sitemap._
      **Done when:** `/sitemap.xml` carries `2026-08-21` for `/`, redeploying changes
      nothing, and deleting a date breaks the boot rather than the sitemap.

## Core UI

- [x] **3. `analytics.js` — consent and tracking in one bootstrap** — Rename
      `static/js/consent.js` → `analytics.js` and add the single
      `track(name, params)` helper beside the consent controller, where `checklist.md`
      says it belongs. `consent.html` renders `window.__tag = {kind: …}` from the
      **same** resolved `site.Tag` — no second source of truth, no prefix-matching in
      JS. `track()` branches on `kind`: `ga4` → `gtag('event', …)`, `gtm` →
      `dataLayer.push({event, …})`, absent → **silent no-op that never throws**. One
      delegated `document`-level listener reads `data-track` / `data-track-*` off the
      element actually used — no inline `onclick`, no per-component listeners. Event
      names as constants in this one module. Load unconditionally from `layout.html`
      (dropping the `NeedsConsent` gate on the script tag; the gate on the **bar
      markup** stays). Stays inside the JS budget.
      _New: the `track` layer. Modifies: `layout.html`, `consent.html`,
      `static/js/consent.js` → `analytics.js`._
      **Done when:** with no tag configured every call is a no-op and the console is
      clean; with `G-…` each call produces exactly one `gtag('event')`.
      _Note: no Go event constants — there is no server-side layer to share a
      vocabulary with. They arrive with Measurement Protocol if it is ever added._

- [x] **4. Wire the event catalog** — Attach all six events declaratively.
      `file_download` on both CV links (`file_name`, `file_extension`, `location`);
      `contact_click` on every `mailto:` (`method: mailto`, `location`);
      `select_content` on primary CTAs (`cta_label`, `location`); `theme_toggle` from
      `theme.js` on change (`to`); `consent_update` from the consent controller
      (`status`); `page_not_found` fired on the 404 from `location.pathname` — the one
      non-declarative event, because the 404 body is a single pre-rendered blob shared
      across every missed URL and cannot know its own path. `location` values come from
      the fixed vocabulary `header | hero | footer`. **No PII in any param** — never the
      email address. Fire on the **outcome**, and never let a download depend on an
      event.
      _Depends on: task 3. Modifies: `chrome.html`, `home.html`, `notfound.html`,
      `theme.js`._
      **Done when:** every action in the catalog produces exactly one hit with the
      expected params, and every download still works with JS disabled.

## Interactions & States

- [x] **5. Esc records a decline** — Change the consent bar's Esc handler from
      "dismiss without recording" to storing a real `denied` choice (versioned +
      timestamped, same path as the button) and emitting `consent_update`. Both are
      privacy-safe; the current behaviour re-prompts a keyboard user on **every visit
      forever**, which is the worse of the two. Update the code comment so the next
      reader sees the decision, not the old rationale. The bar must stay non-modal —
      no focus trap, no focus steal.
      _Depends on: task 3. Modifies: `static/js/analytics.js`._
      Covers: accept, decline, **Esc**, reopen via footer, expiry re-prompt, storage
      blocked.
      **Done when:** Esc closes the bar, it does not return on reload, and
      `analytics_storage` stays `denied`.

## Responsive & Polish

- [x] **6. Footer sits at the bottom** — `body { min-height: 100dvh; display: flex;
      flex-direction: column }` + `main { flex: 1 }`. Pure CSS, no JS, **not**
      `position: fixed`. **This is the task most likely to break something visible:**
      `body` already carries `background-attachment: fixed` with the `--glow` radial
      gradient and an `overflow-x: clip` guard, and becoming a flex container changes
      how children size. Verify the glow still renders identically and that no
      horizontal scroll appears.
      Breakpoints: **375 / 768 / 1280, on the 404 page specifically** (the shortest
      page), plus home and privacy for regression.
      _Modifies: `static/css/styles.css`. Reuses: `--space-*` only. No new tokens._
      **Done when:** the 404 footer sits at the viewport bottom with no dead space
      beneath it at 375 and 1280, long pages are unchanged, and
      `scrollWidth === clientWidth === 375`.

- [x] **7. External links announce themselves** — Add a shared
      `{{define "extlink"}}` partial: `target="_blank"` + `rel="noopener noreferrer"`
      + an `aria-hidden` external-link icon + a visually-hidden "(opens in new tab)" in
      the accessible name. Add the `.visually-hidden` utility (clip-based, **never**
      `display: none` — the text must reach the accessibility tree). Route every
      off-domain link through it: footer social links and the Soarline Studio credit.
      **Internal links stay in the same tab; `mailto:` gets no `target` at all.**
      Enforced in the one partial, not per-template.
      _New: `extlink` partial, `.visually-hidden`. Modifies: `chrome.html`,
      `styles.css`. Icon inherits `currentColor`, sized in `em` — no new tokens._
      **Done when:** every off-domain `href` on every page has both attributes plus a
      cue, no internal link has `target`, and the footer row does not wrap badly at 375.

- [x] **8. Portrait resists casual saving** — `draggable="false"` on the hero portrait
      plus `-webkit-user-drag: none; user-drag: none; user-select: none;` scoped to that
      image. **No** `pointer-events: none`, **no** document-wide right-click or
      selection blocking, **no** contextmenu suppression (not requested). `alt` text and
      `og:image` stay intact. This is friction, not protection, and the brief says so.
      _Modifies: `home.html`, `styles.css`._
      **Done when:** the portrait cannot be dragged to the desktop, right-click and text
      selection work normally everywhere, and the image is unchanged for screen readers
      and crawlers.

## Verify

- [x] **9. Green-build gate and route smoke test** — `go build`, `go vet`, `go fix ./...`,
      `gofmt -l` clean, `go test ./...` passing. Boot the server and smoke-test every
      route: `/`, `/cv` (301), `/cv/download`, `/cv/download.md`, `/privacy`, `/healthz`,
      `/robots.txt`, `/sitemap.xml`, `/llms.txt`, `/site.webmanifest`, `/favicon.ico`,
      a `/static/…` asset, and an unknown path for a real 404. Confirm no PII in logs.
      Boot twice and diff `/sitemap.xml` — must be byte-identical.
      **Done when:** all four Go checks are clean and every route answers as specified.

- [x] **10. Design review** — Run `/design-review` against the brief. Capture
      375 / 768 / 1280 × light/dark, the 404 footer at 375 and 1280, consent bar
      keyboard states, and the external-link focus state. Target **0 axe-core
      violations**.

---

## Blocked — cannot be completed from the codebase

These are **not** done and must not be reported as done. They need access I do not have.

- [ ] **A. Set the tag id in Railway `production`.** Currently unset — which is why the
      consent bar and analytics are absent from production today. Until this is set,
      tasks 3 and 4 ship inert.
- [ ] **B. Turn off GA4 Enhanced Measurement → "File downloads."** This pass sends its
      own `file_download`; leaving both on returns every download **doubled**.
- [ ] **C. Verify each event once in GA4 DebugView** — exactly one hit per action, the
      expected params, no Enhanced Measurement duplicate, and a JS-disabled download
      still counted correctly. Unit tests do not substitute for this.

## Definition of done

- All eight audited violations closed, or explicitly re-scoped in writing.
- Zero new design tokens; zero new colors. A screenshot diff against today shows
  **only** the footer change and the external-link cue.
- Contrast still 17/17 AA in both themes; 0 axe violations.
- Zero horizontal overflow at 375, measured programmatically.
- Two consecutive boots produce a byte-identical `/sitemap.xml`.
- `go build` / `go vet` / `go fix` / `gofmt` / `go test` all clean.
- The three deviations (i18n, server-side Measurement Protocol, pre-commit minifier)
  remain documented decisions in the brief, each with its rationale.
- Blocked items A–C are handed over as explicit deploy steps, not silently dropped.
