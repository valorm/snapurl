package main

import (
    "log"
    "net/http"

    "github.com/valorm/snapurl/internal/config"
    "github.com/valorm/snapurl/internal/datastore"
    "github.com/valorm/snapurl/internal/api"
    "github.com/valorm/snapurl/internal/limiter"
)

func main() {
    // 1) Load configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

    // 2) Open and migrate the database
    db, err := datastore.OpenDB(cfg.DBPath)
    if err != nil {
        log.Fatalf("failed to open database: %v", err)
    }
    defer db.Close()

    // 3) Setup Rate Limiter (e.g. 2 requests/sec per IP)
    rateLimiter := limiter.NewIPRateLimiter(2)

    // 4) Setup HTTP mux and routes
    mux := http.NewServeMux()
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })
    mux.HandleFunc("/shorten", api.ShortenHandler)        // POST
    mux.HandleFunc("/", api.RedirectHandler)              // GET /{shortcode}

    // 5) Start server with rate-limiting middleware
    log.Printf("starting server on %s", cfg.Port)
    if err := http.ListenAndServe(cfg.Port, rateLimiter.Middleware(mux)); err != nil {
        log.Fatalf("server error: %v", err)
    }
}
