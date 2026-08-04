// Command konorlevich serves the personal site: a single self-contained binary
// with every template, asset and piece of content embedded.
//
// main is wiring only — parse config, build the site, run the server. All the
// behaviour lives in internal/.
package main

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/konorlevich/konorlevich/internal/config"
	"github.com/konorlevich/konorlevich/internal/cv"
	"github.com/konorlevich/konorlevich/internal/mailforward"
	"github.com/konorlevich/konorlevich/internal/site"
	"github.com/konorlevich/konorlevich/internal/web"
)

// Everything the service serves travels inside the binary. Startup fails fast
// if any of it is missing or malformed.
var (
	//go:embed templates
	templateFS embed.FS

	//go:embed static
	staticFS embed.FS

	//go:embed cv.yaml
	cvYAML []byte

	//go:embed config.yaml
	defaultConfigYAML []byte
)

// shutdownTimeout bounds how long in-flight requests get to drain. Railway
// sends SIGTERM on redeploy, so a clean exit here is what makes deploys
// zero-downtime.
const shutdownTimeout = 10 * time.Second

func main() {
	logger := newLogger()

	cfg, err := loadConfig(logger)
	if err != nil {
		logger.WithError(err).Fatal("failed to load config")
	}

	content, err := cv.Parse(cvYAML)
	if err != nil {
		logger.WithError(err).Fatal("failed to load CV content")
	}

	staticRoot, err := fs.Sub(staticFS, "static")
	if err != nil {
		logger.WithError(err).Fatal("failed to open embedded static assets")
	}

	siteCfg := site.FromEnv()
	server, err := web.New(web.Options{
		Static:    staticRoot,
		Templates: templateFS,
		CV:        content,
		Site:      siteCfg,
		Log:       logger,
		BuildTime: time.Now(),
	})
	if err != nil {
		logger.WithError(err).Fatal("failed to build site")
	}

	mux := server.Handler()

	// Inbound email forwarding (Resend webhook → forward to FORWARD_TO).
	// Registered only when RESEND_API_KEY, RESEND_FROM and FORWARD_TO are set.
	if fwdCfg := mailforward.ConfigFromEnv(); fwdCfg.Enabled() {
		fwd, err := mailforward.New(fwdCfg, logger)
		if err != nil {
			logger.WithError(err).Fatal("failed to init mail forwarder")
		}
		mux.Handle("POST /webhooks/resend/inbound", fwd.Handler())
		logger.Info("inbound email forwarding enabled at POST /webhooks/resend/inbound")
	} else {
		logger.Info("inbound email forwarding disabled (set RESEND_API_KEY, RESEND_FROM, FORWARD_TO to enable)")
	}

	logger.WithFields(log.Fields{
		"base_url":  siteCfg.BaseURL,
		"analytics": siteCfg.GAID != "",
		"gtm":       siteCfg.GTMID != "",
	}).Info("site built")

	if err := run(cfg.App.Address, web.LogRequests(mux, logger), logger); err != nil {
		logger.WithError(err).Fatal("server error")
	}
}

// run starts the HTTP server and blocks until a termination signal arrives,
// then drains in-flight requests within shutdownTimeout.
func run(addr string, handler http.Handler, logger log.FieldLogger) error {
	// SIGKILL is deliberately absent: it cannot be caught, so listening for it
	// would be a no-op that implies a guarantee the process cannot make.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errs := make(chan error, 1)
	go func() {
		logger.WithField("address", addr).Info("starting server")
		// ErrServerClosed is the expected result of a graceful shutdown, not a
		// failure, so it never reaches the error channel.
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}
	}()

	select {
	case err := <-errs:
		return err
	case <-ctx.Done():
		logger.Info("shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}
	logger.Info("stopped cleanly")
	return nil
}

// newLogger configures the one shared structured logger: JSON, one object per
// line, to stdout.
func newLogger() *log.Logger {
	logger := log.New()
	logger.SetFormatter(&log.JSONFormatter{})
	logger.SetOutput(os.Stdout)

	if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
		parsed, err := log.ParseLevel(lvl)
		if err != nil {
			logger.WithError(err).WithField("log_level", lvl).Warn("invalid LOG_LEVEL, keeping default")
		} else {
			logger.SetLevel(parsed)
		}
	}
	return logger
}

// loadConfig prefers an explicit CONFIG_FILE, falls back to the embedded
// default, and lets PORT override the listen address (the platform convention).
func loadConfig(logger log.FieldLogger) (config.Config, error) {
	var (
		cfg config.Config
		err error
	)
	if path := os.Getenv("CONFIG_FILE"); path != "" {
		logger.WithField("config_file", path).Info("loading config from file")
		cfg, err = config.Load(path)
	} else {
		cfg, err = config.Parse(defaultConfigYAML)
	}
	if err != nil {
		return config.Config{}, err
	}

	if port := os.Getenv("PORT"); port != "" {
		cfg.App.Address = ":" + port
	}
	return cfg, nil
}
