package api

import (
    "github.com/valorm/snapurl/internal/config"
    "log"
    "net/http"
)

// LoggingMiddleware logs request and response details
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("Started %s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
        log.Printf("Completed %s %s", r.Method, r.URL.Path)
    })
}

// AuthMiddleware validates API keys for admin actions
func AuthMiddleware(cfg *config.Config, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        apiKey := r.Header.Get("X-API-Key")
        if !contains(cfg.APIKeys, apiKey) {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
