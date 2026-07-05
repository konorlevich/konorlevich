// Package mailforward turns Resend inbound-email webhooks into forwarded mail.
//
// Flow: Resend receives a message at *@<your-domain> and POSTs an
// "email.received" webhook (metadata only). We verify the Svix signature,
// fetch the full message from Resend's Received-emails API, and re-send it to a
// fixed destination inbox with Reply-To set to the original sender.
package mailforward

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	resend "github.com/resend/resend-go/v2"
	"github.com/sirupsen/logrus"
	svix "github.com/svix/svix-webhooks/go"
)

const (
	maxWebhookBody    = 5 << 20 // 5 MiB — inbound webhook payload cap
	requestTimeout    = 12 * time.Second
	defaultSubjectTag = "[konorlevich.tech] "
)

// Config is read from the environment. From/To/APIKey are required to enable
// forwarding; WebhookSecret is strongly recommended (signatures are only
// verified when it is set).
type Config struct {
	APIKey        string // RESEND_API_KEY
	WebhookSecret string // RESEND_WEBHOOK_SECRET (whsec_...)
	From          string // RESEND_FROM — a verified sender, e.g. "konorlevich.tech <forward@konorlevich.tech>"
	To            string // FORWARD_TO — destination inbox
	Domain        string // FORWARD_DOMAIN — only forward mail addressed to *@<domain> (optional)
	SubjectPrefix string // FORWARD_SUBJECT_PREFIX (optional)
}

// ConfigFromEnv builds a Config from environment variables.
func ConfigFromEnv() Config {
	prefix, ok := os.LookupEnv("FORWARD_SUBJECT_PREFIX")
	if !ok {
		prefix = defaultSubjectTag
	}
	return Config{
		APIKey:        os.Getenv("RESEND_API_KEY"),
		WebhookSecret: os.Getenv("RESEND_WEBHOOK_SECRET"),
		From:          os.Getenv("RESEND_FROM"),
		To:            os.Getenv("FORWARD_TO"),
		Domain:        strings.ToLower(strings.TrimSpace(os.Getenv("FORWARD_DOMAIN"))),
		SubjectPrefix: prefix,
	}
}

// Enabled reports whether the minimum config to forward mail is present.
func (c Config) Enabled() bool {
	return c.APIKey != "" && c.From != "" && c.To != ""
}

// Forwarder handles inbound webhooks.
type Forwarder struct {
	cfg    Config
	log    logrus.FieldLogger
	client *resend.Client
	wh     *svix.Webhook // nil when no secret is configured
}

// New validates config and constructs a Forwarder.
func New(cfg Config, log logrus.FieldLogger) (*Forwarder, error) {
	if !cfg.Enabled() {
		return nil, fmt.Errorf("mailforward: RESEND_API_KEY, RESEND_FROM and FORWARD_TO are required")
	}
	f := &Forwarder{
		cfg:    cfg,
		log:    log,
		client: resend.NewCustomClient(&http.Client{Timeout: 15 * time.Second}, cfg.APIKey),
	}
	if cfg.WebhookSecret != "" {
		wh, err := svix.NewWebhook(cfg.WebhookSecret)
		if err != nil {
			return nil, fmt.Errorf("mailforward: invalid RESEND_WEBHOOK_SECRET: %w", err)
		}
		f.wh = wh
	} else {
		log.Warn("mailforward: RESEND_WEBHOOK_SECRET not set — webhook signatures will NOT be verified")
	}
	return f, nil
}

// --- wire types -------------------------------------------------------------

type webhookEvent struct {
	Type string `json:"type"`
	Data struct {
		EmailID string `json:"email_id"`
		Subject string `json:"subject"`
	} `json:"data"`
}

type receivedEmail struct {
	ID          string       `json:"id"`
	From        string       `json:"from"`
	To          []string     `json:"to"`
	Cc          []string     `json:"cc"`
	ReplyTo     []string     `json:"reply_to"`
	ReceivedFor []string     `json:"received_for"`
	Subject     string       `json:"subject"`
	HTML        string       `json:"html"`
	Text        *string      `json:"text"`
	MessageID   string       `json:"message_id"`
	Attachments []attachment `json:"attachments"`
	Raw         *rawDownload `json:"raw"`
}

type attachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

type rawDownload struct {
	DownloadURL string `json:"download_url"`
	ExpiresAt   string `json:"expires_at"`
}

// --- handler ----------------------------------------------------------------

// Handler returns the HTTP handler for the inbound webhook endpoint.
func (f *Forwarder) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, maxWebhookBody))
		if err != nil {
			http.Error(w, "cannot read body", http.StatusBadRequest)
			return
		}

		// Verify signature against the RAW body (before any parsing).
		if f.wh != nil {
			if err := f.wh.Verify(body, r.Header); err != nil {
				f.log.WithError(err).Warn("mailforward: signature verification failed")
				http.Error(w, "invalid signature", http.StatusUnauthorized)
				return
			}
		}

		var ev webhookEvent
		if err := json.Unmarshal(body, &ev); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		// Ack and ignore any event that isn't an inbound message.
		if ev.Type != "email.received" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if ev.Data.EmailID == "" {
			http.Error(w, "missing email_id", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
		defer cancel()

		email, err := f.getReceivedEmail(ctx, ev.Data.EmailID)
		if err != nil {
			// 502 so Resend/Svix retries the delivery.
			f.log.WithError(err).WithField("email_id", ev.Data.EmailID).Error("mailforward: fetch received email")
			http.Error(w, "fetch failed", http.StatusBadGateway)
			return
		}

		// Only forward mail actually addressed to the configured domain. Ack
		// with 200 (not an error) so Resend does not retry a deliberate skip.
		if !f.addressedToDomain(email) {
			f.log.WithFields(logrus.Fields{
				"domain":       f.cfg.Domain,
				"received_for": email.ReceivedFor,
				"to":           email.To,
			}).Info("mailforward: skipping email not addressed to configured domain")
			w.WriteHeader(http.StatusOK)
			return
		}

		if err := f.sendEmail(ctx, f.buildForward(email)); err != nil {
			f.log.WithError(err).Error("mailforward: forward send")
			http.Error(w, "forward failed", http.StatusBadGateway)
			return
		}

		f.log.WithFields(logrus.Fields{
			"from":    email.From,
			"subject": email.Subject,
			"to":      f.cfg.To,
		}).Info("mailforward: forwarded inbound email")
		w.WriteHeader(http.StatusOK)
	}
}

// addressedToDomain reports whether the email was delivered to a recipient at
// the configured domain. When no domain is configured it forwards everything.
// It prefers received_for (the actual delivery target) and falls back to To.
func (f *Forwarder) addressedToDomain(e receivedEmail) bool {
	if f.cfg.Domain == "" {
		return true
	}
	recipients := e.ReceivedFor
	if len(recipients) == 0 {
		recipients = e.To
	}
	for _, r := range recipients {
		if domainOf(r) == f.cfg.Domain {
			return true
		}
	}
	return false
}

// domainOf extracts the lowercased domain from an address, tolerating a
// display-name form like "Name <user@example.com>".
func domainOf(addr string) string {
	addr = strings.TrimSpace(addr)
	if i := strings.LastIndex(addr, "<"); i >= 0 {
		if j := strings.Index(addr[i:], ">"); j >= 0 {
			addr = addr[i+1 : i+j]
		}
	}
	at := strings.LastIndex(addr, "@")
	if at < 0 {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(addr[at+1:]))
}

