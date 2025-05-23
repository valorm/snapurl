package api

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "strings"
    "time"

    "github.com/valorm/snapurl/internal/service"
)

func ShortenHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req struct {
            URL    string    `json:"url"`
            Expiry time.Time `json:"expiry,omitempty"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.URL) == "" {
            http.Error(w, "Invalid request body", http.StatusBadRequest)
            return
        }

        var expiry *time.Time
        if !req.Expiry.IsZero() {
            expiry = &req.Expiry
        }

        link, err := service.CreateLink(db, strings.TrimSpace(req.URL), expiry)
        if err != nil {
            http.Error(w, "Failed to create link", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(map[string]string{"shortcode": link.Shortcode})
    }
}

func RedirectHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        code := strings.TrimPrefix(r.URL.Path, "/")
        if code == "" {
            http.NotFound(w, r)
            return
        }

        link, err := service.ResolveLink(db, code)
        if err != nil {
            http.Error(w, err.Error(), http.StatusGone)
            return
        }

        _ = service.IncrementHits(db, code)
        http.Redirect(w, r, link.TargetURL, http.StatusFound)
    }
}

func RevokeHandler(db *sql.DB, apiKeys []string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        apiKey := r.Header.Get("X-API-Key")
        if !contains(apiKeys, apiKey) {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        code := strings.TrimPrefix(r.URL.Path, "/")
        if code == "" {
            http.Error(w, "Shortcode missing", http.StatusBadRequest)
            return
        }

        if err := service.RevokeLink(db, code); err != nil {
            http.Error(w, "Failed to revoke link", http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusNoContent)
    }
}

func HealthHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    }
}

func MetricsHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusNotImplemented)
        w.Write([]byte(`{"error":"Not implemented"}`))
    }
}
