# Information Architecture: Checklist Compliance Pass

_Feature slug: `checklist-compliance`_ · Written 2026-08-21 ·
Reads from [`DESIGN_BRIEF.md`](DESIGN_BRIEF.md)

> **Scope note, stated plainly.** This pass adds no pages, no routes, no navigation
> items and no URLs. The site map and navigation model below are therefore a
> *record of what exists*, unchanged, and exist here to make one thing explicit: what
> must still be true after the pass. The sections that carry real decisions are
> **Render-Pipeline Attachment Points**, **Content Date Model**, **Event Wiring Map**
> and **Naming Conventions** — that is where this document earns its place.

---

## Site Map

Unchanged by this pass. `→` marks a non-page response.

- **Home** `/` — the whole CV. The only indexable page, and the only sitemap entry.
  - `/cv` → `301` to `/` (legacy alias; keeps one indexable URL for the content)
  - `/cv/download` → PDF attachment, generated at boot
  - `/cv/download.md` → Markdown attachment, generated at boot
- **Privacy & cookies** `/privacy` — `noindex,follow`, excluded from the sitemap
- **404** — served for every unmatched path, real `404` status, `noindex,follow`,
  never routed by URL (internal key `/__404`)

Ops surfaces (all present, all unchanged): `/healthz`, `/robots.txt`, `/sitemap.xml`,
`/llms.txt`, `/site.webmanifest`, `/favicon.ico`, `/static/…`,
`POST /webhooks/resend/inbound`.

**Depth: one level.** Nothing in this pass adds a second.

## Navigation Model

- **Primary navigation**: 2 items, unchanged. `Experience` (an in-page anchor on home,
  swapping to `Home` off-home) and `CV`. No hamburger at any width — correct per the
  checklist's "no hamburger where a menu isn't needed".
- **Secondary navigation**: none. There are no sections deep enough to need it.
- **Utility navigation**: theme toggle (3-mode) + the `Email me` accent CTA in the
  topbar. Footer carries social links, `Privacy`, and `Cookie settings`.
- **Mobile navigation**: identical structure, reflowed. No behavior change.

**What this pass changes in the chrome:** nothing structural. Footer and social links
get routed through a shared `extlink` partial, and several existing controls gain
`data-track` attributes. No item is added, removed, renamed or reordered.

## Render-Pipeline Attachment Points

This is the part of the IA that actually matters here, because the pipeline is unusual:
**every page is rendered to bytes once at boot and served from a map.** No request ever
executes a template. Anything added must survive that, and two of the eight fixes
interact with it directly.

```
boot
 ├─ config.Load / Parse ......................... config.yaml (embedded)
 ├─ cv.Parse(cv.yaml) ........................... ← NEW: pages[].updated validated here
 ├─ site.FromEnv() .............................. ← NEW: resolves site.Tag{Kind,ID}, fatal on bad prefix
 ├─ web.LoadAssets(static) ...................... minify + content-hash + precompress
 ├─ Renderer.RenderAll(templates) ............... one template set per page
 │    └─ layout → head → consent → chrome → page
 │         ├─ head.html ......................... theme pre-paint (inline, sync)
 │         ├─ consent.html ...................... ← MODIFIED: one tag path from Kind; emits window.__tag
 │         └─ chrome.html ....................... ← MODIFIED: extlink partial + data-track attrs
 ├─ buildOps(...) ............................... ← MODIFIED: sitemap lastmod from content dates
 └─ render.PDF / render.Markdown ................ unchanged

request  →  map[path]*Blob  →  encoding negotiation  →  304 or bytes
```

**Three consequences the build must respect:**

1. **`window.__tag` is baked into the HTML at boot.** It is not a runtime lookup. The
   tag id therefore cannot change without a redeploy — which is correct and desirable,
   but means setting the Railway variable requires a restart to take effect.

2. **The 404 body is one shared blob.** It is rendered once and served for every missed
   path, so it *cannot* know which URL was requested. `page_not_found` reads
   `location.pathname` client-side. This is the single non-declarative event, and the
   reason is structural, not stylistic.

3. **`NeedsConsent()` currently gates whether consent markup exists at all.** With the
   merge to `analytics.js`, the gate on the *script tag* in `layout.html` goes away
   (the file loads unconditionally and no-ops), but the gate on the *cookie bar markup*
   stays — no tag means no cookies, so there is genuinely nothing to ask about.

