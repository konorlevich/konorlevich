---
name: Petr Travkin — Personal Site
description: Warm Minimalism for a hiring reader — paper, ink, and one terracotta stamp.
colors:
  terracotta: "#a8482a"
  terracotta-hover: "#8f3d22"
  terracotta-active: "#79331c"
  verdigris: "#3f6b5f"
  ledger-paper: "#faf7f2"
  card-white: "#ffffff"
  well-linen: "#f2ece1"
  warm-ink: "#2a2620"
  ink-subdued: "#6b6357"
  ink-meta: "#6f685c"
  hairline: "#e6ddce"
  hairline-soft: "#f0e9dd"
  wax-wash: "#f5e4dc"
  wax-wash-text: "#7a3319"
  banked-charcoal: "#1c1917"
  charcoal-raised: "#262220"
  charcoal-well: "#302b27"
  parchment: "#ede6db"
  parchment-subdued: "#b3a99c"
  parchment-meta: "#a2978a"
  ember-terracotta: "#e08a63"
  ember-link: "#f0a184"
  ember-active: "#c7623c"
  dark-hairline: "#3a342f"
  dark-hairline-soft: "#2f2a26"
  dark-verdigris: "#7fb3a3"
  dark-wax-wash: "#3a2a22"
  dark-wax-wash-text: "#f0b79f"
typography:
  display:
    fontFamily: "Fraunces, Georgia, 'Times New Roman', serif"
    fontSize: "clamp(2.75rem, 9vw, 3.75rem)"
    fontWeight: 600
    lineHeight: 1.05
    letterSpacing: "-0.02em"
  headline:
    fontFamily: "Fraunces, Georgia, 'Times New Roman', serif"
    fontSize: "1.875rem"
    fontWeight: 600
    lineHeight: 1.3
    letterSpacing: "-0.02em"
  title:
    fontFamily: "Inter, system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif"
    fontSize: "1.5rem"
    fontWeight: 600
    lineHeight: 1.25
    letterSpacing: "normal"
  lead:
    fontFamily: "Inter, system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif"
    fontSize: "1.25rem"
    fontWeight: 400
    lineHeight: 1.75
    letterSpacing: "normal"
  body:
    fontFamily: "Inter, system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif"
    fontSize: "1.0625rem"
    fontWeight: 400
    lineHeight: 1.6
    letterSpacing: "normal"
  label:
    fontFamily: "JetBrains Mono, ui-monospace, 'SF Mono', Menlo, monospace"
    fontSize: "0.75rem"
    fontWeight: 500
    lineHeight: 1.6
    letterSpacing: "0.08em"
rounded:
  sm: "6px"
  md: "10px"
  lg: "16px"
  full: "999px"
spacing:
  "1": "0.125rem"
  "2": "0.25rem"
  "3": "0.375rem"
  "4": "0.5rem"
  "5": "0.75rem"
  "6": "1rem"
  "7": "1.5rem"
  "8": "2rem"
  "9": "3rem"
  "10": "4rem"
  "11": "6rem"
  "12": "8rem"
components:
  button-primary:
    backgroundColor: "{colors.terracotta}"
    textColor: "{colors.ledger-paper}"
    rounded: "{rounded.full}"
    padding: "0.75rem 2rem"
    height: "44px"
  button-primary-hover:
    backgroundColor: "{colors.terracotta-hover}"
  button-primary-active:
    backgroundColor: "{colors.terracotta-active}"
  button-ghost:
    backgroundColor: "transparent"
    textColor: "{colors.warm-ink}"
    rounded: "{rounded.full}"
    padding: "0.75rem 2rem"
    height: "44px"
  button-sm:
    rounded: "{rounded.full}"
    padding: "0.5rem 1rem"
    height: "38px"
  chip:
    backgroundColor: "{colors.well-linen}"
    textColor: "{colors.ink-subdued}"
    rounded: "{rounded.full}"
    padding: "0.375rem 0.75rem"
  tag:
    backgroundColor: "{colors.well-linen}"
    textColor: "{colors.ink-subdued}"
    typography: "{typography.label}"
    rounded: "{rounded.sm}"
    padding: "2px 0.5rem"
  consent-button:
    backgroundColor: "{colors.warm-ink}"
    textColor: "{colors.ledger-paper}"
    rounded: "{rounded.full}"
    padding: "0.5rem 1.5rem"
    height: "44px"
  theme-toggle:
    backgroundColor: "transparent"
    textColor: "{colors.ink-subdued}"
    rounded: "{rounded.full}"
    padding: "0.5rem 1rem"
    height: "38px"
  brand-mark:
    backgroundColor: "{colors.terracotta}"
    textColor: "{colors.ledger-paper}"
    rounded: "{rounded.full}"
    width: "34px"
    height: "34px"
  card-menu:
    backgroundColor: "{colors.card-white}"
    rounded: "{rounded.md}"
    padding: "0.375rem"
