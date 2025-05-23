package api

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "strings"
    "time"

    "github.com/valorm/snapurl/internal/service"
)

// ShortenHandler handles POST /shorten
func ShortenHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req struct {
            URL    string    `json:"url"`
            Expiry time.Time `json:"expiry,omitempty"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid request body", http.StatusBadRequest)
            return
        }

        // Trim spaces from the URL input
        req.URL = strings.TrimSpace(req.URL)

        if req.URL == "" {
            http.Error(w, "Missing 'url' field", http.StatusBadRequest)
            return
        }

        var expiry *time.Time
        if !req.Expiry.IsZero() {
            expiry = &req.Expiry
        }
        link, err := service.CreateLink(db, req.URL, expiry)
        if err != nil {
            http.Error(w, "Failed to create link", http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(map[string]string{"shortcode": link.Shortcode})
    }
}

// RedirectHandler handles GET /{shortcode}
func RedirectHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Extract shortcode from path
        code := strings.TrimPrefix(r.URL.Path, "/")
        if code == "" {
            http.NotFound(w, r)
            return
        }
        // Resolve link
        link, err := service.ResolveLink(db, code)
        if err != nil {
            http.NotFound(w, r)
            return
        }
        // Increment hits
        if err := service.IncrementHits(db, code); err != nil {
            // Log error but continue
        }
        // Redirect
        http.Redirect(w, r, link.TargetURL, http.StatusFound)
    }
}

// RevokeHandler handles DELETE /{shortcode}
func RevokeHandler(db *sql.DB, apiKeys []string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // TODO: validate API key header, call service.RevokeLink
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
        // TODO: return JSON or Prometheus metrics
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusNotImplemented)
        w.Write([]byte(`{"error":"Not implemented"}`))
    }
}
