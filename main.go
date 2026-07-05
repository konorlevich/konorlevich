package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-pdf/fpdf"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/konorlevich/konorlevich/internal/config"
	"github.com/konorlevich/konorlevich/internal/cv"
	"github.com/konorlevich/konorlevich/internal/mailforward"
)

const (
	defaultConfigFilename   = "config.yaml"
	templateFilename        = "cv_template.html"
	privacyTemplateFilename = "privacy_template.html"
)

var (
	cfg         config.Config
	l           *log.Logger
	pageTmpl    *template.Template
	privacyTmpl *template.Template
)

// templateFuncs are helpers available inside cv_template.html.
var templateFuncs = template.FuncMap{
	// formatDate turns "2006-01-02" into "Jan 2006"; passes through on error.
	"formatDate": func(s string) string {
		if s == "" {
			return "Present"
		}
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return s
		}
		return t.Format("Jan 2006")
	},
	// initials returns up to two uppercase initials from a full name.
	"initials": func(name string) string {
		parts := strings.Fields(name)
		var b strings.Builder
		for _, p := range parts {
			if p == "" {
				continue
			}
			b.WriteString(strings.ToUpper(p[:1]))
			if b.Len() >= 2 {
				break
			}
		}
		return b.String()
	},
	// join concatenates a string slice with ", ".
	"join": func(items []string) string { return strings.Join(items, ", ") },
	// webpOf swaps a .jpg/.jpeg/.png path for its .webp sibling (for <picture> sources).
	"webpOf": func(p string) string {
		for _, ext := range []string{".jpg", ".jpeg", ".png", ".JPG", ".JPEG", ".PNG"} {
			if strings.HasSuffix(p, ext) {
				return strings.TrimSuffix(p, ext) + ".webp"
			}
		}
		return p
	},
}

func main() {
	l = log.New()
	l.SetFormatter(&log.JSONFormatter{})

	// LOG_LEVEL controls verbosity (e.g. debug enables Resend API call logs).
	if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
		parsed, err := log.ParseLevel(lvl)
		if err != nil {
			l.WithError(err).WithField("log_level", lvl).Warn("invalid LOG_LEVEL, keeping default")
		} else {
			l.SetLevel(parsed)
		}
	}

	configFile := os.Getenv("CONFIG_FILE")

	if configFile == "" {
		configFile = defaultConfigFilename
	}

	l.WithField("config_file", configFile).Info("using config file")
	var err error
	cfg, err = config.Load(configFile)
	if err != nil {
		l.WithError(err).Fatal("failed to load config")
	}

	// Parse the page template once at startup (fail fast on a bad template).
	pageTmpl, err = template.New(templateFilename).Funcs(templateFuncs).ParseFiles(templateFilename)
	if err != nil {
		l.WithError(err).Fatal("failed to parse page template")
	}

	// Parse the privacy/cookies page template once at startup.
	privacyTmpl, err = template.New(privacyTemplateFilename).Funcs(templateFuncs).ParseFiles(privacyTemplateFilename)
	if err != nil {
		l.WithError(err).Fatal("failed to parse privacy template")
	}

	appCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	defer stop()

	handler := &http.ServeMux{}
	handler.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	handler.Handle("GET /", serveHTML())
	handler.Handle("GET /cv", serveHTML())
	handler.Handle("GET /privacy", servePrivacy())
	handler.Handle("GET /cv/download", servePDF())
	handler.Handle("POST /contact", submitContactForm())

	// Inbound email forwarding (Resend webhook → forward to FORWARD_TO).
	// Enabled only when RESEND_API_KEY, RESEND_FROM and FORWARD_TO are set.
	if fwdCfg := mailforward.ConfigFromEnv(); fwdCfg.Enabled() {
		fwd, err := mailforward.New(fwdCfg, l)
		if err != nil {
			l.WithError(err).Fatal("failed to init mail forwarder")
		}
		handler.Handle("POST /webhooks/resend/inbound", fwd.Handler())
		l.Info("inbound email forwarding enabled at POST /webhooks/resend/inbound")
	} else {
		l.Info("inbound email forwarding disabled (set RESEND_API_KEY, RESEND_FROM, FORWARD_TO to enable)")
	}

	server := &http.Server{
		Addr:    cfg.App.Address,
		Handler: logRequests(handler, l),
	}

	go func() {
		select {
		case <-appCtx.Done():
			return
		default:
		}

		l.WithField("app_address", cfg.App.Address).Info("Starting server")
		if err := server.ListenAndServe(); err != nil {
			l.WithError(err).Error("failed to start server")
			stop()
		}
	}()

	<-appCtx.Done()
	l.Info("shutting down")
	shutdownContext, cancelShutdownContext := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelShutdownContext()
	err = server.Shutdown(shutdownContext)
	if err != nil {
		l.WithError(err).Fatalf("can't shut the server down gracefully")
	}

}