---

# Design System: Petr Travkin — Personal Site

## Overview

**Creative North Star: "The Warm Ledger"**

A well-kept professional record, set on good paper. The system behaves like a
ledger kept by someone precise: an editorial serif names the person, a humanist
sans explains, and a mono face records anything factual — dates, skill
categories, file extensions, the 404 code. One terracotta stamp marks the single
action worth taking. Nothing else on the page is allowed to compete for it.

The mood is **warm, confident, unhurried**. Space is spent generously because
there is no anxiety about holding attention: 96px between sections on desktop, a
reading column capped at 42rem inside a 72rem page, and a hero that gives the
name a full line before anything else is asked of the reader. Warmth is
structural rather than decorative — it lives in the paper (`#faf7f2`, never
white), the ink (`#2a2620`, never black), and shadows tinted brown rather than
gray. The result should read as a person, not a corporate CV and not a product
landing page.

Restraint here is a claim, not a style. The visitor is deciding whether this
engineer is credible in about twenty seconds, and the page's own composure is
part of the evidence. Two rejections are binding: **no borrowed SaaS-landing
polish** (gradients on hero text, glass panels, floating blurred blobs, animated
counters, logo clouds), and **light is the identity** — the warm-charcoal dark
theme is a designed alternate that follows the reader's system preference, never
the site's default look.

**Key Characteristics:**
- Warm paper and warm ink; pure black and pure white never carry text.
- Exactly one accent hue, reserved for action, authorship, and emphasis.
- Three type voices with strictly separate jobs: name, explain, record.
- Flat at rest; shadow is a response to state, not ambient decoration.
- Full-round pills for anything interactive; 16px cards for anything containing.
- Every value comes from a token — the shipped CSS hardcodes almost nothing.
- Both themes hold WCAG AA, and the page works fully with JavaScript off.

## Colors

A warm-neutral field with one stamp: paper, ink, and terracotta, plus a single
muted teal reserved for one semantic job.

### Primary
- **Sealing-Wax Terracotta** (`#a8482a`): the site's only accent. It carries the
  email CTA, links, section eyebrows, the brand mark, company names in the
  experience list, the achievement bullets (at 55% opacity), and the focus ring.
  Its hover (`#8f3d22`) and active (`#79331c`) steps deepen rather than lighten —
  the button presses *into* the paper. In dark, the role passes to **Ember
  Terracotta** (`#e08a63`) for surfaces and **Ember Link** (`#f0a184`) for text,
  because the light value fails AA on charcoal.

### Secondary
- **Verdigris** (`#3f6b5f`): a muted teal with exactly one shipped use — the dot
  on the availability chip. It signals "open to work" without recruiting the
  action color. Dark counterpart **`#7fb3a3`**. Adding a second use for this hue
  is a system change, not a styling choice.

### Neutral
- **Warm Ledger Paper** (`#faf7f2`): the page. Never white.
- **Card White** (`#ffffff`): the one place pure white appears — surfaces that
  must lift off the paper (the download menu, the consent bar, the ring around
  the portrait).
- **Well Linen** (`#f2ece1`): recessed fills — chips, skill tags, inline code,
  hovered rows.
