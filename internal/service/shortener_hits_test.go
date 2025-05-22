package service

import (
    "database/sql"
    "testing"

    _ "github.com/mattn/go-sqlite3"
)

func TestIncrementHits(t *testing.T) {
    // In-memory DB & table
    db, _ := sql.Open("sqlite3", ":memory:")
    defer db.Close()
    db.Exec(`
        CREATE TABLE links (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            shortcode TEXT UNIQUE NOT NULL,
            target_url TEXT NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            hits INTEGER DEFAULT 0,
            expires_at TIMESTAMP NULL,
            revoked BOOLEAN DEFAULT 0
        );
    `)

    // Insert a test row
    code := "hitcode"
    db.Exec("INSERT INTO links (shortcode, target_url) VALUES (?, ?)", code, "https://x")

    // Increment once
    if err := IncrementHits(db, code); err != nil {
        t.Fatalf("first increment: %v", err)
    }
    // Verify
    var hits int
    db.QueryRow("SELECT hits FROM links WHERE shortcode = ?", code).Scan(&hits)
    if hits != 1 {
        t.Fatalf("expected 1 hit, got %d", hits)
    }

    // Increment again
    IncrementHits(db, code)
    db.QueryRow("SELECT hits FROM links WHERE shortcode = ?", code).Scan(&hits)
    if hits != 2 {
        t.Fatalf("expected 2 hits, got %d", hits)
    }

    // Try incrementing nonexistent code
    if err := IncrementHits(db, "nope"); err == nil {
        t.Fatal("expected error for nonexistent code")
    }
}
