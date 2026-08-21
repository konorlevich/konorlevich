# Design Review: Personal Webpage — Petr Travkin

Reviewed against: `DESIGN_BRIEF.md`
Philosophy: **Warm Minimalism**
Date: 2026-07-04
Build reviewed: `github.com/konorlevich/konorlevich` (Go `html/template` + `cv.yaml`), served at `http://localhost:8080/`

## Screenshots Captured

| Screenshot | Breakpoint | Description |
| ---------- | ---------- | ----------- |
| `screenshots/review-home-desktop-1280.png` | Desktop (1280×800) | Full page, light theme |
| `screenshots/review-home-tablet-768.png` | Tablet (768×1024) | Full page, light theme |
| `screenshots/review-home-mobile-375.png` | Mobile (375×812) | Full page, light theme |
| `screenshots/review-home-dark-desktop-1280.png` | Desktop (1280×800) | Full page, **dark** (`prefers-color-scheme: dark`) |

> All screenshots are in `.design/personal-webpage/screenshots/`. Captured via Chrome DevTools Protocol at exact viewports with full-page capture.

## Summary

The build is a faithful, polished realization of Warm Minimalism and delivers on the brief's core strategy — one warm entrance, two clear doors (Experience / Work), with the email CTA reachable from every section. Responsive behavior genuinely *reorganizes* (hero stacks, experience 1→2 columns, projects 1→2→3) rather than just shrinking, and there's no horizontal overflow at 375px. **The biggest finding: dark mode is live (via a `prefers-color-scheme` block in `tokens.css`) even though the brief scoped v1 as light-only — and it ships with a visible rendering bug (a gray smudge across the hero).** That plus one contrast failure are the must-fixes.

## Resolution (applied 2026-07-04)

All three must-fixes were applied and re-verified (see re-captured `screenshots/review-home-dark-desktop-1280.png`):
- ✅ **#1 Hero glow** — now driven by a theme-aware `--glow` token; the dark-mode smudge is gone.
- ✅ **#2 Dark mode** — decision: **embraced as an intended feature** (light remains default). Brief's Out-of-Scope line + tokens note updated.
- ✅ **#3 Contrast** — `--color-text-tertiary` darkened to `#6f685c` (light, 5.16:1) and `#94897d` (dark, 5.11:1); both pass AA.
- ✅ Bonus: `white-space: nowrap` added to `.btn` (fixes the tablet "Let's talk" wrap).

Should-fix (copy) and Could-improve items remain open for the designer.

## Resolution 2 (applied 2026-07-05 — polish pass)

Remaining Should-fix / Could-improve items addressed:
- ✅ **Should-fix #2 (copy)** — the truncated "…cut API response times." line was already fixed to "cutting API response time 10× — from 2s to 0.2s." The leftover `TODO` on the current Payment-gateway role was removed and its achievement rewritten as a complete, confident line (no fabricated metrics).
- ✅ **Could-improve #1 (real screenshots)** — captured live hero shots of all four project sites (kidsspace.ge, ai-news.ge, speakadoo.club, tgchathelperbot.tech) via headless Chrome, optimized to 960×600 `jpg` + `webp` siblings under `/static/img/proj-*`, wired into `cv.yaml`, and the project card upgraded to `<picture>` (webp with jpg fallback). Placeholders no longer shown.
- ✅ **Could-improve #2 (OG image)** — added a brand-faithful 1200×630 `/static/img/og.png` (rendered with the site's own Fraunces/Inter/JetBrains Mono + Warm Minimalism palette). `og:image`/`twitter:image` (absolute URLs), `og:url`, `og:image:alt`, and `twitter:card: summary_large_image` added to the template head.
- ✅ **Could-improve #4 (photo)** — already present (`static/img/photo.{jpg,jpeg,webp}` via `<picture>`); confirmed rendering.
- ✅ **Could-improve #3 (facts chips 2-up ≤479px)** — added a `@media (max-width: 479px)` rule: `.facts` becomes a 2-col grid with the long availability chip spanning its own full-width row. _Note: could not be screenshot-verified because the local headless Chrome pins the CSS layout viewport to ~600px min; the rule is served and standard-correct._
- ✅ **Bonus — PDF alignment** — `/cv/download` now includes the Location · Availability · Languages meta line, the summary paragraph, and a full **Projects** section (name, badge, clickable URL, description). Verified in the generated PDF text.

Build: `go build` clean, `gofmt` clean, all endpoints verified (page, `/static/img/og.png`, project images, PDF).

## Must Fix

