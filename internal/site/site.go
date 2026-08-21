// Package site holds the per-deployment settings that templates need: the
// canonical base URL, the resolved Google tag and consent policy. Everything is
// read from the environment with safe defaults, so no id or hostname is ever
// hardcoded in a template.
package site

import (
	"fmt"
	"os"
	"strings"
)

// DefaultBaseURL is used when BASE_URL is unset (local development).
const DefaultBaseURL = "https://konorlevich.tech"

// ConsentMaxAgeDays is how long a stored consent choice stays valid before the
// banner asks again (~6 months).
const ConsentMaxAgeDays = 180

// ConsentVersion is bumped to force re-consent for everyone.
const ConsentVersion = 1

// TagKind is which flavour of Google tag an id belongs to. The id's prefix is
// resolved into this exactly once, at boot, because the kind decides both which
// snippet the page renders and how every event is sent. A GA4 measurement id
// pasted into GTM_ID and rendered as a container silently collects nothing, so
// the type — not a prefix match in a template or in JS — is what everything
// downstream branches on.
type TagKind uint8

const (
	// TagNone means no tag is configured: a clean no-op for local development
	// and previews. Nothing is emitted and the site sets no cookies at all.
	TagNone TagKind = iota
	// TagGA4 is a "G-…" measurement id, served with gtag.js. No <noscript>
	// iframe — that does nothing for gtag.
	TagGA4
	// TagGTM is a "GTM-…" container id, served with the Tag Manager snippet
	// plus its <noscript> iframe.
	TagGTM
)

// Tag is the one resolved Google tag id that reaches the templates. There is
// deliberately no way to hold two at once: two tags on one page means
// double-counted pageviews and wrecked engagement metrics.
type Tag struct {
	Kind TagKind
	ID   string
}

// IsSet reports whether a tag is configured at all.
func (t Tag) IsSet() bool { return t.Kind != TagNone }

// IsGA4 and IsGTM let templates branch on the resolved kind without ever
// inspecting the id itself.
func (t Tag) IsGA4() bool { return t.Kind == TagGA4 }
func (t Tag) IsGTM() bool { return t.Kind == TagGTM }

// JSKind is the token handed to the client as window.__tag.kind, so the
// browser's track() helper branches on the same resolved value the server used
// rather than re-deriving it from the id.
func (t Tag) JSKind() string {
	switch t.Kind {
	case TagGA4:
		return "ga4"
	case TagGTM:
		return "gtm"
	default:
		return ""
	}
}

// tagEnvVars is the precedence order for reading the tag id. GTM_ID is the
// canonical name; GA_ID and GA4_ID are accepted aliases for deployments that
// already use them. The first non-empty variable in this order wins and is the
// only one that reaches the templates — later ones are ignored, never merged.
var tagEnvVars = []string{"GTM_ID", "GA_ID", "GA4_ID"}

// ResolveTag turns a raw environment value into a typed Tag.
//
// The prefix is matched case-insensitively (ids are uppercase by convention,
// but a pasted value should not fail on case alone). An empty value is a clean
// no-op. Anything set but unrecognised is an error: a typo'd id must surface on
// the deploy, not as three months of missing data.
func ResolveTag(raw string) (Tag, error) {
	id := strings.ToUpper(strings.TrimSpace(raw))
	if id == "" {
		return Tag{}, nil
	}
	switch {
	case strings.HasPrefix(id, "GTM-"):
		return Tag{Kind: TagGTM, ID: id}, nil
	case strings.HasPrefix(id, "G-"):
		return Tag{Kind: TagGA4, ID: id}, nil
	case strings.HasPrefix(id, "UA-"):
		// Universal Analytics stopped processing data in 2023. Rendering it
		// would collect nothing, so it is a stale value to be replaced.
		return Tag{}, fmt.Errorf("site: Universal Analytics id %q is no longer processed by Google; replace it with a GA4 \"G-…\" measurement id or a \"GTM-…\" container id", id)
	default:
		return Tag{}, fmt.Errorf("site: unrecognised Google tag id %q: expected a \"G-…\" measurement id or a \"GTM-…\" container id", id)
	}
}

// Config is injected into every page's template data.
type Config struct {
	// BaseURL is the canonical origin, without a trailing slash, e.g.
	// "https://konorlevich.tech". Used to build absolute canonical/OG URLs.
	BaseURL string

	// Tag is the single resolved Google tag. A zero Tag disables analytics
	// entirely — no library, no consent bootstrap, no cookie banner.
	Tag Tag

	// TagSource names the environment variable Tag came from, for the boot log.
	// Empty when no tag is configured.
	TagSource string

	// ConsentMaxAgeDays and ConsentVersion drive the consent bootstrap.
	ConsentMaxAgeDays int
	ConsentVersion    int
}

// FromEnv builds a Config from environment variables. An unrecognised tag id is
// an error rather than a warning: consistent with the rest of the boot, a
// misconfigured deploy must fail loudly instead of serving a page that quietly
// measures nothing.
func FromEnv() (Config, error) {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("BASE_URL")), "/")
	if base == "" {
		base = DefaultBaseURL
	}

	cfg := Config{
		BaseURL:           base,
		ConsentMaxAgeDays: ConsentMaxAgeDays,
		ConsentVersion:    ConsentVersion,
	}

	for _, name := range tagEnvVars {
		raw := strings.TrimSpace(os.Getenv(name))
		if raw == "" {
			continue
		}
		tag, err := ResolveTag(raw)
		if err != nil {
			return Config{}, fmt.Errorf("%s: %w", name, err)
		}
		cfg.Tag = tag
		cfg.TagSource = name
		break
	}
	return cfg, nil
}

// NeedsConsent reports whether a consent-gated tag is configured. When it is
// false the cookie banner and consent scripts are omitted from the page — the
// site sets no cookies at all, so there is nothing to ask about.
func (c Config) NeedsConsent() bool { return c.Tag.IsSet() }

// URL joins the base URL with an absolute path ("/privacy" -> ".../privacy").
func (c Config) URL(path string) string {
	if path == "" || path == "/" {
		return c.BaseURL + "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return c.BaseURL + path
}
