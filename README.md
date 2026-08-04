# konorlevich

Personal website for Petr Travkin — a single self-contained Go binary that
renders every page from `cv.yaml` at boot and forwards inbound email.

## Run

```sh
go run .           # serves on :8080 (override with PORT or CONFIG_FILE)
```

## How it works

Everything the service serves is compiled into the binary with `//go:embed`:
templates, static assets and `cv.yaml`. Nothing is read from disk at runtime, so
the binary runs from any working directory.

At startup the service:

1. parses and validates `cv.yaml` (missing required fields are a fatal error),
2. minifies every CSS/JS asset and content-hashes it for cache busting,
3. renders every page, the PDF, the Markdown CV and all ops files **to bytes**,
4. precompresses each of them with Brotli and gzip and computes an `ETag`.

A request therefore never executes a template or touches the disk — it picks a
prepared buffer, negotiates an encoding, and writes it (or returns `304`).
A broken template or a missing asset fails the boot, never a user's request.

Content lives in `cv.yaml`; templates in `templates/`; assets in `static/`.

Routes:

| Method & path | Purpose |
| ------------- | ------- |
| `GET /` | The webpage |
| `GET /cv` | `301` → `/` (legacy alias; one indexable URL for the content) |
| `GET /cv/download` | CV as a generated PDF |
| `GET /cv/download.md` | CV as Markdown |
| `GET /privacy` | Privacy & cookies page (`noindex,follow`) |
| `GET /healthz` | Liveness probe for the platform health check |
| `GET /robots.txt` | References the sitemap and llms.txt |
| `GET /sitemap.xml` | Indexable pages only |
| `GET /llms.txt` | Markdown mirror of the page facts, for AI crawlers |
| `GET /site.webmanifest` | PWA manifest |
| `GET /favicon.ico` | Multi-size ICO (16/32/48) at the root |
| `GET /static/…` | CSS, fonts, images, icons |
| `POST /webhooks/resend/inbound` | Resend inbound-email webhook (see below) |

Any other path returns a real `404` with the designed 404 page.

### Site configuration

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `BASE_URL` | `https://konorlevich.tech` | Canonical origin for canonical/OG/sitemap URLs. |
| `GA_ID` | _unset_ | Google Analytics measurement id. **Unset means no analytics, no consent banner and no JavaScript at all.** |
| `GTM_ID` | _unset_ | Google Tag Manager container id. Unset emits no GTM snippet. |
| `PORT` | `8080` | Listen port (set by Railway). |
| `CONFIG_FILE` | _unset_ | Load config from this file instead of the embedded default. |
| `LOG_LEVEL` | `info` | logrus level. |

Analytics is consent-gated: Consent Mode v2 is set to **denied** by an inline
`<head>` script before `gtag/js` loads, so GA runs cookieless until the visitor
presses Accept. The choice is stored in `localStorage` for ~6 months.

## Inbound email forwarding

Email sent to `*@konorlevich.tech` is received by [Resend](https://resend.com),
which POSTs an `email.received` webhook to `POST /webhooks/resend/inbound`. The
service verifies the signature, fetches the full message from Resend's
Received-emails API, and re-sends it to `FORWARD_TO` with **Reply-To set to the
original sender** (so replies go straight back to whoever wrote in).

The endpoint is only registered when `RESEND_API_KEY`, `RESEND_FROM` and
`FORWARD_TO` are all set; otherwise it logs that forwarding is disabled.

### Environment variables

| Variable | Required | Description |
| -------- | -------- | ----------- |
| `RESEND_API_KEY` | yes | Resend API key (used to fetch the message and send the forward). |
| `RESEND_FROM` | yes | Verified sender for the forward, e.g. `konorlevich.tech <forward@konorlevich.tech>`. Must be on a domain verified in Resend. |
| `FORWARD_TO` | yes | Destination inbox, e.g. `konorlevich@gmail.com`. |
| `RESEND_WEBHOOK_SECRET` | recommended | Svix signing secret (`whsec_…`) from the webhook's dashboard page. When set, signatures are enforced; when unset, verification is skipped (a warning is logged). |
| `FORWARD_DOMAIN` | no | Only forward mail actually addressed to `*@<domain>` (matched against `received_for`, falling back to `To`). Off-domain mail is acked and skipped. When unset, all received mail is forwarded. |
| `FORWARD_SUBJECT_PREFIX` | no | Subject prefix for forwarded mail (default `[konorlevich.tech] `). |

### Resend setup

1. **Verify the domain** `konorlevich.tech` in Resend and add its DNS records.
2. **Enable inbound** for the domain (adds the MX record that routes
   `*@konorlevich.tech` to Resend).
3. **Create a webhook** with the `email.received` event, pointing at
   `https://<your-host>/webhooks/resend/inbound`. Copy its signing secret into
   `RESEND_WEBHOOK_SECRET`.
4. Set `RESEND_API_KEY`, `RESEND_FROM`, `FORWARD_TO` in the deployment env.

### Notes / limits (v1)

- **Attachments are not re-attached.** If the original had any, the forward
  includes a note plus a link to download the original message (Resend's signed
  raw-email URL, which expires).
- Forwarding is synchronous: on a transient Resend failure the endpoint returns
  `502` so Resend/Svix retries delivery.

## Test

```sh
go test ./...
```