// buildForward turns a received email into the outbound forward request.
// Exported behaviour is unit-tested; keep it pure (no I/O).
func (f *Forwarder) buildForward(e receivedEmail) *resend.SendEmailRequest {
	banner := f.banner(e)

	htmlBody := banner.html
	if e.HTML != "" {
		htmlBody += e.HTML
	} else if e.Text != nil {
		htmlBody += "<pre style=\"white-space:pre-wrap;font-family:inherit\">" + html.EscapeString(*e.Text) + "</pre>"
	}

	textBody := banner.text
	if e.Text != nil {
		textBody += *e.Text
	} else if e.HTML != "" {
		textBody += "(HTML-only message — see the HTML version.)"
	}

	req := &resend.SendEmailRequest{
		From:    f.cfg.From,
		To:      []string{f.cfg.To},
		Subject: f.cfg.SubjectPrefix + e.Subject,
		Html:    htmlBody,
		Text:    textBody,
	}
	if e.From != "" {
		req.ReplyTo = e.From // reply goes straight to the original sender
	}
	if e.MessageID != "" {
		req.Headers = map[string]string{"X-Original-Message-Id": e.MessageID}
	}
	return req
}

type banner struct {
	html string
	text string
}

func (f *Forwarder) banner(e receivedEmail) banner {
	recFor := strings.Join(e.ReceivedFor, ", ")
	if recFor == "" {
		recFor = strings.Join(e.To, ", ")
	}

	var attachNote string
	if n := len(e.Attachments); n > 0 {
		plural := "attachment"
		if n > 1 {
			plural = "attachments"
		}
		attachNote = fmt.Sprintf("%d %s not included in this forward.", n, plural)
		if e.Raw != nil && e.Raw.DownloadURL != "" {
			attachNote += " Download the original message: " + e.Raw.DownloadURL
		}
	}

	// HTML banner
	var h strings.Builder
	h.WriteString(`<div style="font:13px/1.5 -apple-system,Segoe UI,Roboto,sans-serif;color:#6b6357;background:#f5e4dc;border-radius:8px;padding:12px 14px;margin:0 0 16px">`)
	h.WriteString("<strong>Forwarded from konorlevich.tech</strong><br>")
	h.WriteString("From: " + html.EscapeString(e.From) + "<br>")
	h.WriteString("To: " + html.EscapeString(recFor))
	if attachNote != "" {
		h.WriteString("<br>" + html.EscapeString(attachNote))
	}
	h.WriteString("</div>")

	// Plain-text banner
	var t strings.Builder
	t.WriteString("— Forwarded from konorlevich.tech —\n")
	t.WriteString("From: " + e.From + "\n")
	t.WriteString("To: " + recFor + "\n")
	if attachNote != "" {
		t.WriteString(attachNote + "\n")
	}
	t.WriteString("\n")

	return banner{html: h.String(), text: t.String()}
}

// --- Resend API calls -------------------------------------------------------

// getReceivedEmail fetches the full inbound message from Resend's
// Received-emails API. The resend-go SDK does not expose a typed method for
// this endpoint yet, so we drive it through the SDK client's NewRequest/Perform
// (reusing its auth, base URL and error handling).
func (f *Forwarder) getReceivedEmail(ctx context.Context, id string) (receivedEmail, error) {
	var out receivedEmail
	req, err := f.client.NewRequest(ctx, http.MethodGet, "emails/receiving/"+id, nil)
	if err != nil {
		return out, err
	}

	start := time.Now()
	f.log.WithField("email_id", id).Debug("resend: GET received email")
	if _, err := f.client.Perform(req, &out); err != nil {
		f.log.WithError(err).WithField("email_id", id).Error("resend: GET received email failed")
		return out, fmt.Errorf("resend get received email: %w", err)
	}
	f.log.WithFields(logrus.Fields{
		"email_id":    id,
		"duration_ms": time.Since(start).Milliseconds(),
	}).Debug("resend: GET received email response")
	return out, nil
}

func (f *Forwarder) sendEmail(ctx context.Context, msg *resend.SendEmailRequest) error {
	start := time.Now()
	f.log.WithFields(logrus.Fields{
		"to":      msg.To,
		"subject": msg.Subject,
	}).Debug("resend: POST send email")
	resp, err := f.client.Emails.SendWithContext(ctx, msg)
	if err != nil {
		f.log.WithError(err).Error("resend: POST send email failed")
		return err
	}
	f.log.WithFields(logrus.Fields{
		"to":          msg.To,
		"email_id":    resp.Id,
		"duration_ms": time.Since(start).Milliseconds(),
	}).Debug("resend: POST send email response")
	return nil
}