- **Warm Ink** (`#2a2620`): body text and headings; also the inverse surface for
  the consent buttons.
- **Ink Subdued** (`#6b6357`): dates, meta lines, intro copy, secondary nav.
- **Ink Meta** (`#6f685c`): the smallest text — skill categories, file
  extensions, the footer note. Tuned to 5.16:1 on paper, not chosen by eye.
- **Hairline** (`#e6ddce`) / **Hairline Soft** (`#f0e9dd`): default borders and
  the quieter dividers between sections.
- **Wax Wash** (`#f5e4dc`) with **Wax Wash Text** (`#7a3319`): the text-selection
  pair, so highlighting stays inside the palette.

### Dark theme
Dark is authored, not inverted. **Banked Charcoal** (`#1c1917`) is warm
near-black; **Charcoal Raised** (`#262220`) and **Charcoal Well** (`#302b27`)
carry the same lift-and-recess relationship as their light counterparts;
**Parchment** (`#ede6db`) is a warm off-white, never `#fff`. The dark values live
once, as inert `--dark-*` tokens, and are mapped into the semantic tokens by two
blocks: an explicit `[data-theme="dark"]` and a `:root:not([data-theme])`
system-preference fallback. One source of truth, two entry points.

A four-color status set (`--color-status-*`) is defined in `tokens.css` but
currently unused. It is reserved, not part of the shipped vocabulary; don't
introduce it as decoration.

### Named Rules

**The One Stamp Rule.** There is exactly one accent hue. Anything that wants
emphasis either takes terracotta or takes none — a second accent color is a
redesign, not a variation. Verdigris is not an exception; it is a single reserved
semantic mark.

**The Never-Pure Rule.** Pure black and pure white never carry text and never
serve as the page background, in either theme. White appears only as a lifted
surface; black appears only inside shadow alpha.

**The Authored-Dark Rule.** The dark palette is designed, not derived. Never
generate a dark value by inverting or `filter`-ing a light one, and never reuse a
light shadow in dark — the dark set is black-based and separately tuned.

## Typography

**Display Font:** Fraunces (with Georgia, Times New Roman, serif)
**Body Font:** Inter (with system-ui, -apple-system, Segoe UI, Roboto)
**Label/Mono Font:** JetBrains Mono (with ui-monospace, SF Mono, Menlo)

All three are self-hosted latin-subset variable WOFF2 files with
`font-display: swap`, served from `/static/fonts`. No third-party font host, ever
— it is both a performance and a privacy commitment.

**Character:** A soft, slightly editorial serif carries the person's name and the
section titles; a neutral humanist sans does all the explaining; a quiet mono
records facts. The mono is an engineer's fingerprint — annotation, not a terminal
costume.

### Hierarchy
- **Display** (Fraunces 600, `clamp(2.75rem, 9vw, 3.75rem)`, line-height 1.05,
  tracking −0.02em): the hero name only. One per page.
- **Headline** (Fraunces 600, 1.875rem → 2.25rem at ≥768px, line-height 1.3):
  section titles, the legal page title, the 404 title. The contact section's
  "Let's talk" runs one step larger (2.25rem) because it is the page's closing
  argument.
- **Title** (Inter 600, 1.5rem, line-height 1.25): job roles in the experience
  list. Deliberately *not* the serif — see the rule below.
- **Lead** (Inter 400, 1.25rem, line-height 1.75): section intros and the contact
  subheading; the one paragraph size that sits above body.
- **Body** (Inter 400, 1.0625rem/17px, line-height 1.6): everything else. 17px
  rather than 16px is a warmth decision — it reads slightly generous. Paragraphs
  are capped at 42rem globally; individual blocks tighten further (hero tagline
  32ch, hero intro 46ch, contact sub 40ch, consent copy 62ch).
- **Label** (JetBrains Mono 500, 0.75rem, tracking 0.08em, uppercase): section
  eyebrows, skill categories, skill tags, file extensions. The hero eyebrow and
  experience dates run one step larger (0.875rem) but the same voice.

### Named Rules

