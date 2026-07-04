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
	defaultConfigFilename = "config.yaml"
	templateFilename      = "cv_template.html"
)

var (
	cfg      config.Config
	l        *log.Logger
	pageTmpl *template.Template
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

	appCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	defer stop()

	//conn, err := kafka.DialLeader(appCtx, "tcp", cfg.Kafka.Address, cfg.Kafka.Topic.Name, cfg.Kafka.Topic.Partition)
	//if err != nil {
	//	l.WithError(err).Fatal("failed to dial kafka leader")
	//}
	//defer func() {
	//	if err = conn.Close(); err != nil {
	//		l.WithError(err).Error("failed to close writer:", err)
	//	}
	//}()

	//err = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	//if err != nil {
	//	l.WithError(err).Fatal("failed to set write kafka connection deadline")
	//}

	handler := &http.ServeMux{}
	handler.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	handler.Handle("GET /", serveHTML())
	handler.Handle("GET /cv", serveHTML())
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
		Handler: handler,
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

		// Add CV info into PDF
		pdf.SetFont("Arial", "B", 16)
		pdf.Cell(40, 10, cv.Name)
		pdf.Ln(10)

		// Links section
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(40, 10, "Links")
		pdf.Ln(8)
		pdf.SetFont("Arial", "", 12)
		for _, link := range cv.Links {
			pdf.Cell(40, 10, fmt.Sprintf("%s: %s", link.Name, link.URL))
			pdf.Ln(8)
		}

		// Work Experience
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(40, 10, "Work Experience")
		pdf.Ln(10)

		pdf.SetFont("Arial", "", 12)
		for _, exp := range cv.WorkExperience {
			pdf.Cell(40, 10, fmt.Sprintf("%s - %s", exp.Company, exp.Role))
			pdf.Ln(8)
			pdf.Cell(40, 10, fmt.Sprintf("From: %s to %s", exp.From, func() string {
				if exp.To == "" {
					return "Present"
				}
				return exp.To
			}()))
			pdf.Ln(8)

			// Add Skills
			pdf.Cell(40, 10, "Skills:")
			pdf.Ln(6)
			for _, skill := range exp.Skills {
				pdf.Cell(40, 10, fmt.Sprintf("- %s", skill))
				pdf.Ln(6)
			}

			// Add Achievements
			pdf.Ln(6)
			pdf.Cell(40, 10, "Achievements:")
			pdf.Ln(6)
			for _, achievement := range exp.Achievements {
				pdf.MultiCell(190, 10, fmt.Sprintf("- %s", achievement), "", "", false)
			}
			pdf.Ln(12)
		}

		err = pdf.Output(w)
		if err != nil {
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
