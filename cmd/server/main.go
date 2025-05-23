package main

import (
 
    "github.com/valorm/snapurl/internal/api"
    "github.com/valorm/snapurl/internal/config"
    "github.com/valorm/snapurl/internal/datastore"
    "github.com/valorm/snapurl/internal/limiter"
    "github.com/valorm/snapurl/internal/telemetry"
    "log"
    "net/http"
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

    // NOTE: ServeMux doesn't support method + path pattern matching,
    // so you might want to handle method inside handlers or use a router lib.
    mux.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodPost {
            api.ShortenHandler(db).ServeHTTP(w, r)
            return
        }
        http.NotFound(w, r)
    })

    // Use wildcard route for redirect and revoke: match all other paths
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            api.RedirectHandler(db).ServeHTTP(w, r)
        case http.MethodDelete:
            api.RevokeHandler(db, cfg.APIKeys).ServeHTTP(w, r)
        default:
            http.NotFound(w, r)
        }
    })

    mux.HandleFunc("/health", api.HealthHandler())
    mux.HandleFunc("/metrics", api.MetricsHandler())

    handler := rateLimiter.Middleware(
        api.LoggingMiddleware(
            api.AuthMiddleware(cfg, mux),
        ),
    )

    log.Printf("Starting server on %s", cfg.Port)
    if err := http.ListenAndServe(cfg.Port, handler); err != nil {
        log.Fatal(err)
    }
}
