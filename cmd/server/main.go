package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/valorm/snapurl/internal/api"
	"github.com/valorm/snapurl/internal/config"
	"github.com/valorm/snapurl/internal/datastore"
	"github.com/valorm/snapurl/internal/limiter"
	"github.com/valorm/snapurl/internal/telemetry"
)

func main() {
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Open (and migrate) DB
	db, err := datastore.OpenDB(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize rate limiter & telemetry
	rateLimiter := limiter.NewIPRateLimiter(cfg.RateLimit)
	telemetry.Init()

	// Build router
	mux := http.NewServeMux()

	// Public endpoints
	mux.Handle("/shorten", api.ShortenHandler(db))
	mux.Handle("/{shortcode}", api.RedirectHandler(db))
	mux.Handle("/health", api.HealthHandler())
	mux.Handle("/metrics", api.MetricsHandler(db))

	// Protected endpoint with authentication middleware
	mux.Handle("/"+cfg.APIKeys[0], api.AuthMiddleware(cfg, api.RevokeHandler(db, cfg.APIKeys)))

	// Serve the UI
	mux.Handle("/", http.FileServer(http.Dir("public")))

	// Apply middleware: recovery → logging → rate limiting
	handler := rateLimiter.Middleware(
		api.RecoveryMiddleware(
			api.LoggingMiddleware(mux),
		),
	)

	server := &http.Server{
		Addr:    cfg.Port,
		Handler: handler,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("Starting server on %s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