## Content Date Model

Per-page dates, as decided in Phase 3 — accepting that two of the three values are not
consumed today, in exchange for every page already carrying a date the moment it becomes
indexable.

**Location:** a `pages:` map in `cv.yaml`, keyed by the page's request path. Content
stays data, never code, consistent with the rest of the content layer.

```yaml
pages:
  "/":        { updated: 2026-08-21 }
  "/privacy": { updated: 2026-07-05 }
  "/__404":   { updated: 2026-07-05 }
```

**Rules:**

- `updated` is the date that page's **content** actually changed. Touching CSS, a
  footer link, a template or a dependency is **not** a content change and must not bump
  it.
- Validated at boot: every `pageSpec` must resolve to a non-zero date, or the process
  does not start. Missing dates fail the deploy, not a crawler's expectations.
- `pageSpec` gains an `updated` field populated from this map; `sitemapXML` reads it
  instead of `buildTime`. `buildTime` keeps its other jobs (`Last-Modified`, cache
  validators) — those *are* legitimately deploy-time.
- Format: `YYYY-MM-DD` (W3C datetime), emitted verbatim.
- **Verification (checklist-mandated):** two consecutive boots with no content edit must
  produce a **byte-identical** sitemap. This becomes a `go test`.

**Growth path:** when a page count grows past a handful, or a section hub appears, this
map moves to its own `content/pages.yaml` and a hub's `lastmod` becomes the max over its
children. Not needed at three pages.

## Event Wiring Map

Declarative and centralized — one `document`-level delegated listener reads attributes
off the element actually used. No inline `onclick`, no per-component listeners.

| Event | Attached to | Mechanism |
| --- | --- | --- |
| `file_download` | `/cv/download`, `/cv/download.md` links (topbar `CV`, hero, 404) | `data-track` on the `<a>` |
| `contact_click` | every `mailto:` link (topbar, hero, footer) | `data-track` on the `<a>` |
| `select_content` | primary CTAs | `data-track` on the `<a>`/`<button>` |
| `theme_toggle` | theme control | called from `theme.js` on change |
| `consent_update` | Accept / Decline / Esc | called from the consent controller |
| `page_not_found` | the 404 page | fired on load from `location.pathname` |

`location` params use a fixed vocabulary: `header` \| `hero` \| `footer`. Any new
surface picks one of these or the list grows deliberately — never ad-hoc.

## Content Hierarchy

Unchanged on every page. Recorded only as the regression baseline the review phase
checks against.

### Home `/`
1. **Hero** — name, tagline, intro, portrait, primary actions. The identity claim.
2. **Experience** — the substance, and the primary nav's only anchor target.
3. **Skills** — scannable support for the above.
4. **Contact** — the conversion, repeated from the hero.
5. **Footer** — colophon, elsewhere links, creator credit.

### Privacy `/privacy`
1. **What is collected and the reader's choice** — the `<h1>` answers the question
   directly, before any legal framing.
2. Cookie specifics, lifetimes, legal basis.
3. Withdrawal route (`Cookie settings` in the footer).

### 404
1. Plain reassurance that nothing is broken on the visitor's end.
2. Two routes back into real content.
3. **Footer — which after this pass sits at the bottom of the viewport, not mid-screen.**

## User Flows

### Visitor downloads the CV (the site's only conversion)
1. Visitor lands on `/`.
2. Uses `CV` in the topbar, or a download action in the hero.
3. `data-track="file_download"` fires through the delegated listener.
   - If a tag is configured → one `gtag('event','file_download',…)` hit.
   - If not → `track()` returns silently; the download is unaffected.
4. Server sets `Content-Disposition: attachment` and writes the prepared blob.
5. **The download never depends on the event.** Analytics failing must never cost a
   download — the `<a href>` does the work, JS only observes.

### Visitor meets the consent bar
1. First visit with a tag configured → bar reveals, without stealing focus.
2. Consent Mode is already `denied` by default from the synchronous head script; GA runs
   cookieless in the meantime.
3. Visitor acts:
   - **Accept** → `granted` stored (versioned + timestamped) → `consent_update`
   - **Decline** → `denied` stored → `consent_update`
   - **Esc** → **`denied` stored** (changed in this pass; previously dismissed
     without recording, which re-prompted keyboard users on every visit forever)
