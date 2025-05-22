package service

import (
    "database/sql"
    "testing"
    "time"

    _ "github.com/mattn/go-sqlite3"
)

func TestCreateAndResolveLink(t *testing.T) {
    // In-memory DB
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("open DB: %v", err)
    }
    defer db.Close()

    // Create table
    _, err = db.Exec(`
        CREATE TABLE links (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            shortcode TEXT UNIQUE NOT NULL,
            target_url TEXT NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            hits INTEGER DEFAULT 0,
            expires_at TIMESTAMP NULL,
            revoked BOOLEAN DEFAULT 0
        )
    `)
    if err != nil {
        t.Fatalf("create table: %v", err)
    }

    // 1) Create link without expiry
    link, err := CreateLink(db, "https://example.com", nil)
    if err != nil {
        t.Fatalf("CreateLink: %v", err)
    }
    if link.Shortcode == "" {
        t.Fatal("empty shortcode generated")
    }

    // Resolve it
    resolved, err := ResolveLink(db, link.Shortcode)
    if err != nil {
        t.Fatalf("ResolveLink: %v", err)
    }
    if resolved.Shortcode != link.Shortcode || resolved.TargetURL != link.TargetURL {
        t.Fatal("resolved link doesn't match original")
    }

    // 2) Create with expiry in the past
    past := time.Now().Add(-1 * time.Hour)
    expiredLink, err := CreateLink(db, "https://expired.com", &past)
    if err != nil {
        t.Fatalf("CreateLink (expired): %v", err)
    }

    // Attempt to resolve expired
    _, err = ResolveLink(db, expiredLink.Shortcode)
    if err == nil {
        t.Fatal("expected error resolving expired link")
    }
}