**The Three Voices Rule.** Fraunces names, Inter explains, JetBrains Mono
records. A piece of text belongs to exactly one voice, decided by what it *is*,
not by how much attention it needs.

**The Serif Names, Sans Informs Rule.** The display serif is for the person and
the page's own structure. Job titles, company roles, and any content the reader
is *scanning for information* stay in Inter — `.exp-role` overrides the global
heading font on purpose. Don't "fix" it.

**The Mono-Is-Fact Rule.** Mono is only ever applied to something factual and
short: a date range, a category, a technology name, an extension, a status code.
Never a sentence, never a heading, never for atmosphere.

## Layout

Mobile-first, single column, one centered measure. `.container` is
`max-width: 72rem` with 24px gutters that open to 48px at ≥768px; the reading
measure is `--max-width-content: 42rem` (~68ch) and paragraphs inherit it
globally, so no prose block ever runs the full page width. A wide bound
(`60rem`) exists for future grid sections.

Spacing is an 8px base with generous multipliers. The high steps carry the page's
rhythm: 64px section padding on mobile → 96px at ≥768px, 48px between experience
entries, 128px available for major breaks. Consecutive `.section` elements
collapse their shared gap (`.section + .section { padding-top: 0 }`) so rhythm
stays even rather than doubling.

Breakpoints are documented as a comment block, not tokens (CSS custom properties
can't be used in media queries): 375px mobile · **768px tablet** · **1024px small
desktop** · 1280px · 1536px. Only two carry real structural change:

- **≥768px** — the nav links appear (hidden on mobile, where the header is brand
  + theme toggle + CTA); the hero becomes `row-reverse` with a 200px portrait;
  experience items become a `12rem 1fr` two-column grid with dates in the left
  rail; the consent bar goes from stacked to one row.
- **≥1024px** — the hero name locks to its full 3.75rem and the portrait grows to
  240px.
- **≤479px** — the facts row becomes a 2-up grid with the long availability chip
  spanning both columns, and the theme toggle sheds its label to keep the header
  on one line.

Horizontal overflow is defended structurally: `overflow-x: clip` on both `html`
and `body`, plus `scroll-margin-top: 6rem` on `:target` so anchor jumps clear the
sticky header.

### Named Rules

**The Two Measures Rule.** Content lives at 42rem; the page lives at 72rem. The
gap between them is deliberate air, not wasted space — never widen prose to fill
the container.

**The Reverse Hero Rule.** From 768px up the hero reads text-left,
portrait-right (`flex-direction: row-reverse` on a source order that puts the
portrait first). The reading eye must land on the name, and the DOM must give the
portrait to a screen reader after it. Keep both facts true.

## Elevation & Depth

**Flat by default, lift on state.** Surfaces sit directly on the paper and are
separated by hairline borders and tonal shifts (paper → card → well), not by
resting shadows. Shadow is a *response*: to hover, to focus, to something
arriving, or to an object that is genuinely floating above the page.

The four shipped exceptions are all justified: the portrait carries a soft
shadow and a 3px card-white ring because it is a physical object on the page; the
download menu and the consent bar are overlays; and every focus ring is a shadow
by construction.

### Shadow Vocabulary
- **`--shadow-sm`** (`0 1px 2px rgba(60, 45, 30, 0.06)`): the primary button at
  rest — barely there, just enough to seat it.
- **`--shadow-md`** (`0 4px 12px rgba(60, 45, 30, 0.08)`): hover lift on the
  primary button and consent buttons; the portrait; the skip link when focused.
- **`--shadow-lg`** (`0 12px 32px rgba(60, 45, 30, 0.12)`): true overlays only —
  the download menu card and the consent bar.
- **`--shadow-focus`** (`0 0 0 3px rgba(168, 72, 42, 0.35)`): the global focus
  ring. `:focus-visible` sets `outline: none` and replaces it with this ring plus
  a 6px radius, so focus is always terracotta and always visible on every
  interactive element.

Dark theme swaps all four for black-based values at higher alpha
(`0.40` → `0.55`), never the light values.

### Named Rules

**The Flat-By-Default Rule.** A new surface starts with zero shadow. Add one only
to answer a state change or to float a genuine overlay — and then take it from
the vocabulary above rather than inventing a value.

**The Warm Shadow Rule.** Light-theme shadows are tinted `rgba(60, 45, 30, …)`,
never neutral gray or black. A gray shadow on warm paper reads as dirt.

## Shapes

Two shapes carry the whole system. **Interactive things are pills**
(`--border-radius-full`, 999px): buttons at every size, chips, tags on the
availability row, the brand mark, the portrait, the theme toggle, the consent
buttons. **Containing things are cards** (`--border-radius-lg`, 16px). Between
them sit two utility radii: 10px (`md`) for the download menu's floating card,
and 6px (`sm`) for small inline fills — skill tags, inline `<code>`, and the
implicit radius the focus ring draws around any element it lands on.

Borders are hairlines, always 1px, always a palette hairline color. The system
has no double borders, no dashed strokes, no decorative rules, and no clipping
shapes or angled geometry. Depth and separation come from tone and space first,
a hairline second, a shadow last.

### Named Rules

**The Pill-or-Card Rule.** If a visitor can act on it, it is a pill. If it holds
content, it is a 16px card. Anything proposing a third silhouette needs a reason
that survives being said out loud.

## Components

The component philosophy is **confident, never loud**: elements state themselves
clearly and then stop. Generous padding, one solid accent, no borders competing
for attention, no second color asking to be noticed.

### Buttons
- **Shape:** full pill (`999px`), min-height 44px — a real touch target, not a
  coincidence.
- **Primary:** solid terracotta with paper-colored text, 12px/32px padding,
  `--shadow-sm` at rest. Hover deepens the fill and lifts to `--shadow-md`;
  active deepens again and presses down `translateY(1px)`. Labels never wrap
  (`white-space: nowrap`).
- **Ghost:** transparent with a hairline border and ink text. Hover darkens the
  border to full ink and fills with Well Linen. Used for every secondary action —
  "See my experience", "Download CV", the 404's return path.
- **Small (`.btn-sm`):** 38px, 8px/16px padding, label at 0.875rem. The header CTA
  only.
- **Focus:** never a browser outline — the terracotta `--shadow-focus` ring, from
  the global `:focus-visible` rule.

### Chips and Tags
- **Chip:** pill, Well Linen fill, hairline-soft border, 0.875rem medium label,
  with a 7px dot in Ink Meta. The availability chip swaps its dot to Verdigris —
  the only place that hue appears.
- **Tag:** the mono voice at 0.75rem on Well Linen with a 6px radius, no border.
  Skill and technology names only; deliberately quieter than chips because they
  are supporting evidence, not headline facts.

### Cards / Containers
- **Corner style:** 16px (`--border-radius-lg`) via `--card-radius`.
- **Background:** Card White lifting off the paper; Charcoal Raised in dark.
- **Shadow strategy:** none at rest — see Elevation. The two shipped floating
  surfaces (download menu, consent bar) take `--shadow-lg`.
- **Border:** a single hairline.

### Navigation
The topbar is sticky with a translucent paper background
(`color-mix(in srgb, var(--color-bg-primary) 88%, transparent)`) plus a
saturating backdrop blur, closed by a hairline-soft bottom border, 64px min
height. The brand is the serif name beside a 34px terracotta initials disc. Nav
links are 0.875rem medium in Ink Subdued, darkening to full ink on hover, and are
hidden below 768px — on mobile the header carries brand, theme toggle, and the
email CTA only. There is no hamburger menu and no mobile drawer; the page is
short enough that scrolling is the navigation.

### Theme toggle
A real labeled button, not an icon-only control: the written mode ("Auto" /
"Light" / "Dark") is the state cue, and a half-filled 12px disc mirrors it —
never color alone. It ships `hidden` and is revealed by `theme.js`, because
without JavaScript the choice cannot be persisted and `prefers-color-scheme`
already handles that case correctly. Below 480px it keeps the disc and drops the
label.

### Download menu (signature component)
A native `<details>` disclosure with zero JavaScript: the `<summary>` is styled
as a ghost button with a caret that rotates 180° when open, and the panel is a
Card White surface at 10px radius with `--shadow-lg`, centered under the trigger.
Each row pairs a full-contrast label with a mono extension in Ink Meta; hover and
`:focus-visible` share one treatment (Well Linen fill, terracotta text). This is
the house pattern for any future menu — reach for `<details>` before reaching for
a script.

### Consent bar (signature component)
A non-modal `<aside>` pinned to the bottom, Card White over a hairline top
border with `--shadow-lg`, respecting `env(safe-area-inset-bottom)`. Copy stacks
above the buttons on mobile and sits left-of-them at ≥768px. It rises in with a
400ms `translateY(100%)` → `0` animation that is switched off entirely under
`prefers-reduced-motion`.

**The Equal Choice Rule.** Accept and Decline share one identical treatment —
solid Warm Ink, not the terracotta CTA — so the site's action color can never
steer a consent decision. Both are 44px, both are equal width on mobile. Any
change that makes one of them easier than the other breaks this rule.

### Motion
A small vocabulary, all of it 150ms `cubic-bezier(0.4, 0, 0.2, 1)` unless noted:
button press `translateY(1px)`, arrow links nudging `translateX(3px)` on hover,
color and border transitions on the toggle and nav, the caret's rotate, and the
consent bar's 400ms ease-out rise. Every duration token collapses to `0ms` under
`prefers-reduced-motion`, so honoring reduced motion is automatic for anything
that uses the tokens — and a hard failure for anything that hardcodes a duration.

## Do's and Don'ts

### Do:
- **Do** take every value from a token. `tokens.css` is the single source of
  truth; component CSS references it and hardcodes nothing.
- **Do** keep terracotta for action, authorship, and emphasis — the CTA, links,
  eyebrows, company names, the brand mark, the focus ring.
- **Do** hold WCAG AA in *both* themes and check the small text specifically
  (Ink Meta and Parchment Meta were tuned to 5.16:1 and 6.0:1; they have no
  headroom to spare).
- **Do** give every interactive element a 44px touch target (38px only for the
  header's small CTA and the theme toggle, both pointer-adjacent).
- **Do** prefer a native element over a script — `<details>` for the download
  menu, `prefers-color-scheme` for the default theme, `<aside>` for the consent
  bar. The page must stay fully usable with JavaScript off.
- **Do** cap prose at 42rem and tighten further for hero and contact copy.
- **Do** define any new dark value once, as a `--dark-*` token, and map it in
  both the `[data-theme="dark"]` and `:root:not([data-theme])` blocks.

### Don't:
- **Don't** import SaaS-landing polish: no gradient text, no glassmorphism
  panels, no floating blurred blobs, no animated counters, no logo clouds, no
  "trusted by" bars. *(Confirmed binding.)*
- **Don't** make dark the default or the identity. Light warm paper is the site;
  dark follows the reader's system preference. *(Confirmed binding.)*
- **Don't** introduce a second accent hue, or give Verdigris a second job.
- **Don't** use `--easing-bounce` (`cubic-bezier(0.34, 1.56, 0.64, 1)`). It is
  defined in `tokens.css` but nothing in the shipped system uses it; overshoot
  reads as tacky against this restraint. Real objects decelerate — stay with
  `--easing-default` and `--easing-out`.
- **Don't** put a resting shadow on a new surface, or use a neutral-gray shadow
  in the light theme.
- **Don't** set body copy, headings, or atmosphere in the mono face, and don't
  let mono accumulate into a terminal motif.
- **Don't** replace the global `:focus-visible` ring with a browser outline or a
  per-component focus style.
- **Don't** hardcode a transition duration; use the duration tokens so
  `prefers-reduced-motion` keeps working.
- **Don't** remove or restyle the "Website created by Soarline Studio" footer
  credit — it is a binding commitment recorded in PRODUCT.md, and its JSON-LD
  `creator` block must always agree with the visible text.
