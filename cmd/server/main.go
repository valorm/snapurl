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
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatal(err)
    }

    db, err := datastore.OpenDB(cfg.DBPath)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    rateLimiter := limiter.NewIPRateLimiter(cfg.RateLimit)
    telemetry.Init()

    mux := http.NewServeMux()

    // Public: POST /shorten
    mux.Handle("/shorten", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodPost {
            api.ShortenHandler(db).ServeHTTP(w, r)
            return
        }
        http.NotFound(w, r)
    }))

    // Shared: GET for redirect and DELETE for revocation on /
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            api.RedirectHandler(db).ServeHTTP(w, r)
        case http.MethodDelete:
            api.AuthMiddleware(cfg, api.RevokeHandler(db, cfg.APIKeys)).ServeHTTP(w, r)
        default:
            http.NotFound(w, r)
        }
    })

    // Public: health and metrics
    mux.Handle("/health", api.HealthHandler())
    mux.Handle("/metrics", api.MetricsHandler())

    // Apply rate limiting and logging globally
    handler := rateLimiter.Middleware(api.LoggingMiddleware(mux))

    log.Printf("Starting server on %s", cfg.Port)
    if err := http.ListenAndServe(cfg.Port, handler); err != nil {
        log.Fatal(err)
    }
}