// logRequests is HTTP access-log middleware: it logs one line per request with
// method, path, status, response size and latency once the handler returns.
func logRequests(next http.Handler, l *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lw := &logResponseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(lw, r)

		l.WithFields(log.Fields{
			"method":      r.Method,
			"path":        r.URL.Path,
			"status":      lw.status,
			"bytes":       lw.bytes,
			"duration_ms": time.Since(start).Milliseconds(),
			"remote_addr": r.RemoteAddr,
			"user_agent":  r.UserAgent(),
		}).Info("http request")
	})
}

// logResponseWriter wraps http.ResponseWriter to capture the status code and
// number of bytes written for access logging.
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

// Handle contact form submission
func submitContactForm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form submission", http.StatusBadRequest)
			return
		}

		// Send email
		//TODO: Notify

		_, _ = w.Write([]byte("Contact form submitted successfully"))
	}
}

// Serve HTML page
func serveHTML() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cvData, err := readCV()
		if err != nil {
			l.WithError(err).Error("failed to read CV file")
			http.Error(w, "Could not read CV", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := pageTmpl.Execute(w, cvData); err != nil {
			l.WithError(err).Error("can't generate page")
			http.Error(w, "Could not generate page", http.StatusInternalServerError)
			return
		}
	}
}

// Serve the privacy / cookies page
func servePrivacy() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cvData, err := readCV()
		if err != nil {
			l.WithError(err).Error("failed to read CV file")
			http.Error(w, "Could not read CV", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := privacyTmpl.Execute(w, cvData); err != nil {
			l.WithError(err).Error("can't generate privacy page")
			http.Error(w, "Could not generate page", http.StatusInternalServerError)
			return
		}
	}
}

