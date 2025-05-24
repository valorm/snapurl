package main

import (
    "log"
    "net/http"

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

    // Apply middleware: recovery → logging → rate limiting
    handler := rateLimiter.Middleware(
        api.RecoveryMiddleware(
            api.LoggingMiddleware(mux),
        ),
    )

    log.Printf("Starting server on %s", cfg.Port)
    if err := http.ListenAndServe(cfg.Port, handler); err != nil {
        log.Fatal(err)
    }
}