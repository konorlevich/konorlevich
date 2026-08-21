# Design Brief: Cookie Consent Banner + Privacy Page

Feature slug: `cookie-banner`
Date: 2026-07-05
Stack: Go `html/template` + `static/` assets, extending the existing `konorlevich.tech` personal site.
Reuses the visual system from `.design/personal-webpage/` — **Warm Minimalism**.

## Problem

Someone lands on the CV site from an EU-based recruiter's LinkedIn share. The moment the page loads, Google Analytics (`G-RH8KWHKMPZ`) fires from `<head>` and drops analytics cookies — **before the visitor has agreed to anything**. There's no notice, no choice, no way to say no, and no record of consent. Under GDPR/ePrivacy that's unlawful tracking. The human friction is quiet but real: a visitor's activity is being measured without them being asked, and the site owner is carrying legal risk on a page whose entire job is to make a trustworthy first impression.

## Solution

A calm, non-blocking consent bar that appears once at the bottom of the page. It explains, in one plain sentence, that the site uses Google Analytics, and offers **two equally-weighted choices: Accept or Decline**. Analytics stays *off by default* — the Google tag loads in Consent Mode v2 "denied" state and only begins setting cookies if the visitor explicitly accepts. The choice is remembered so the bar never nags. A short **`/privacy`** page, linked from the bar, honestly describes what Google Analytics collects and how to change your mind. A quiet **"Cookie settings"** link in the footer reopens the bar at any time, so withdrawing consent is exactly as easy as giving it.

The visitor never has to think about this to use the site. The recruiter reads the CV; the bar waits at the bottom; nothing is blocked. Compliance is achieved by *behavior* (the gate), not by a wall.

## Experience Principles

1. **Honest gate over decorative notice** — The banner's job is the actual consent decision, not a "we use cookies 🍪 OK" acknowledgment. Decline must genuinely stop tracking, and it must be as easy as Accept. No implied consent, no pre-checked anything, no cookie wall.
2. **Quiet by default over attention-grabbing** — This is a personal CV site; the content is the star. The bar is present and reachable but non-blocking, low-drama, and dismissed forever after one choice. It must never compete with the hero or feel like an ad.
3. **Equal choice over nudged choice** — Accept and Decline carry the same visual weight. This deliberately overrides the site's "terracotta = the one primary CTA" habit: here, steering the user toward Accept would be a GDPR dark pattern. Neutrality is the correct design, not a compromise.

## Aesthetic Direction

- **Philosophy**: **Warm Minimalism** (inherited — see `.design/personal-webpage/DESIGN_BRIEF.md`). The banner is a native citizen of the existing page, not a bolted-on third-party widget.
- **Tone**: Calm, plain-spoken, respectful. Reads like a considerate host, not a legal department or a growth hack.
- **Reference points**: The restraint of GOV.UK's cookie banner (plain language, equal buttons, no manipulation); the warmth of the site's own hero card and project cards.
- **Anti-references**: The typical SaaS/CMP banner — full-screen dark-overlay modals, a giant glowing "Accept All" next to a greyed-out "Manage", flashing borders, "we value your privacy" boilerplate. Nothing that feels like Cookiebot / OneTrust / a marketing interruption.

## Existing Patterns

The banner must reuse the established system in `static/css/tokens.css` — no new tokens.

- **Typography**: Body `--font-family-body` (Inter) at `--font-size-sm` (14px) for banner text; `--font-family-mono` (JetBrains Mono) available for a small eyebrow if desired. Buttons follow the existing `.btn` conventions.
- **Colors**: Card surface `--color-bg-secondary` (`#ffffff`) lifting off the paper, hairline `--color-border-primary` (`#e6ddce`), text `--color-text-primary` / `--color-text-secondary`, links `--color-text-link` (terracotta `#a8482a`). Accent `--color-accent-primary` (`#a8482a`) used *sparingly and equally* per Principle 3.
- **Spacing**: `--space-*` scale (base `--space-4` = 8px); banner padding around `--space-6`/`--space-7`.
- **Elevation / shape**: `--shadow-lg` for the bar's lift off the page; `--card-radius` (`--border-radius-lg`) for its corners; `--shadow-focus` for focus rings.
- **Components**: The existing **`.btn`** component (with `white-space: nowrap`, 44px touch target, hover/focus/active states) is the starting vocabulary for both banner buttons. The footer (`cv_template.html:201`) and its `.footer-links` nav are where the "Cookie settings" link lives.

## Component Inventory

