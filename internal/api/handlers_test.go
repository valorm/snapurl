package api

import (
    "bytes"
    "database/sql"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/valorm/snapurl/internal/datastore"
    "github.com/valorm/snapurl/internal/models"
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
    if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
        t.Fatalf("invalid JSON response: %v", err)
    }
    shortcode := resp["shortcode"]
    if shortcode == "" {
        t.Fatal("missing shortcode")
    }

    var link models.Link
    err := db.QueryRow(
        "SELECT id, shortcode, target_url, created_at, hits, expires_at, revoked FROM links WHERE shortcode = ?",
        shortcode,
    ).Scan(&link.ID, &link.Shortcode, &link.TargetURL, &link.CreatedAt, &link.Hits, &link.ExpiresAt, &link.Revoked)
    if err != nil {
        t.Fatalf("db query failed: %v", err)
    }
    if link.TargetURL != "https://example.com" {
        t.Errorf("stored URL = %q; want %q", link.TargetURL, "https://example.com")
    }
}

func TestRedirectHandler(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    link, err := service.CreateLink(db, "https://example.com", nil)
    if err != nil {
        t.Fatalf("CreateLink: %v", err)
    }

    req := httptest.NewRequest(http.MethodGet, "/"+link.Shortcode, nil)
    rr := httptest.NewRecorder()

    handler := RedirectHandler(db)
    handler.ServeHTTP(rr, req)

    if rr.Code != http.StatusFound {
        t.Errorf("Expected 302 Found, got %d", rr.Code)
    }

    if loc := rr.Header().Get("Location"); loc != link.TargetURL {
        t.Errorf("Expected redirect to %q, got %q", link.TargetURL, loc)
    }

    var hits int
    err = db.QueryRow("SELECT hits FROM links WHERE shortcode = ?", link.Shortcode).Scan(&hits)
    if err != nil {
        t.Fatalf("DB query error: %v", err)
    }
    if hits != 1 {
        t.Errorf("Expected hits to be 1, got %d", hits)
    }
}

func TestRevokeHandlerAndRevokedRedirect(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    link, err := service.CreateLink(db, "https://revokable.com", nil)
    if err != nil {
        t.Fatalf("CreateLink: %v", err)
    }

    err = service.RevokeLink(db, link.Shortcode)
    if err != nil {
        t.Fatalf("RevokeLink: %v", err)
    }

    req := httptest.NewRequest(http.MethodGet, "/"+link.Shortcode, nil)
    rr := httptest.NewRecorder()
    RedirectHandler(db).ServeHTTP(rr, req)

    if rr.Code != http.StatusGone {
        t.Errorf("Expected 410 Gone, got %d", rr.Code)
    }
}

func TestRevokeHandlerAPI(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    link, err := service.CreateLink(db, "https://torevoke.com", nil)
    if err != nil {
        t.Fatalf("CreateLink: %v", err)
    }

    req := httptest.NewRequest(http.MethodDelete, "/"+link.Shortcode, nil)
    req.Header.Set("X-API-Key", "default_key_1")
    rr := httptest.NewRecorder()

    handler := RevokeHandler(db, []string{"default_key_1"})
    handler.ServeHTTP(rr, req)

    if rr.Code != http.StatusNoContent {
        t.Errorf("Expected 204 No Content, got %d", rr.Code)
    }

    var revoked bool
    err = db.QueryRow("SELECT revoked FROM links WHERE shortcode = ?", link.Shortcode).Scan(&revoked)
    if err != nil {
        t.Fatalf("DB query failed: %v", err)
    }
    if !revoked {
        t.Error("Expected link to be marked as revoked")
    }
}
