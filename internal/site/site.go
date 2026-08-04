// Package site holds the per-deployment settings that templates need: the
// canonical base URL, analytics ids and consent policy. Everything is read from
// the environment with safe defaults, so no id or hostname is ever hardcoded in
// a template.
package site

import (
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

// Config is injected into every page's template data.
type Config struct {
	// BaseURL is the canonical origin, without a trailing slash, e.g.
	// "https://konorlevich.tech". Used to build absolute canonical/OG URLs.
	BaseURL string

	// GAID is the Google Analytics measurement id (GA_ID). Empty disables GA
	// entirely — no library, no consent bootstrap, no cookie banner.
	GAID string

	// GTMID is the Google Tag Manager container id (GTM_ID). Empty emits
	// nothing at all.
	GTMID string

	// ConsentMaxAgeDays and ConsentVersion drive the consent bootstrap.
	ConsentMaxAgeDays int
	ConsentVersion    int
}

// FromEnv builds a Config from environment variables.
func FromEnv() Config {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("BASE_URL")), "/")
	if base == "" {
		base = DefaultBaseURL
	}
	return Config{
		BaseURL:           base,
		GAID:              strings.TrimSpace(os.Getenv("GA_ID")),
		GTMID:             strings.TrimSpace(os.Getenv("GTM_ID")),
		ConsentMaxAgeDays: ConsentMaxAgeDays,
		ConsentVersion:    ConsentVersion,
	}
}

// NeedsConsent reports whether any consent-gated tag is configured. When it is
// false the cookie banner and consent scripts are omitted from the page — the
// site sets no cookies at all, so there is nothing to ask about.
func (c Config) NeedsConsent() bool { return c.GAID != "" || c.GTMID != "" }

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
