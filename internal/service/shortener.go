package service

import (
    "database/sql"
    "fmt"
    "time"

    "github.com/valorm/snapurl/internal/models"
    "github.com/valorm/snapurl/pkg/util"
)

// CreateLink generates a unique code and persists a new link
func CreateLink(db *sql.DB, targetURL string, expiry *time.Time) (models.Link, error) {
    var link models.Link
    var err error
    var unique bool

    // Generate unique code (retry on collision)
    for i := 0; i < 10; i++ {
        code, err := util.GenerateCode(8)
        if err != nil {
            return models.Link{}, fmt.Errorf("generate code: %w", err)
        }

        // Check uniqueness
        var dummy int
        err = db.QueryRow("SELECT 1 FROM links WHERE shortcode = ?", code).Scan(&dummy)
        if err == sql.ErrNoRows {
            link.Shortcode = code
            unique = true
            break // found a unique code
        }
        if err != nil && err != sql.ErrNoRows {
            return models.Link{}, fmt.Errorf("check code uniqueness: %w", err)
        }
        // else: collided, retry
    }

    if !unique {
        return models.Link{}, fmt.Errorf("failed to generate unique code after 10 attempts")
    }

    // Prepare record
    createdAt := time.Now()
    expiresAt := sql.NullTime{}
    if expiry != nil {
        expiresAt.Time = *expiry
        expiresAt.Valid = true
    }

    // Insert into database
    res, err := db.Exec(
        "INSERT INTO links (shortcode, target_url, created_at, expires_at) VALUES (?, ?, ?, ?)",
        link.Shortcode, targetURL, createdAt, expiresAt,
    )
    if err != nil {
        return models.Link{}, fmt.Errorf("insert link: %w", err)
    }

    id, _ := res.LastInsertId()
    link.ID = int(id)
    link.TargetURL = targetURL
    link.CreatedAt = createdAt
    link.Hits = 0
    link.ExpiresAt = expiresAt
    link.Revoked = false

    return link, nil
}

// ResolveLink retrieves a valid active link by code
func ResolveLink(db *sql.DB, code string) (models.Link, error) {
    var link models.Link
    var expiresAt sql.NullTime

    err := db.QueryRow(
        "SELECT id, target_url, created_at, hits, expires_at, revoked FROM links WHERE shortcode = ?",
        code,
    ).Scan(&link.ID, &link.TargetURL, &link.CreatedAt, &link.Hits, &expiresAt, &link.Revoked)
    if err == sql.ErrNoRows {
        return models.Link{}, fmt.Errorf("link not found")
    }
    if err != nil {
        return models.Link{}, fmt.Errorf("query link: %w", err)
    }

    link.Shortcode = code
    link.ExpiresAt = expiresAt

    // Check expiry
    if expiresAt.Valid && expiresAt.Time.Before(time.Now()) {
        return models.Link{}, fmt.Errorf("link expired")
    }
    if link.Revoked {
        return models.Link{}, fmt.Errorf("link revoked")
    }

    return link, nil
}

// IncrementHits atomically increments the hit counter for a given shortcode.
func IncrementHits(db *sql.DB, code string) error {
    res, err := db.Exec(
        "UPDATE links SET hits = hits + 1 WHERE shortcode = ?",
        code,
    )
    if err != nil {
        return fmt.Errorf("increment hits: %w", err)
    }
    rows, err := res.RowsAffected()
    if err != nil {
        return fmt.Errorf("check rows affected: %w", err)
    }
    if rows == 0 {
        return fmt.Errorf("no link found to increment")
    }
    return nil
}
