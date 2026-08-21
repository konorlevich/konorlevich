package web

import (
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/konorlevich/konorlevich/internal/cv"
	"github.com/konorlevich/konorlevich/internal/site"
)

func testAssets(t *testing.T, at time.Time) *Assets {
	t.Helper()
	assets, err := LoadAssets(fstest.MapFS{
		"css/styles.css": &fstest.MapFile{Data: []byte("body{color:#000}")},
	}, at)
	require.NoError(t, err)
	return assets
}

func testCV() cv.CV {
	return cv.CV{
		Name:           "Test Person",
		Tagline:        "A tagline.",
		Email:          "test@example.test",
		Location:       "Somewhere",
		WorkExperience: []cv.Experience{{Company: "Co", Role: "Eng", From: "2020-01-01"}},
	}
}

// The checklist's explicit verification: two consecutive deploys with no content
// edits must produce a byte-identical sitemap. A <lastmod> that moves on every
// deploy is a signal Google learns to ignore and drops from scheduling.
func TestSitemapIsIdenticalAcrossDeploys(t *testing.T) {
	cfg := site.Config{BaseURL: "https://example.test"}
	entries := []SitemapEntry{{Loc: "https://example.test/", LastMod: "2026-08-04"}}

	firstDeploy := time.Date(2026, 8, 21, 9, 0, 0, 0, time.UTC)
	secondDeploy := firstDeploy.Add(72 * time.Hour)

	first, err := buildOps(cfg, testCV(), entries, testAssets(t, firstDeploy), firstDeploy)
	require.NoError(t, err)
	second, err := buildOps(cfg, testCV(), entries, testAssets(t, secondDeploy), secondDeploy)
	require.NoError(t, err)

	assert.Equal(t, string(first["/sitemap.xml"].Identity), string(second["/sitemap.xml"].Identity),
		"redeploying without a content edit must not change the sitemap")
	// The ETag is a hash of the body, so matching ETags prove byte-identity
	// independently of the string comparison above.
	assert.Equal(t, first["/sitemap.xml"].ETag, second["/sitemap.xml"].ETag,
		"a changed ETag would make crawlers refetch an unchanged sitemap")
}

func TestSitemapUsesContentDateNotBuildDate(t *testing.T) {
	buildDay := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
	entries := []SitemapEntry{{Loc: "https://example.test/", LastMod: "2026-08-04"}}

	ops, err := buildOps(site.Config{BaseURL: "https://example.test"}, testCV(), entries,
		testAssets(t, buildDay), buildDay)
	require.NoError(t, err)

	body := string(ops["/sitemap.xml"].Identity)
	assert.Contains(t, body, "<lastmod>2026-08-04</lastmod>", "must carry the content date")
	assert.NotContains(t, body, "2026-08-21", "must never carry the build date")
	assert.Contains(t, body, "<loc>https://example.test/</loc>")
}

// A page with no content date must fail the boot rather than quietly falling
// back to the build time — the exact lie this replaced.
func TestRenderAllRequiresAContentDate(t *testing.T) {
	buildTime := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)

	templates := fstest.MapFS{
		"templates/layout.html":         &fstest.MapFile{Data: []byte(`{{define "layout"}}<!doctype html>{{template "content" .}}{{end}}`)},
		"templates/partials/p.html":     &fstest.MapFile{Data: []byte(`{{define "unused"}}{{end}}`)},
		"templates/pages/home.html":     &fstest.MapFile{Data: []byte(`{{define "content"}}home{{end}}`)},
		"templates/pages/notfound.html": &fstest.MapFile{Data: []byte(`{{define "content"}}404{{end}}`)},
		"templates/pages/privacy.html":  &fstest.MapFile{Data: []byte(`{{define "content"}}privacy{{end}}`)},
	}

	t.Run("missing dates fail", func(t *testing.T) {
		content := testCV() // no Pages at all
		r := NewRenderer(testAssets(t, buildTime), site.Config{BaseURL: "https://example.test"}, content)
		_, _, err := r.RenderAll(templates, buildTime)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "updated")
	})

	t.Run("dates present succeed and reach the sitemap", func(t *testing.T) {
		content := testCV()
		content.Pages = map[string]cv.Page{
			"/":          {Updated: "2026-08-04"},
			"/privacy":   {Updated: "2026-07-01"},
			NotFoundPath: {Updated: "2026-07-01"},
		}
		r := NewRenderer(testAssets(t, buildTime), site.Config{BaseURL: "https://example.test"}, content)
		pages, entries, err := r.RenderAll(templates, buildTime)
		require.NoError(t, err)
		assert.Len(t, pages, 3)

		// Only the indexable page reaches the sitemap; noindex pages never do.
		require.Len(t, entries, 1)
		assert.Equal(t, "https://example.test/", entries[0].Loc)
		assert.Equal(t, "2026-08-04", entries[0].LastMod)
	})
}