4. Bar retires. `Cookie settings` in the footer reopens it — withdrawal as easy as
   granting.

### Visitor follows an off-domain link
1. Sees the link, with an external-link icon beside it.
2. Screen-reader users hear the accessible name including "(opens in new tab)".
3. Opens in a new tab with `rel="noopener noreferrer"`.
4. **Internal links are untouched** — same tab, always. `mailto:` gets no `target`.

### Crawler requests the sitemap
1. `GET /sitemap.xml` → prepared blob.
2. One `<url>`: `/`, with `lastmod` from `pages["/"].updated`.
3. A redeploy with no content edit returns **byte-identical** XML.

## Naming Conventions

The pick-one rule. One canonical label per concept, verbatim everywhere.

| Concept | Label in UI | Notes |
| --- | --- | --- |
| The CV file | **CV** | Nav label and download control. Never "Resume" or "Résumé" — one word, sitewide. |
| Start an email | **Email me** | The accent CTA. Not "Contact", not "Get in touch". |
| Reopen consent | **Cookie settings** | Footer. Matches the privacy page's wording for withdrawal. |
| Accept analytics | **Accept** | Equal weight and size with Decline. Never "Accept all" — there is one tracker, so "all" is theatre. |
| Refuse analytics | **Decline** | Never "Reject", never "Manage preferences". |
| Theme states | **Dark / Light / Auto** | Written words are the non-color state cue. |
| Leaves the site | **(opens in new tab)** | Visually-hidden, verbatim. |
| Site credit | **Website created by Soarline Studio** | Verbatim in footer and `creditText`. |

**In code**, the same discipline: event names are Go constants matching the JS strings
exactly, defined once. Two spellings of `file_download` is two half-populated reports.

## Component Reuse Map

| Component | Used on | Behavior differences |
| --- | --- | --- |
| `layout.html` | all 3 pages | one template set parsed per page, so `"content"` cannot collide |
| `head.html` | all 3 | per-page title/description/robots/canonical; theme pre-paint identical |
| `chrome.html` → `header` | all 3 | nav swaps `Experience` anchor ↔ `Home` via `.IsHome` |
| `chrome.html` → `footer` | all 3 | identical; **after this pass, pinned to viewport bottom on short pages** |
| `consent.html` | all 3 | emits exactly one tag path from `site.Tag.Kind`; nothing when unset |
| `cookiebar` | all 3 | markup present only when a tag is configured |
| **`extlink`** (new) | footer, social links, creator credit | one definition; every off-domain link routes through it |
| **`.visually-hidden`** (new) | wherever `extlink` renders | clip-based, never `display:none` |

## Content Growth Plan

Nothing in this pass grows. Recorded thresholds for revisiting:

- **Indexable pages > ~5** → move `pages:` to its own content file; add section hubs
  with `lastmod` as max-over-children.
- **A second locale becomes real** → §5 lands as a set (locale prefix, cookie, hreflang
  + `x-default`, completeness gate, native review). Explicitly out of scope now; see
  the brief.
- **A contact form appears** → `form_start` / `generate_lead` / `form_error` join the
  catalog, and server-side Measurement Protocol stops being deferrable, because a form
  submit is a conversion the browser cannot be trusted to report.
- **Projects return** → `.design/project-pages/` already holds a brief, IA and task
  list from the version removed in `e9de2d7`. Not this pass.

No pagination, search or archive at this scale.

## URL Strategy

Unchanged. Recorded because the checklist treats URL rules as structural.

- **Pattern**: flat, one level. `/`, `/privacy`, `/cv/download[.md]`.
- **Dynamic segments**: none. No route takes a parameter.
- **Query parameters**: exactly one is read — `?theme=dark|light|auto`, an optional
  deep link handled by the pre-paint script. It is never required, never generated in a
  link, and never affects the canonical URL.
- **Trailing slashes**: `GET /{$}` anchors the root exactly; the catch-all returns a
  real 404 rather than swallowing unknown paths as 200.
- **Canonicals**: self-referential and absolute, built from `BASE_URL` (defaulting to
  `https://konorlevich.tech`). The 404 canonicalizes to `/`.
- **Legacy**: `/cv` → `301` `/`. One indexable URL per piece of content.
