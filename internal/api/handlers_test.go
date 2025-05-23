package api

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"

    "github.com/valorm/snapurl/internal/config"
    "github.com/valorm/snapurl/internal/datastore"
    _ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("open DB: %v", err)
    }
    if err := datastore.RunMigrations(db); err != nil {
        t.Fatalf("migrations: %v", err)
    }
    return db
}

func TestFullWorkflow(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    // Fake config for auth
    cfg := &config.Config{
        Port:      ":8080",
        DBPath:    ":memory:",
        RateLimit: 10,
        APIKeys:   []string{"test-key"},
    }

    // 1) Create
    createBody := `{"url":"https://example.com","expiry":"` +
        time.Now().Add(time.Hour).UTC().Format(time.RFC3339) + `"}`
    req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(createBody))
    req.Header.Set("Content-Type", "application/json")
    rr := httptest.NewRecorder()
    ShortenHandler(db).ServeHTTP(rr, req)
    if rr.Code != http.StatusCreated {
        t.Fatalf("Create: want 201, got %d", rr.Code)
    }
    var crResp struct{ Shortcode string }
    if err := json.NewDecoder(rr.Body).Decode(&crResp); err != nil {
        t.Fatalf("Create decode: %v", err)
    }
    code := crResp.Shortcode

    // 2) Redirect
    rr = httptest.NewRecorder()
    RedirectHandler(db).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/"+code, nil))
    if rr.Code != http.StatusFound {
        t.Errorf("Redirect: want 302, got %d", rr.Code)
    }
    if loc := rr.Header().Get("Location"); loc != "https://example.com" {
        t.Errorf("Redirect Location: want https://example.com, got %q", loc)
    }

    // 3) Revoke
    req = httptest.NewRequest(http.MethodDelete, "/"+code, nil)
    req.Header.Set("X-API-Key", "test-key")
    rr = httptest.NewRecorder()
    RevokeHandler(db, cfg.APIKeys).ServeHTTP(rr, req)
    if rr.Code != http.StatusNoContent {
        t.Errorf("Revoke: want 204, got %d", rr.Code)
    }

    // 4) Access after revoke
    rr = httptest.NewRecorder()
    RedirectHandler(db).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/"+code, nil))
    if rr.Code != http.StatusGone {
        t.Errorf("Post-revoke: want 410, got %d", rr.Code)
    }

    // 5) Metrics
    rr = httptest.NewRecorder()
    MetricsHandler(db).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/metrics", nil))
    if rr.Code != http.StatusOK {
        t.Errorf("Metrics: want 200, got %d", rr.Code)
    }
    var m map[string]uint64
    if err := json.NewDecoder(rr.Body).Decode(&m); err != nil {
        t.Fatalf("Metrics decode: %v", err)
    }
    if m["urls_created"] != 1 {
        t.Errorf("urls_created: want 1, got %d", m["urls_created"])
    }
    if m["redirects_served"] != 1 {
        t.Errorf("redirects_served: want 1, got %d", m["redirects_served"])
    }

    // 6) Health
    rr = httptest.NewRecorder()
    HealthHandler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/health", nil))
    if rr.Code != http.StatusOK {
        t.Errorf("Health: want 200, got %d", rr.Code)
    }
    var h map[string]string
    if err := json.NewDecoder(rr.Body).Decode(&h); err != nil {
        t.Fatalf("Health decode: %v", err)
    }
    if h["status"] != "ok" {
        t.Errorf("Health status: want ok, got %q", h["status"])
    }
}
