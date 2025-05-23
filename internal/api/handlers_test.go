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
    _ "github.com/mattn/go-sqlite3"
)

func TestShortenHandler(t *testing.T) {
    // 1) In-memory DB and migrations
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("open DB: %v", err)
    }
    defer db.Close()
    if err := datastore.RunMigrations(db); err != nil {
        t.Fatalf("migrations: %v", err)
    }

    // 2) Create handler
    handler := ShortenHandler(db)

    // 3) Prepare request
    body := map[string]string{"url": "https://example.com"}
    buf, _ := json.Marshal(body)
    req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(buf))
    req.Header.Set("Content-Type", "application/json")
    rr := httptest.NewRecorder()

    // 4) Call handler
    handler.ServeHTTP(rr, req)

    // 5) Check response code
    if rr.Code != http.StatusCreated {
        t.Fatalf("want 201; got %d", rr.Code)
    }

    // 6) Parse response
    var resp map[string]string
    if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
        t.Fatalf("invalid JSON response: %v", err)
    }
    shortcode, ok := resp["shortcode"]
    if !ok || shortcode == "" {
        t.Fatal("response missing shortcode")
    }

    // 7) Verify DB record
    var link models.Link
    err = db.QueryRow(
        "SELECT id, shortcode, target_url, created_at, hits, expires_at, revoked FROM links WHERE shortcode = ?",
        shortcode,
    ).Scan(
        &link.ID,
        &link.Shortcode,
        &link.TargetURL,
        &link.CreatedAt,
        &link.Hits,
        &link.ExpiresAt,
        &link.Revoked,
    )
    if err != nil {
        t.Fatalf("db query failed: %v", err)
    }
    if link.TargetURL != "https://example.com" {
        t.Errorf("stored URL = %q; want %q", link.TargetURL, "https://example.com")
    }
}