// Serve PDF file
func servePDF() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cv, err := readCV()
		if err != nil {
			http.Error(w, "Could not read CV", http.StatusInternalServerError)
			return
		}

		pdf := fpdf.New("P", "mm", "A4", "")
		pdf.AddPage()
		// Translate UTF-8 content into the core font's cp1252 encoding so
		// characters like em dashes and × render correctly (not mojibake).
		tr := pdf.UnicodeTranslatorFromDescriptor("")

		// Name
		pdf.SetFont("Arial", "B", 16)
		pdf.Cell(40, 10, tr(cv.Name))
		pdf.Ln(9)

		// Location · Availability · Languages — one quiet meta line
		meta := make([]string, 0, 3)
		if cv.Location != "" {
			meta = append(meta, cv.Location)
		}
		if cv.Availability != "" {
			meta = append(meta, cv.Availability)
		}
		if len(cv.Languages) > 0 {
			meta = append(meta, strings.Join(cv.Languages, ", "))
		}
		if len(meta) > 0 {
			pdf.SetFont("Arial", "", 11)
			pdf.SetTextColor(107, 99, 87) // warm secondary
			pdf.MultiCell(190, 6, tr(strings.Join(meta, "  ·  ")), "", "L", false)
			pdf.SetTextColor(0, 0, 0)
		}

		// Summary paragraph
		if cv.Summary != "" {
			pdf.Ln(1)
			pdf.SetFont("Arial", "", 11)
			pdf.MultiCell(190, 6, tr(cv.Summary), "", "L", false)
		}
		pdf.Ln(4)

		// Links section — rendered as clickable hyperlinks
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(40, 10, "Links")
		pdf.Ln(8)
		pdf.SetFont("Arial", "", 12)
		for _, link := range cv.Links {
			label := tr(link.Name + ": ")
			pdf.CellFormat(pdf.GetStringWidth(label)+1, 8, label, "", 0, "L", false, 0, "")
			// display without the mailto: scheme, but keep it as the link target
			display := strings.TrimPrefix(link.URL, "mailto:")
			pdf.SetTextColor(168, 72, 42) // terracotta accent
			pdf.SetFont("Arial", "U", 12) // underlined
			pdf.CellFormat(0, 8, tr(display), "", 1, "L", false, 0, link.URL)
			pdf.SetFont("Arial", "", 12)
			pdf.SetTextColor(0, 0, 0)
		}

		// Work Experience
		pdf.Ln(2)
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(40, 10, "Work Experience")
		pdf.Ln(10)

		for _, exp := range cv.WorkExperience {
			to := exp.To
			if to == "" {
				to = "Present"
			}
			pdf.SetFont("Arial", "B", 12)
			pdf.Cell(40, 8, tr(fmt.Sprintf("%s - %s", exp.Company, exp.Role)))
			pdf.Ln(6)
			pdf.SetFont("Arial", "", 12)
			pdf.Cell(40, 8, tr(fmt.Sprintf("From: %s to %s", exp.From, to)))
			pdf.Ln(8)

			pdf.Cell(40, 8, fmt.Sprintf("Skills: %s", tr(strings.Join(exp.Skills, ", "))))
			pdf.Ln(8)

			pdf.Cell(40, 8, "Achievements:")
			pdf.Ln(6)
			for _, achievement := range exp.Achievements {
				pdf.MultiCell(190, 6, tr(fmt.Sprintf("- %s", achievement)), "", "", false)
			}
			pdf.Ln(8)
		}

		// Projects
		if len(cv.Projects) > 0 {
			pdf.SetFont("Arial", "B", 14)
			pdf.Cell(40, 10, "Projects")
			pdf.Ln(10)
			for _, p := range cv.Projects {
				title := p.Name
				if p.Badge != "" {
					title = fmt.Sprintf("%s (%s)", p.Name, p.Badge)
				}
				pdf.SetFont("Arial", "B", 12)
				pdf.Cell(40, 8, tr(title))
				pdf.Ln(6)
				if p.URL != "" {
					pdf.SetFont("Arial", "U", 11)
					pdf.SetTextColor(168, 72, 42) // terracotta accent
					pdf.CellFormat(0, 6, tr(p.URL), "", 1, "L", false, 0, p.URL)
					pdf.SetTextColor(0, 0, 0)
				}
				if p.Description != "" {
					pdf.SetFont("Arial", "", 11)
					pdf.MultiCell(190, 6, tr(p.Description), "", "L", false)
				}
				pdf.Ln(4)
			}
		}

		// Serve inline with a sensible filename if saved.
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition",
			fmt.Sprintf("inline; filename=%q", strings.ReplaceAll(cv.Name, " ", "-")+"-CV.pdf"))
		if err = pdf.Output(w); err != nil {
			http.Error(w, "Could not generate PDF", http.StatusInternalServerError)
		}
	}
}

// Read CV YAML file
func readCV() (cv.CV, error) {
	var cv cv.CV
	data, err := os.ReadFile("cv.yaml")
	if err != nil {
		return cv, err
	}
	err = yaml.Unmarshal(data, &cv)
	return cv, err
}

//// Send email using SMTP
//func sendEmail(email, link, position string) error {
//	auth := smtp.PlainAuth("", cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.Server)
//	to := []string{cfg.SMTP.From}
//	subject := "New Job Opportunity"
//	body := fmt.Sprintf("Email: %s\nLink: %s\nPosition: %s", email, link, position)
//	msg := []byte("To: " + cfg.SMTP.From + "\r\n" +
//		"Subject: " + subject + "\r\n" +
//		"\r\n" + body + "\r\n")
//	err := smtp.SendMail(fmt.Sprintf("%s:%d", cfg.SMTP.Server, cfg.SMTP.Port), auth, cfg.SMTP.From, to, msg)
//	return err
//}
