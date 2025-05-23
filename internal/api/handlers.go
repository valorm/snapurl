package api

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "strings"
)

// ShortenHandler handles POST /shorten
func ShortenHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // TODO: Parse JSON body {url, expiry}
        // TODO: Call service.CreateLink()
        // TODO: Return {shortcode} in JSON 201

        w.WriteHeader(http.StatusNotImplemented)
        w.Write([]byte(`{"error":"Not implemented"}`))
    }
}

// RedirectHandler handles GET /{shortcode}
func RedirectHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Extract shortcode from path (assumes "/{shortcode}" pattern)
        shortcode := strings.TrimPrefix(r.URL.Path, "/")
        if shortcode == "" {
            http.NotFound(w, r)
            return
        }

        // TODO: Call service.ResolveLink()
        // TODO: 302 redirect or 404/410 if invalid

        w.WriteHeader(http.StatusNotImplemented)
        w.Write([]byte(`{"error":"Not implemented"}`))
    }
}

// RevokeHandler handles DELETE /{shortcode}
func RevokeHandler(db *sql.DB, apiKeys []string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // TODO: Validate API key from header
        // TODO: Call service.RevokeLink()
        // TODO: Return 204 No Content

        w.WriteHeader(http.StatusNotImplemented)
        w.Write([]byte(`{"error":"Not implemented"}`))
    }
}

// HealthHandler handles GET /health
func HealthHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    }
}

// MetricsHandler handles GET /metrics
func MetricsHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // TODO: Return JSON with counters (urls_created, redirects_served, active_links)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusNotImplemented)
        w.Write([]byte(`{"error":"Not implemented"}`))
    }
}