| Component | Status | Notes |
| --------- | ------ | ----- |
| Consent bar container | New | Non-blocking, fixed to bottom, `role="dialog"` + `aria-label`, warm card styling. Hidden once a choice is stored. |
| Banner copy + GA disclosure line | New | One plain sentence + link to `/privacy`. |
| Accept button | New (extends `.btn`) | Equal weight to Decline. Grants `analytics_storage`. |
| Decline button | New (extends `.btn`) | Equal weight to Accept — same size/prominence. Keeps `analytics_storage` denied. |
| Consent Mode v2 default block | New (JS in `<head>`) | Sets `analytics_storage: 'denied'` **before** `gtag/js` loads; reads stored choice and replays it. This is the real gate. |
| Consent state manager | New (small JS) | localStorage read/write (versioned + timestamped), show/hide bar, wire buttons, `gtag('consent','update',…)`. |
| "Cookie settings" footer link | Modify `cv_template.html` footer | Reopens the bar so consent can be withdrawn/changed. |
| `/privacy` page + route | New | Lightweight page: what GA collects, retention, legal basis, how to withdraw. New `GET /privacy` handler in `main.go`. |

## Key Interactions

- **First visit (no stored choice):** Consent Mode defaults to `denied` (set in `<head>` before gtag). The page renders fully; GA loads but sets no cookies (cookieless pings only). The bar slides up from the bottom after paint (respecting `prefers-reduced-motion` → no slide, just appear). Focus is *not* trapped — the page stays usable.
- **Click Accept:** Store `{status:'granted', version, timestamp}` in localStorage → `gtag('consent','update',{analytics_storage:'granted'})` → GA begins normal (cookie-based) measurement → bar dismisses (fade/slide out) → focus returns to the page.
- **Click Decline:** Store `{status:'denied', …}` → consent stays denied (no update needed, or explicit `'denied'`) → bar dismisses → no analytics cookies are ever set.
- **Return visit:** Stored choice is read in `<head>` and replayed into Consent Mode *before* gtag configures; the bar does **not** appear.
- **Withdraw / change mind:** Footer "Cookie settings" link re-shows the bar with the current choice reflected; choosing the opposite updates Consent Mode and, on decline-after-accept, GA stops setting new cookies going forward.
- **Consent expiry:** If the stored `timestamp` is older than ~12 months (or `version` is bumped after a policy change), treat as no choice and re-prompt.
- **Keyboard:** Tab reaches Accept then Decline then the policy link; Enter/Space activate; visible `--shadow-focus` ring on each; Esc does **not** count as consent (bar simply stays until an explicit choice, or closes without granting — decline-equivalent is safer: leaving it dismissed keeps denied).

## Responsive Behavior

- **Mobile (≤767px):** Full-width bar pinned to the bottom, text on its own line, the two buttons side-by-side on a row beneath (equal width) — or stacked full-width if space is tight, Accept and Decline still identical in size. Comfortable 44px targets. Must not obscure the persistent bottom content; sits above safe-area inset (`env(safe-area-inset-bottom)`).
- **Tablet/Desktop (≥768px):** Bar becomes a contained card (max-width aligned to `--max-width-page`, or a bottom-left/full-width bottom strip), text and both buttons on one row with the buttons right-aligned but equal to each other.
- **Behavior change, not just size:** buttons move from inline-row (desktop) to their own row (mobile). No horizontal overflow at 375px.

## Accessibility Requirements

- **Contrast**: All banner text and both button labels meet **WCAG AA** (≥4.5:1 small text). Terracotta accent on white and white-on-terracotta both already pass in the token system; verify the chosen button treatment. The policy link uses `--color-text-link` (AA on the bar surface).
- **Semantics**: Bar is a labeled region (`role="dialog"` / `aria-label="Cookie consent"` or a `<section aria-label>`), announced but **not** modal/focus-trapping (it's non-blocking). Buttons are real `<button>`s with clear text ("Accept", "Decline" — not icons).
- **Keyboard**: Fully operable; logical tab order; visible focus ring (`--shadow-focus`) on every interactive element; no keyboard trap.
- **Motion**: Honor `prefers-reduced-motion` — replace slide-in with a plain appear.
- **Targets**: ≥44px touch targets on both buttons and the footer link.
- **No dark patterns** (accessibility *and* legal): equal button prominence, no color/size trickery to steer toward Accept.

## Out of Scope

- **Granular per-category preferences / a "Manage preferences" panel.** There is exactly one non-essential purpose (Google Analytics); a multi-toggle panel would be theater. Binary Accept/Decline only. (Revisit only if more trackers are ever added.)
- **A consent-management platform (Cookiebot, OneTrust, etc.).** This is hand-rolled, first-party, no third-party script.
- **Ad/marketing consent signals** (`ad_storage`, `ad_user_data`, `ad_personalization`). GA here is analytics-only; only `analytics_storage` is gated. (Consent Mode defaults for ad signals may be set to denied for completeness but there is no ad tech to enable.)
- **Server-side consent logging / audit database.** Consent is recorded client-side in localStorage; no backend consent ledger in v1.
- **Geo-targeting the banner** (showing it only to EU visitors). v1 shows it to everyone — simpler and strictly safer.
- **Cookieless-analytics migration.** We keep Google Analytics and gate it (Path A), rather than switching analytics vendors.
- **Contact-form (`POST /contact`) changes** — unrelated stub, untouched.
