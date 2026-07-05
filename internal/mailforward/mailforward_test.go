package mailforward

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func testForwarder(t *testing.T) *Forwarder {
	t.Helper()
	log := logrus.New()
	log.SetOutput(io_discard{})
	f, err := New(Config{
		APIKey:        "re_test",
		From:          "konorlevich.tech <forward@konorlevich.tech>",
		To:            "inbox@example.com",
		SubjectPrefix: "[konorlevich.tech] ",
	}, log)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return f
}

type io_discard struct{}

func (io_discard) Write(p []byte) (int, error) { return len(p), nil }

func TestConfigEnabled(t *testing.T) {
	if (Config{APIKey: "k", From: "f"}).Enabled() {
		t.Error("should be disabled without To")
	}
	if !(Config{APIKey: "k", From: "f", To: "t"}).Enabled() {
		t.Error("should be enabled with API key, From and To")
	}
}

func TestBuildForward(t *testing.T) {
	f := testForwarder(t)
	txt := "hello there"
	msg := f.buildForward(receivedEmail{
		From:        "Alice <alice@somewhere.com>",
		To:          []string{"hello@konorlevich.tech"},
		ReceivedFor: []string{"hello@konorlevich.tech"},
		Subject:     "Question about your site",
		HTML:        "<p>hello there</p>",
		Text:        &txt,
		MessageID:   "<abc@mail>",
	})

	if msg.From != "konorlevich.tech <forward@konorlevich.tech>" {
		t.Errorf("From = %q", msg.From)
	}
	if len(msg.To) != 1 || msg.To[0] != "inbox@example.com" {
		t.Errorf("To = %v", msg.To)
	}
	if msg.Subject != "[konorlevich.tech] Question about your site" {
		t.Errorf("Subject = %q", msg.Subject)
	}
	// Reply-To must be the original sender so replies reach them directly.
	if msg.ReplyTo != "Alice <alice@somewhere.com>" {
		t.Errorf("ReplyTo = %q", msg.ReplyTo)
	}
	if !strings.Contains(msg.Html, "Forwarded from konorlevich.tech") {
		t.Error("HTML banner missing")
	}
	if !strings.Contains(msg.Html, "<p>hello there</p>") {
		t.Error("HTML body missing")
	}
	if !strings.Contains(msg.Text, "hello there") {
		t.Error("text body missing")
	}
	if msg.Headers["X-Original-Message-Id"] != "<abc@mail>" {
		t.Errorf("message-id header = %q", msg.Headers["X-Original-Message-Id"])
	}
}

func TestBuildForwardEscapesAndAttachments(t *testing.T) {
	f := testForwarder(t)
	msg := f.buildForward(receivedEmail{
		From:        `"E<vil>" <x@x.com>`,
		ReceivedFor: []string{"hello@konorlevich.tech"},
		Subject:     "Report",
		HTML:        "<p>body</p>",
		Attachments: []attachment{{Filename: "a.pdf"}, {Filename: "b.png"}},
		Raw:         &rawDownload{DownloadURL: "https://dl.example/x.eml"},
	})
	// The sender string must be HTML-escaped in the banner (no raw "<vil>").
	if strings.Contains(msg.Html, "E<vil>") {
		t.Error("sender not HTML-escaped in banner")
	}
	if !strings.Contains(msg.Html, "2 attachments not included") {
		t.Errorf("attachment note missing: %s", msg.Html)
	}
	if !strings.Contains(msg.Html, "https://dl.example/x.eml") {
		t.Error("raw download link missing")
	}
}

func TestDomainOf(t *testing.T) {
	cases := map[string]string{
		"hello@konorlevich.tech":  "konorlevich.tech",
		"Name <user@Example.COM>": "example.com",
		"  spaced@domain.com  ":   "domain.com",
		"no-at-sign":              "",
		"a@b@konorlevich.tech":    "konorlevich.tech",
	}
	for in, want := range cases {
		if got := domainOf(in); got != want {
			t.Errorf("domainOf(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestAddressedToDomain(t *testing.T) {
	f := testForwarder(t)

	// No domain configured → forward everything.
	if !f.addressedToDomain(receivedEmail{To: []string{"x@other.com"}}) {
		t.Error("empty domain should forward everything")
	}

	f.cfg.Domain = "konorlevich.tech"

	if !f.addressedToDomain(receivedEmail{ReceivedFor: []string{"hello@konorlevich.tech"}}) {
		t.Error("received_for on the domain should be accepted")
	}
	// Falls back to To when received_for is empty.
	if !f.addressedToDomain(receivedEmail{To: []string{"Support <hi@Konorlevich.Tech>"}}) {
		t.Error("To on the domain should be accepted (case-insensitive)")
	}
	// received_for takes precedence; an off-domain received_for is rejected even
	// if a matching address appears in To (e.g. Bcc-style delivery).
	if f.addressedToDomain(receivedEmail{
		ReceivedFor: []string{"leak@evil.com"},
		To:          []string{"hello@konorlevich.tech"},
	}) {
		t.Error("off-domain received_for should be rejected")
	}
	if f.addressedToDomain(receivedEmail{To: []string{"someone@evil.com"}}) {
		t.Error("off-domain mail should be rejected")
	}
}

func TestBuildForwardTextOnly(t *testing.T) {
	f := testForwarder(t)
	txt := "plain body"
	msg := f.buildForward(receivedEmail{
		From:    "a@b.com",
		Subject: "hi",
		Text:    &txt,
	})
	if !strings.Contains(msg.Html, "plain body") {
		t.Error("text-only message should be wrapped into HTML body")
	}
}
