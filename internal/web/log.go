package web

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// requestIDHeader is echoed back so a client (or the platform) can correlate a
// response with its log line.
const requestIDHeader = "X-Request-Id"

// LogRequests emits exactly one structured line per request: enough to feed any
// log aggregator unmodified, and never anything from a form body or cookie.
func LogRequests(next http.Handler, log logrus.FieldLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		id := r.Header.Get(requestIDHeader)
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set(requestIDHeader, id)

		lw := &logResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(lw, r)

		log.WithFields(logrus.Fields{
			"request_id":  id,
			"method":      r.Method,
			"path":        r.URL.Path,
			"status":      lw.status,
			"bytes":       lw.bytes,
			"duration_ms": time.Since(start).Milliseconds(),
			"client_ip":   clientIP(r),
			"referer":     r.Referer(),
			"user_agent":  r.UserAgent(),
			"protocol":    r.Proto,
		}).Info("http request")
	})
}

// newRequestID returns a short random hex id. crypto/rand cannot fail in
// practice; on the impossible error path the timestamp still yields a usable id.
func newRequestID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(b[:])
}

// clientIP prefers the first hop in X-Forwarded-For (Railway terminates TLS in
// front of the app), falling back to the socket address.
func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		if first, _, found := strings.Cut(fwd, ","); found {
			return strings.TrimSpace(first)
		}
		return strings.TrimSpace(fwd)
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

// logResponseWriter captures the status code and byte count for the access log.
type logResponseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *logResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *logResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}
