package site

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveTag(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		wantKind TagKind
		wantID   string
		wantJS   string
		wantErr  string
	}{
		{name: "empty is a clean no-op", raw: "", wantKind: TagNone, wantJS: ""},
		{name: "whitespace only is a clean no-op", raw: "   ", wantKind: TagNone, wantJS: ""},
		{name: "GA4 measurement id", raw: "G-RH8KWHKMPZ", wantKind: TagGA4, wantID: "G-RH8KWHKMPZ", wantJS: "ga4"},
		{name: "GTM container id", raw: "GTM-ABC1234", wantKind: TagGTM, wantID: "GTM-ABC1234", wantJS: "gtm"},
		{name: "surrounding whitespace is trimmed", raw: "  G-RH8KWHKMPZ\n", wantKind: TagGA4, wantID: "G-RH8KWHKMPZ", wantJS: "ga4"},
		{name: "lowercase is normalised, not rejected", raw: "g-rh8kwhkmpz", wantKind: TagGA4, wantID: "G-RH8KWHKMPZ", wantJS: "ga4"},
		{name: "lowercase container is normalised", raw: "gtm-abc1234", wantKind: TagGTM, wantID: "GTM-ABC1234", wantJS: "gtm"},

		// GTM- must win over G-, since "GTM-x" also starts with "G".
		{name: "GTM is not mistaken for GA4", raw: "GTM-G1234", wantKind: TagGTM, wantID: "GTM-G1234", wantJS: "gtm"},

		{name: "Universal Analytics is rejected", raw: "UA-12345-1", wantErr: "no longer processed"},
		{name: "lowercase UA is also rejected", raw: "ua-12345-1", wantErr: "no longer processed"},
		{name: "unknown prefix is fatal", raw: "XX-12345", wantErr: "unrecognised"},
		{name: "bare word is fatal", raw: "analytics", wantErr: "unrecognised"},
		{name: "near-miss without hyphen is fatal", raw: "GTM12345", wantErr: "unrecognised"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tag, err := ResolveTag(tc.raw)

			if tc.wantErr != "" {
				require.Error(t, err, "a set-but-unrecognised id must fail the boot, not warn")
				assert.Contains(t, err.Error(), tc.wantErr)
				assert.False(t, tag.IsSet(), "a rejected id must not leak a usable tag")
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantKind, tag.Kind)
			assert.Equal(t, tc.wantID, tag.ID)
			assert.Equal(t, tc.wantJS, tag.JSKind())
		})
	}
}

// Exactly one branch may ever be true: two tags on one page double-counts every
// pageview, which is the failure the typed Tag exists to make impossible.
func TestTagBranchesAreMutuallyExclusive(t *testing.T) {
	for _, raw := range []string{"", "G-ABC123", "GTM-ABC123"} {
		tag, err := ResolveTag(raw)
		require.NoError(t, err)

		emitting := 0
		if tag.IsGA4() {
			emitting++
		}
		if tag.IsGTM() {
			emitting++
		}
		assert.LessOrEqual(t, emitting, 1, "more than one tag path would emit for %q", raw)
		assert.Equal(t, tag.IsSet(), emitting == 1, "IsSet must agree with whether a path emits")
	}
}

func TestFromEnvTagPrecedence(t *testing.T) {
	t.Run("GTM_ID is canonical and wins over both aliases", func(t *testing.T) {
		t.Setenv("GTM_ID", "GTM-CANONICAL")
		t.Setenv("GA_ID", "G-ALIAS1")
		t.Setenv("GA4_ID", "G-ALIAS2")

		cfg, err := FromEnv()
		require.NoError(t, err)
		assert.Equal(t, TagGTM, cfg.Tag.Kind)
		assert.Equal(t, "GTM-CANONICAL", cfg.Tag.ID)
		assert.Equal(t, "GTM_ID", cfg.TagSource)
	})

	t.Run("GA_ID is used when GTM_ID is unset", func(t *testing.T) {
		t.Setenv("GTM_ID", "")
		t.Setenv("GA_ID", "G-FROMGAID")
		t.Setenv("GA4_ID", "G-FROMGA4ID")

		cfg, err := FromEnv()
		require.NoError(t, err)
		assert.Equal(t, TagGA4, cfg.Tag.Kind)
		assert.Equal(t, "G-FROMGAID", cfg.Tag.ID)
		assert.Equal(t, "GA_ID", cfg.TagSource)
	})

	t.Run("GA4_ID is the last resort", func(t *testing.T) {
		t.Setenv("GTM_ID", "")
		t.Setenv("GA_ID", "")
		t.Setenv("GA4_ID", "G-FROMGA4ID")

		cfg, err := FromEnv()
		require.NoError(t, err)
		assert.Equal(t, "G-FROMGA4ID", cfg.Tag.ID)
		assert.Equal(t, "GA4_ID", cfg.TagSource)
	})

	t.Run("no tag configured disables consent entirely", func(t *testing.T) {
		t.Setenv("GTM_ID", "")
		t.Setenv("GA_ID", "")
		t.Setenv("GA4_ID", "")

		cfg, err := FromEnv()
		require.NoError(t, err)
		assert.False(t, cfg.Tag.IsSet())
		assert.False(t, cfg.NeedsConsent(), "no tag means no cookies, so nothing to ask about")
		assert.Empty(t, cfg.TagSource)
		assert.Empty(t, cfg.Tag.JSKind())
	})

	t.Run("a typo'd id fails the boot", func(t *testing.T) {
		t.Setenv("GTM_ID", "OOPS-123")
		t.Setenv("GA_ID", "")
		t.Setenv("GA4_ID", "")

		_, err := FromEnv()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "GTM_ID", "the error must name the offending variable")
	})

	t.Run("a rejected alias still fails the boot", func(t *testing.T) {
		t.Setenv("GTM_ID", "")
		t.Setenv("GA_ID", "UA-999-1")
		t.Setenv("GA4_ID", "")

		_, err := FromEnv()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "GA_ID")
	})
}

func TestFromEnvBaseURL(t *testing.T) {
	t.Run("defaults when unset", func(t *testing.T) {
		t.Setenv("BASE_URL", "")
		cfg, err := FromEnv()
		require.NoError(t, err)
		assert.Equal(t, DefaultBaseURL, cfg.BaseURL)
	})

	t.Run("trailing slash is stripped so URL() never doubles it", func(t *testing.T) {
		t.Setenv("BASE_URL", "https://example.test/")
		cfg, err := FromEnv()
		require.NoError(t, err)
		assert.Equal(t, "https://example.test", cfg.BaseURL)
		assert.Equal(t, "https://example.test/", cfg.URL("/"))
		assert.Equal(t, "https://example.test/privacy", cfg.URL("/privacy"))
		assert.False(t, strings.Contains(cfg.URL("/privacy"), "//privacy"))
	})
}
