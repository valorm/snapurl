package main

import (
    "fmt"
    "log"

    "github.com/valorm/snapurl/internal/config"
    "github.com/valorm/snapurl/internal/datastore"
    "github.com/valorm/snapurl/internal/models"
)

func main() {
    // 1) Load config to get DB path
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("load config: %v", err)
    }

    // 2) Open DB (runs migrations)
    db, err := datastore.OpenDB(cfg.DBPath)
    if err != nil {
        log.Fatalf("open db: %v", err)
    }
    defer db.Close()

    // 3) Clear any existing test record
    db.Exec("DELETE FROM links WHERE shortcode = ?", "abc123")

    // 4) Insert a test link
    _, err = db.Exec(
        "INSERT INTO links (shortcode, target_url) VALUES (?, ?)",
        "abc123", "https://example.com",
    )
    if err != nil {
        log.Fatalf("insert link: %v", err)
    }

    // 5) Query it back into models.Link
    var link models.Link
    row := db.QueryRow(
        "SELECT id, shortcode, target_url, created_at, hits, expires_at, revoked FROM links WHERE shortcode = ?",
        "abc123",
    )
    err = row.Scan(
        &link.ID,
        &link.Shortcode,
        &link.TargetURL,
        &link.CreatedAt,
        &link.Hits,
        &link.ExpiresAt,
        &link.Revoked,
    )
    if err != nil {
        log.Fatalf("scan link: %v", err)
    }

    // 6) Validate and print result
    if link.Shortcode != "abc123" || link.TargetURL != "https://example.com" {
        log.Fatalf("model mismatch: got %+v", link)
    }

    fmt.Printf("✅ Link model works: %+v\n", link)
}
