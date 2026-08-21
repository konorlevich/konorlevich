# Design Review: Checklist-Compliance Refactor

Reviewed against: `.design/personal-webpage/DESIGN_BRIEF.md`
Philosophy: **Warm Minimalism**
Date: 2026-08-03

## Screenshots Captured

| Screenshot | Breakpoint | Description |
| ---------- | ---------- | ----------- |
| `screenshots/review-home-desktop-1280.png` | Desktop (1280×900) | Home, full page, light |
| `screenshots/review-home-tablet-768.png` | Tablet (768×1024) | Home, two-column hero collapses |
| `screenshots/review-home-mobile-375.png` | Mobile (375×812) | Home, single column |
| `screenshots/review-home-dark-desktop-1280.png` | Desktop (1280×900) | Home in dark mode |
| `screenshots/review-home-light-desktop-1280.png` | Desktop (1280×900) | Home with light forced |
| `screenshots/review-privacy-*.png` | 375 / 768 / 1280 | Privacy & cookies page |
| `screenshots/review-404-*.png` | 375 / 768 / 1280 | The new 404 page |
| `screenshots/review-home-skiplink-focus.png` | Desktop | Skip link revealed on first Tab |
| `screenshots/review-consent-focus-accept.png` | Desktop | Focus ring on the Accept button |
| `screenshots/review-theme-toggle.png` | Desktop | Theme toggle in the header |

> All screenshots are in `.design/checklist-compliance/screenshots/`.

## Summary

The refactor moved the site from "renders correctly" to "renders correctly and is
built the way the standing conventions describe" — embedded assets, everything
pre-rendered and precompressed at boot, real 404s, a full ops-route and SEO
suite. The Warm Minimalism layer came through untouched: same palette, same
type, same restraint, verified identical at all three breakpoints.

**The biggest finding was my own.** The first pass read the stale
`NOTE ON DARK MODE` comment in `tokens.css` (which still said "out of scope for
v1") and deleted the `prefers-color-scheme` block as an auto-activating-dark
regression. The brief's line 97 says the opposite — dark mode was promoted to a
shipped feature in the 2026-07-04 post-review update. Deleting it removed a real
feature. It is now restored *and* given the three-mode toggle the conventions
require, which is what the checklist actually asks for when both themes exist.

Measured, not eyeballed: **0 axe violations** on all three pages in both themes,
**0 horizontal overflow** at 375/768/1280, and **every measured text pair passes
WCAG AA in both themes** (worst case 5.16:1 light, 6.10:1 dark).

## Must Fix

_None outstanding._ Both defects found during the review were fixed in this pass:

1. ~~**Dark mode was deleted.**~~ Restored in `static/css/tokens.css` with the
   palette values defined once as raw `--dark-*` tokens and two mapping blocks
   (explicit `[data-theme="dark"]`, and `:root:not([data-theme])` under
   `prefers-color-scheme`). The second path means **dark still works with
   JavaScript disabled** — the toggle is an enhancement, not a dependency.
2. ~~**`.eyebrow` was undefined CSS.**~~ `templates/pages/privacy.html` used
   `class="eyebrow"`, which is not defined anywhere in `styles.css` — a
   pre-existing bug that rendered the "Privacy & cookies" label as unstyled body
   text, and which I had propagated to the new 404 page. Both now use the
   existing `.section-eyebrow` (mono, uppercase, letterspaced, accent), reusing
   the established class rather than adding a new one.

## Should Fix

1. **The theme toggle is JS-gated by design, and that is a real trade-off.** The
   button ships `hidden` and `theme.js` reveals it, because a cycle that cannot
   persist is worse than no control. Visitors with JS off keep system-preference
   dark but cannot override it. Acceptable — but worth a conscious confirmation
   rather than an accident. See `screenshots/review-theme-toggle.png`.
2. **`og:image` is still the light-theme card.** `static/img/og.png` was made for
   the light palette. Now that dark ships, a viewer in a dark client sees a light
   preview card. Cosmetic and low-traffic, but a dark variant behind
   `prefers-color-scheme` is not possible for OG — pick one deliberately.

## Could Improve

1. **`--dark-*` mapping is duplicated across two blocks.** The *values* live once,
   but the ~24-line semantic mapping is repeated for the explicit and the
   system-preference path. `light-dark()` would collapse both into a single
   declaration per token; it is deliberately not used here because its browser
   floor (2024) is newer than this site's, and a fallback failure would silently
   break colour rather than degrade. Revisit when the floor moves.
2. **Skills categories use a fixed `min-width: 9rem` label column.** It holds at
   all three breakpoints today, but a longer future category name would wrap
   awkwardly. A `minmax()` grid would be more durable.
3. **The mobile theme toggle drops its text label** below 480px (icon only,
   with the state still exposed via `aria-label`). The half-disc icon carries
   the meaning acceptably, but the written mode is the stronger cue — worth
   revisiting if header space can be found.

## What Works Well

- **The aesthetic survived a structural rewrite intact.** Templates were split
  into a layout plus five partials and every page re-rendered through a new
  pipeline, yet the visual output at all three breakpoints is unchanged from the
  pre-refactor build. That is the payoff of having had real tokens.
- **Dark mode is genuinely designed, not inverted.** Warm charcoal `#1c1917`
  against warm off-white `#ede6db`, with the terracotta accent lightened to
  `#e08a63` specifically so it clears AA *as text* — the exact trap the
  conventions warn about for warm accents. Shadows were re-authored darker and
  more transparent rather than reused from light.
- **Consent architecture is the strongest part of the build.** Denied-by-default
  Consent Mode v2 in an inline synchronous head script, a non-modal `<aside>`
  that never traps focus, Esc-dismisses-without-granting, and equal-weight
  Accept/Decline. With `GA_ID` unset the site now ships **zero** JavaScript and
  sets **zero** cookies — the banner disappears entirely rather than asking about
  something that isn't there.
- **The 404 reads in the site's own voice.** "That page isn't here — the link may
  be old, or I may have moved something. Nothing's broken on your end." Warm,
  plainspoken, first person, and it takes responsibility instead of blaming the
  visitor. Exactly the brief's tone.
- **Accessibility holds up under measurement.** Skip link lands on the first Tab
  and is genuinely visible; the `--shadow-focus` ring is applied verbatim
  everywhere; touch targets are ≥44px; `prefers-reduced-motion` zeroes both
  duration and translate distance.