1. **Dark-mode hero "smudge" — hardcoded light glow doesn't adapt.** `static/css/styles.css` `body { background-image: radial-gradient(120% 60% at 50% -10%, #fffdf9 …) }` uses a hardcoded near-white color. In dark mode it renders as a harsh gray blob across the top of the page. See `screenshots/review-home-dark-desktop-1280.png` (top third). _Fix: drive the glow from a token (e.g. a new `--glow` variable set per theme), or wrap it in `@media (prefers-color-scheme: light)`, or remove it. Tie this to the dark-mode decision below._

2. **Dark mode is shipping despite the brief scoping v1 as light-only — make it a decision, not an accident.** `static/css/tokens.css` includes a `@media (prefers-color-scheme: dark)` override, so every visitor whose OS is set to dark (a large share) sees the dark theme. It actually looks good (warm charcoal, lightened terracotta, off-white text — a real palette, not an inversion), but right now it's unintended per the brief and carries bug #1. _Fix: pick one — **(a)** honor the brief: neutralize dark for v1 by forcing the light palette (drop/guard the `prefers-color-scheme` block), or **(b)** embrace dark as an intended feature, fix bug #1, and update the brief's "Out of Scope" line. Recommendation: **(b)** — it's already 90% there and strengthens the page._

3. **`--color-text-tertiary` fails WCAG AA (2.91:1).** Measured: `#9a9184` on paper `#faf7f2` = **2.91:1** (needs 4.5:1 for small text). It's used for small text: `.skill-cat` labels, `.footer-note`, and the project placeholder. The brief requires AA. _Fix: darken the token to ≥4.5:1 — verified value **`#6f685c`** = 5.16:1. (Dark-mode tertiary `#7d746a` = 3.81:1 also only passes for large text; darken slightly if used on small text.)_

## Should Fix

1. **"Let's talk" button wraps to two lines at tablet.** In the `.work-nudge` row layout (≥768px), the button column is tight and "Let's talk" breaks to "Let's / talk". See `screenshots/review-home-tablet-768.png` (work-with-me nudge). _Fix: add `white-space: nowrap` to `.btn` (and/or give the nudge button column `flex-shrink: 0`)._

2. **Confirm the copy before ship.** The achievements/tagline are a strong first draft, but one line still reads "…cut API response times." with no figure (`cv.yaml`, Linkprofit). _Fix: fill the real metric or reword so it doesn't read as a truncation._

## Could Improve

1. **Replace project placeholders with real screenshots.** The diagonal-stripe monogram placeholders look intentional, but the brief calls proof of built work the single strongest asset for the freelance door. Real thumbnails of kidsspace.ge / ai-news.ge / tgchathelperbot.tech would convert far harder. _Add images under `/static/img` and set `screenshot:` in `cv.yaml`._

2. **Add an OG preview image.** `<meta name="twitter:card" content="summary">` and OG tags are present, but there's no `og:image`, so shared links render as text-only cards. Since both audiences arrive by sharing the link, a 1200×630 image is high-leverage. _Add `/static/img/og.png` + `<meta property="og:image">`._

3. **Facts chips stack to three lines at 375px.** Not a bug (they wrap cleanly), but a 2-up arrangement would tighten the hero on small phones. _Optional._

4. **Add the photo when ready.** The "PT" avatar fallback is genuinely good (works in both themes), so there's no urgency — but "warm & human" peaks with a real face. The hero already supports it via `photo:` in `cv.yaml`.

## What Works Well

- **Aesthetic fidelity is high.** Fraunces display serif, restrained terracotta accent (used only for CTAs / links / company names), warm paper, and mono technical accents (eyebrows, dates, tags) read unmistakably as the intended Warm Minimalism. Nothing generic-AI about it.
- **Visual hierarchy is clear.** Name → tagline → CTA is the obvious first read; the two doors (Experience, Work) are distinct sections with their own eyebrows.
- **The CTA strategy is executed.** "Email me" appears in the top bar, hero, work nudge, and contact — "one action, always reachable" from the brief, delivered.
- **Responsive reorganizes, not just shrinks.** Verified across 375 / 768 / 1280: hero stacks→side-by-side, experience 1→2 columns, projects 1→2→3, no horizontal scroll at 375.
- **Accessibility foundation is solid.** Semantic landmarks (`header`/`main`/`footer`/`nav`), visible focus rings (`--shadow-focus`), `prefers-reduced-motion` handling, meaningful `alt`/`aria-hidden`, 44px CTA targets — all present. Primary/secondary text and CTAs pass AA (only tertiary fails, see Must-Fix 3).
- **Dark mode is a bonus, not an inversion.** Once bug #1 is fixed, it's a shippable warm dark theme with correct accent lightening and AA-passing primary/secondary text.
