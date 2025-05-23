package api

import (
    "bytes"
    "database/sql"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/valorm/snapurl/internal/datastore"
    "github.com/valorm/snapurl/internal/service"
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

func TestShortenHandler(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    handler := ShortenHandler(db)
    body := map[string]string{"url": "https://example.com"}
    buf, _ := json.Marshal(body)
    req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(buf))
    req.Header.Set("Content-Type", "application/json")
    rr := httptest.NewRecorder()
    handler.ServeHTTP(rr, req)

    if rr.Code != http.StatusCreated {
        t.Fatalf("want 201; got %d", rr.Code)
    }
    var resp map[string]string
    _ = json.NewDecoder(rr.Body).Decode(&resp)
    if resp["shortcode"] == "" {
        t.Fatal("missing shortcode")
    }
}

func TestRedirectHandler(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    link, _ := service.CreateLink(db, "https://example.com", nil)
    req := httptest.NewRequest(http.MethodGet, "/"+link.Shortcode, nil)
    rr := httptest.NewRecorder()
    RedirectHandler(db).ServeHTTP(rr, req)

    if rr.Code != http.StatusFound {
        t.Errorf("Expected 302, got %d", rr.Code)
    }
    if loc := rr.Header().Get("Location"); loc != link.TargetURL {
        t.Errorf("Redirect to %q; got %q", link.TargetURL, loc)
    }
}

func TestRevokeHandlerAndRevokedRedirect(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    link, _ := service.CreateLink(db, "https://revokable.com", nil)
    _ = service.RevokeLink(db, link.Shortcode)

    req := httptest.NewRequest(http.MethodGet, "/"+link.Shortcode, nil)
    rr := httptest.NewRecorder()
    RedirectHandler(db).ServeHTTP(rr, req)

    if rr.Code != http.StatusGone {
        t.Errorf("Expected 410 Gone; got %d", rr.Code)
    }
}

func TestRevokeHandlerAPI(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    link, _ := service.CreateLink(db, "https://torevoke.com", nil)
    req := httptest.NewRequest(http.MethodDelete, "/"+link.Shortcode, nil)
    req.Header.Set("X-API-Key", "default_key_1")
    rr := httptest.NewRecorder()
    RevokeHandler(db, []string{"default_key_1"}).ServeHTTP(rr, req)

    if rr.Code != http.StatusNoContent {
        t.Errorf("Expected 204 No Content; got %d", rr.Code)
    }
}

func TestMetricsHandler(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    // create and redirect
    link, _ := service.CreateLink(db, "https://example.com", nil)
    _ = service.IncrementHits(db, link.Shortcode)

    req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
    rr := httptest.NewRecorder()
    MetricsHandler(db).ServeHTTP(rr, req)

    if rr.Code != http.StatusOK {
        t.Fatalf("want 200; got %d", rr.Code)
    }
    var met map[string]uint64
    _ = json.NewDecoder(rr.Body).Decode(&met)
    if met["urls_created"] < 1 {
        t.Errorf("urls_created = %d; want >=1", met["urls_created"])
    }
    if met["redirects_served"] < 1 {
        t.Errorf("redirects_served = %d; want >=1", met["redirects_served"])
    }
}
