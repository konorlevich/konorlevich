package web

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
)

// minCompressBytes is the size below which compressing is not worth the header
// overhead (and often makes the payload larger).
const minCompressBytes = 512

// Blob is one immutable response body, compressed once at boot and then only
// ever written to the wire. Everything the hot path needs — the encodings, the
// ETag, the content type — is precomputed here.
type Blob struct {
	ContentType string
	Identity    []byte
	Brotli      []byte // nil when compression was skipped or unhelpful
	Gzip        []byte // nil when compression was skipped or unhelpful
	ETag        string
	ModTime     time.Time

	// CacheControl is sent verbatim. Content-addressed assets get an immutable
	// year; HTML gets no-cache so a redeploy is picked up immediately.
	CacheControl string
}

// NewBlob compresses body once and computes its ETag.
func NewBlob(contentType string, body []byte, cacheControl string, modTime time.Time) *Blob {
	sum := sha256.Sum256(body)
	b := &Blob{
		ContentType:  contentType,
		Identity:     body,
		ETag:         `"` + hex.EncodeToString(sum[:])[:16] + `"`,
		ModTime:      modTime,
		CacheControl: cacheControl,
	}
	if len(body) >= minCompressBytes && compressible(contentType) {
		b.Brotli = smaller(brotliEncode(body), body)
		b.Gzip = smaller(gzipEncode(body), body)
	}
	return b
}

// compressible reports whether a content type benefits from compression.
// Already-compressed formats (woff2, webp, png, jpeg) are left alone.
func compressible(ct string) bool {
	ct = strings.ToLower(ct)
	switch {
	case strings.HasPrefix(ct, "text/"),
		strings.Contains(ct, "javascript"),
		strings.Contains(ct, "json"),
		strings.Contains(ct, "xml"),
		strings.Contains(ct, "svg"):
		return true
	}
	return false
}

// smaller returns encoded only when it actually beat the original.
func smaller(encoded, original []byte) []byte {
	if len(encoded) == 0 || len(encoded) >= len(original) {
		return nil
	}
	return encoded
}

func brotliEncode(b []byte) []byte {
	var buf bytes.Buffer
	w := brotli.NewWriterLevel(&buf, brotli.BestCompression)
	if _, err := w.Write(b); err != nil {
		return nil
	}
	if err := w.Close(); err != nil {
		return nil
	}
	return buf.Bytes()
}

func gzipEncode(b []byte) []byte {
	var buf bytes.Buffer
	w, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil
	}
	if _, err := w.Write(b); err != nil {
		return nil
	}
	if err := w.Close(); err != nil {
		return nil
	}
	return buf.Bytes()
}

// Serve writes the blob, negotiating the encoding and honouring conditional
// requests. status lets the same machinery serve a 200 page and a 404 page.
func (b *Blob) Serve(w http.ResponseWriter, r *http.Request, status int) {
	h := w.Header()
	h.Set("Content-Type", b.ContentType)
	h.Set("Cache-Control", b.CacheControl)
	h.Set("ETag", b.ETag)
	h.Set("Vary", "Accept-Encoding")
	if !b.ModTime.IsZero() {
		h.Set("Last-Modified", b.ModTime.UTC().Format(http.TimeFormat))
	}

	// Conditional request: nothing changed, so send headers only.
	if match := r.Header.Get("If-None-Match"); match != "" && etagMatches(match, b.ETag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	body := b.Identity
	switch negotiate(r.Header.Get("Accept-Encoding")) {
	case "br":
		if b.Brotli != nil {
			body = b.Brotli
			h.Set("Content-Encoding", "br")
		}
	case "gzip":
		if b.Gzip != nil {
			body = b.Gzip
			h.Set("Content-Encoding", "gzip")
		}
	}

	h.Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(status)
	if r.Method != http.MethodHead {
		_, _ = w.Write(body)
	}
}

// negotiate picks the best encoding we have, preferring Brotli. An encoding
// listed with q=0 is explicitly refused and skipped.
func negotiate(accept string) string {
	if accept == "" {
		return ""
	}
	var hasBr, hasGzip bool
	for part := range strings.SplitSeq(accept, ",") {
		name, params, _ := strings.Cut(strings.TrimSpace(part), ";")
		name = strings.TrimSpace(strings.ToLower(name))
		if name != "br" && name != "gzip" {
			continue
		}
		if qualityOf(params) == 0 {
			continue
		}
		if name == "br" {
			hasBr = true
		} else {
			hasGzip = true
		}
	}
	if hasBr {
		return "br"
	}
	if hasGzip {
		return "gzip"
	}
	return ""
}

// qualityOf reads the q= parameter of an Accept-Encoding entry, defaulting to 1
// when absent or unparseable.
func qualityOf(params string) float64 {
	for p := range strings.SplitSeq(params, ";") {
		key, value, found := strings.Cut(strings.TrimSpace(p), "=")
		if !found || strings.ToLower(strings.TrimSpace(key)) != "q" {
			continue
		}
		q, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return 1
		}
		return q
	}
	return 1
}

// etagMatches implements the If-None-Match comparison, including "*" and
// comma-separated lists, using weak comparison as RFC 9110 requires.
func etagMatches(header, etag string) bool {
	if strings.TrimSpace(header) == "*" {
		return true
	}
	want := strings.TrimPrefix(etag, "W/")
	for candidate := range strings.SplitSeq(header, ",") {
		got := strings.TrimPrefix(strings.TrimSpace(candidate), "W/")
		if got == want {
			return true
		}
	}
	return false
}
