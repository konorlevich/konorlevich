# Design Review: Cookie Consent Banner + Privacy Page

Reviewed against: `DESIGN_BRIEF.md`
Philosophy: **Warm Minimalism** (inherited)
Date: 2026-07-05
Build reviewed: `github.com/konorlevich/konorlevich` (Go `html/template` + `static/js/consent.js`), served at `http://localhost:8080/`. Screenshots and behavioral checks captured via headless Chrome over the Chrome DevTools Protocol (real device-metrics emulation; Google Analytics host nullrouted so captures render offline).

## Screenshots Captured

| Screenshot | Breakpoint | Description |
| ---------- | ---------- | ----------- |
| `screenshots/review-bar-desktop-1280.png` | Desktop (1280×800) | Home + consent bar, light — copy left / equal buttons right |
| `screenshots/review-bar-tablet-768.png` | Tablet (768×1024) | Home + consent bar, light |
| `screenshots/review-bar-mobile-375.png` | Mobile (375×812) | Home + consent bar, light — stacked, equal-width buttons |
| `screenshots/review-bar-focus-accept.png` | Desktop (1280×800) | Keyboard focus ring on the Accept button (non-blocking: page scrolled/tabbed underneath) |
| `screenshots/review-bar-dark-desktop-1280.png` | Desktop (1280×800) | Consent bar in **dark mode** (`prefers-color-scheme: dark`) |
| `screenshots/review-privacy-desktop-1280.png` | Desktop (1280×1400) | `/privacy` page, light |
| `screenshots/review-privacy-mobile-375.png` | Mobile (375×900) | `/privacy` page, light |

> All screenshots are in `.design/cookie-banner/screenshots/`.

## Summary

The build is a faithful, low-drama realization of the brief: a non-blocking Warm-Minimalism bar with two genuinely equal Accept/Decline buttons, an honest `/privacy` page, and — the part that actually matters for a GDPR feature — a **working consent gate** verified in a real browser. Contrast passes AA in both light and dark, there's zero horizontal overflow at any breakpoint, and the keyboard focus ring is visible. The findings below are refinements, not defects: the two worth acting on are a **semantic role mismatch** (`role="dialog"` on a deliberately non-modal bar) and **config duplicated across three files**. There are no must-fixes.

## Behavioral Verification (the crux for a consent feature)

Run in a real headless browser against the running site (not unit stubs):

