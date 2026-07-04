# konorlevich

Personal website for Petr Travkin — a single Go service that renders the page
from `cv.yaml` (via `html/template`) and forwards inbound email.

## Run

```sh
go run .           # serves on the address in config.yaml (default :8080)
```

Content lives in `cv.yaml`; static assets (CSS, self-hosted fonts, photo) in
`static/`. The page template is `cv_template.html`.

Routes:

| Method & path | Purpose |
| ------------- | ------- |
| `GET /`, `GET /cv` | The webpage |
| `GET /cv/download` | CV as a generated PDF |
| `GET /static/…` | CSS, fonts, images |
| `POST /webhooks/resend/inbound` | Resend inbound-email webhook (see below) |

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
