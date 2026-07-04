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
	if len(msg.ReplyTo) != 1 || msg.ReplyTo[0] != "Alice <alice@somewhere.com>" {
		t.Errorf("ReplyTo = %v", msg.ReplyTo)
	}
	if !strings.Contains(msg.HTML, "Forwarded from konorlevich.tech") {
		t.Error("HTML banner missing")
	}
	if !strings.Contains(msg.HTML, "<p>hello there</p>") {
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
	if strings.Contains(msg.HTML, "E<vil>") {
		t.Error("sender not HTML-escaped in banner")
	}
	if !strings.Contains(msg.HTML, "2 attachments not included") {
		t.Errorf("attachment note missing: %s", msg.HTML)
	}
	if !strings.Contains(msg.HTML, "https://dl.example/x.eml") {
		t.Error("raw download link missing")
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
	if !strings.Contains(msg.HTML, "plain body") {
		t.Error("text-only message should be wrapped into HTML body")
	}
}