| Check | Result |
| ----- | ------ |
| `consent 'default'` = `analytics_storage: denied` present in `dataLayer` | ✅ |
| …and ordered **before** `gtag('config')` | ✅ `defaultBeforeConfig: true` |
| No `_ga*` cookie present before consent | ✅ (see caveat in Could-Improve #1) |
| Click **Accept** → `consent 'update'` = `granted` pushed | ✅ |
| …choice persisted to localStorage (`status: granted`) + bar dismissed | ✅ |
| Reload → stored choice **replays** as `granted`, bar stays hidden | ✅ `replayGranted: granted` |
| Footer "Cookie settings" → bar reopens | ✅ |
| Change to **Decline** → `consent 'update'` = `denied` pushed + persisted | ✅ |

This exercises the entire state machine end-to-end in a real DOM with real clicks. The consent **signals** the gate depends on are all correct and correctly ordered.

## Resolution (applied 2026-07-05)

Both should-fixes were applied and re-verified in a real browser:
- ✅ **Should-fix #1 (semantics)** — dropped `role="dialog"` + `aria-modal="false"` on both templates; the bar is now a native `<aside>` complementary landmark labelled by `cookie-bar-title`. Re-checked: `role: null`, still reveals/hides correctly, focus never trapped.
- ✅ **Should-fix #2 (duplication)** — extracted the denied-default + replay + config (storage key, version, max-age, GA id) into one shared blocking bootstrap, `static/js/consent-default.js`, referenced by both pages; `consent.js` now reads `window.CONSENT_CONFIG` (defensive fallback only if the bootstrap fails to load). Re-verified: denied-default still fires **before** `config`, Accept→persist→reload-replay intact on both `/` and `/privacy`.

Could-improve items (live `_ga` test, no-JS footer button, SR announce, `/privacy` noindex) remain open for the owner.

## Must Fix

_None._ No broken functionality, no AA contrast failures, no major deviations from the brief.

## Should Fix

1. **`role="dialog"` + `aria-modal="false"` is a semantic mismatch for a non-modal bar.** `cv_template.html` / `privacy_template.html` mark the bar `role="dialog"`, but by design it does not trap focus, isn't modal, and (correctly) doesn't steal focus on auto-show. `dialog` sets assistive-tech expectations (modality, focus management) the bar intentionally doesn't meet. _Fix: drop `role="dialog"` and let the native `<aside>` be a `complementary` landmark with its existing `aria-label`/`aria-labelledby` — or, if you keep `dialog`, move focus into the bar when it appears. Given the non-blocking intent, the landmark is the better fit._

2. **Consent config is duplicated across three files.** The storage key (`cookie-consent`), `VERSION`, `MAX_AGE` (~12mo), the GA ID (`G-RH8KWHKMPZ`), and the entire `consent 'default' denied` block live inline in both `cv_template.html` and `privacy_template.html`, and again (key/version/max-age) in `static/js/consent.js`. A version bump or GA-ID change must be made in three places or the replay logic silently drifts. _Fix: centralize — e.g. extract the default-denied block into one small `/static/js/consent-default.js` referenced by both pages, and read `VERSION`/key from a single source. (The inline default must stay synchronous in `<head>`, but it can still be one shared file.)_

## Could Improve

1. **The live `_ga` cookie test was not observable in this environment.** Google's tag host is nullrouted here (offline capture), so `gtag.js` never actually ran — meaning "no `_ga` before consent" is confirmed by the correct denied-default signal, not by watching Google's own library set/withhold the cookie. _Before shipping: load the page in a real browser with network, confirm no `_ga*` cookie exists on load, click Accept, confirm `_ga`/`_ga_RH8KWHKMPZ` then appear, and that Decline keeps them away. Everything points to this working; it's the one thing not yet observed with real GA network._

2. **The footer "Cookie settings" button is inert without JavaScript.** With JS off it's a dead control. This is low-stakes (no JS → GA never loads and the bar never shows, so there's nothing to manage), but a dead button is slightly untidy. _Optional: render it only after `consent.js` wires it, or leave as-is._

3. **Screen-reader discoverability on auto-show.** Because the bar deliberately doesn't steal focus, SR users meet it only when they tab to the end of the DOM. That's an acceptable trade for non-modality, and it resolves naturally if you adopt Should-Fix #1. _Optional: a polite `aria-live` announcement if you want it flagged on appearance._

4. **`/privacy` is `noindex`.** Intentional-looking (keeps a thin legal page out of search), but if you'd rather it be discoverable, drop the `<meta name="robots" content="noindex">`.

## What Works Well

- **The gate is correct — and proven in a real browser.** Denied-by-default before config, granted-on-accept, replay-on-reload, and withdraw-on-decline all verified against the running site (see Behavioral Verification). This is the hard part of a consent feature and it's right.
- **Equal choice, no dark pattern.** Both buttons share identical computed styles (`stylesEqual: true` — same background, color, font, padding, min-height, radius); they differ only by label width. Deliberately not the terracotta primary CTA, so the site's "one real action" colour never nudges the consent decision. Exactly the brief's Principle 3.
- **AA contrast in both themes.** Light: title 15.0, description 5.9, policy link 5.8, button labels 14.1. Dark: 12.7 / 6.8 / 5.8 / 14.1. All ≥4.5:1.
- **Dark mode is a real theme, not an inversion.** The bar adapts to a warm charcoal surface with light, high-contrast buttons — a native citizen of the site's dark palette (`review-bar-dark-desktop-1280.png`).
- **Responsive reorganizes, not just shrinks.** Copy-over-buttons stacked on mobile → copy-left / buttons-right row at ≥768px. Zero horizontal overflow measured at 375 / 768 / 1280.
- **Accessible interaction.** Real `<button>`s, visible terracotta focus ring on keyboard focus (`review-bar-focus-accept.png`), 44px targets, Esc dismisses without counting as consent, `prefers-reduced-motion` respected, non-blocking so page content stays fully usable.
- **Honest privacy page.** Plain-language, specific (cookieless-until-accept, 2-year cookie lifetime, consent stored locally, Article 6(1)(a) basis, decline-is-equal), and stylistically indistinguishable from the rest of the site.
