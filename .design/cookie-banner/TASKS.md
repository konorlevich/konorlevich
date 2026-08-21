# Build Tasks: Cookie Consent Banner + Privacy Page

Generated from: `.design/cookie-banner/DESIGN_BRIEF.md`
Date: 2026-07-05
Stack: Go `html/template` + `static/` assets, extending `konorlevich.tech`. Philosophy: **Warm Minimalism** (inherited from `.design/personal-webpage/`).

> Each task is a vertical slice (structure + styling + behavior together), independently buildable and verifiable against the brief. Ordered **risk-first**: the actual consent *gate* is proven before the visible bar, because a pretty banner over ungated analytics is the one outcome worse than no banner. Work top-to-bottom; check off after confirming.

## Context from codebase scan (read before starting)
- **GA fires unconditionally**: `cv_template.html:4–12` loads `gtag/js?id=G-RH8KWHKMPZ` and calls `gtag('config', …)` in `<head>` with no gate. The `<script async src=gtag>` is currently the **first** thing in `<head>` — the Consent Mode default MUST run *before* it, so the head needs reordering.
- **No JS on the site yet**: `static/` has only `css/`, `fonts/`, `img/`. This feature introduces the first `static/js/` file. Add it to the existing `GET /static/` file server (`main.go:112`) — no new static route needed.
- **Buttons exist**: `.btn`, `.btn-primary`, `.btn-ghost`, `.btn-sm` in `static/css/styles.css:98–134`. Equal-weight Accept/Decline can be two peers (e.g. both `.btn-ghost`, or `.btn-primary` + a same-sized `.btn-ghost`) — resolved in the bar task.
- **Tokens**: reuse `static/css/tokens.css` as-is. No new tokens (`--color-bg-secondary`, `--shadow-lg`, `--card-radius`, `--shadow-focus`, `--space-*`).
- **Footer**: `cv_template.html:201–210`, `.footer-links` nav — where the "Cookie settings" withdraw link goes.
- **Routes**: `main.go:111–116` (`GET /`, `GET /cv`, `GET /cv/download`, `POST /contact`). Templates parsed once at startup. `/privacy` needs a new `GET /privacy` handler + a template.

---

## Foundation

- [x] **Consent gate — Consent Mode v2 default + head reorder (RISK-FIRST, the real compliance mechanism)**: Insert a synchronous `gtag('consent','default',{ analytics_storage:'denied', ad_storage:'denied', ad_user_data:'denied', ad_personalization:'denied' })` block **before** the `gtag/js` script in `<head>`, and replay any stored choice into a `consent`/`update` call before `gtag('config', …)`. "Done" = on a fresh load with no stored choice, **GA sets no `_ga*` cookies** (verify in devtools Application → Cookies) and only cookieless pings go out; a simulated `granted` value results in `_ga` cookies. _Modifies `cv_template.html` head. This is the task the whole feature exists for — build and verify it first._

- [x] **Consent state manager (`static/js/consent.js`)**: New JS module owning localStorage: read/write `{status, version, timestamp}`, treat a choice older than ~12 months (or a bumped `version`) as absent, and expose a tiny API (`getChoice()`, `setChoice(status)` → writes storage + calls `gtag('consent','update',{analytics_storage: status==='granted'?'granted':'denied'})`, `shouldShow()`). "Done" = calling the API from the console flips GA cookies on/off and persists across reloads; expired/absent choice reports `shouldShow() === true`. _New; depends on the gate's Consent Mode contract._

## Core UI

- [x] **Consent bar (ESTABLISHES WARM MINIMALISM inheritance)**: Build the non-blocking bottom bar — a warm card (`--color-bg-secondary`, `--shadow-lg`, `--card-radius`, hairline `--color-border-primary`) with one plain sentence ("This site uses Google Analytics to understand traffic. …") and a `/privacy` link (`--color-text-link`). Semantic labeled region (`role="dialog"` + `aria-label="Cookie consent"`), **not** focus-trapping. Hidden by default; shown only when `shouldShow()` is true. "Done" = the bar looks like a native part of the page (not a third-party widget), sits quietly at the bottom, and content above stays fully usable. _New; reuses tokens + `.container`. Validate the aesthetic here before wiring behavior._

- [x] **Accept / Decline buttons — equal weight (overrides "terracotta = the one CTA")**: Two peer buttons of identical size/prominence, extending `.btn`. No visual nudge toward Accept (Principle 3 / GDPR). Wire each to `consent.js`: Accept → `setChoice('granted')`, Decline → `setChoice('denied')`, both then dismiss the bar. "Done" = clicking Accept sets `_ga` cookies and hides the bar; Decline hides it and sets none; neither button is visually favored. _New (extends `.btn`); depends on consent state manager._

- [x] **`/privacy` page + `GET /privacy` route**: New route in `main.go` + a template rendering an honest, plain-language notice — what Google Analytics collects, retention, legal basis (consent), and how to withdraw (the footer "Cookie settings" link). Styled in Warm Minimalism, reusing the site header/footer shell. "Done" = `/privacy` renders on warm paper, is linked from the bar, and reads like the rest of the site. _New; parses at startup like the existing template._

## Interactions & States

- [x] **Consent lifecycle + footer "Cookie settings" withdraw link**: Add a `.footer-links` item that reopens the bar reflecting the current choice, so withdrawing is as easy as granting. Cover the full lifecycle. Covers: first-visit show, accept, decline, return-visit (bar stays hidden), reopen-from-footer, change-of-mind (decline-after-accept stops new GA cookies). _Modifies `cv_template.html` footer (`:201`) + `consent.js`._

- [x] **Edge, keyboard & motion states**: Robust behaviors — `prefers-reduced-motion` replaces the slide-in with a plain appear; logical tab order (Accept → Decline → policy link) with visible `--shadow-focus` rings; Esc does **not** grant consent (bar simply dismisses as no-choice/denied-safe or stays); expired-consent re-prompt path. Covers: reduced-motion, focus order, keyboard activation, Esc, expiry. _Modifies bar markup + `consent.js`._

## Responsive & Polish

- [x] **Responsive pass**: Mobile (≤767px) full-width bar, text on its own line, Accept/Decline as an equal-width row (or equal stacked) with 44px targets, respecting `env(safe-area-inset-bottom)`; desktop (≥768px) contained card with text + both buttons inline, buttons equal to each other. Breakpoints: sm 375 / md 768. "Done" = no horizontal overflow at 375px; buttons stay equal at every width. _Cross-component._

- [x] **Accessibility & no-dark-patterns pass**: WCAG **AA** contrast on all bar text + both button labels + the policy link on the bar surface; labeled non-modal region announced correctly; ≥44px touch targets; visible focus everywhere; confirm Accept and Decline are genuinely equal (size, contrast, position) — the legal *and* a11y requirement. Checks pulled from brief → Accessibility Requirements. _Cross-component._

## Review

- [x] **Design review**: Run `/design-review` against the brief once the bar + `/privacy` are built and the server is running (captures the bar's responsive + interactive states, and verifies GA cookies are actually gated, into `.design/cookie-banner/screenshots/`).
