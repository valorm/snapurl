package telemetry

import (
    "database/sql"
    "sync/atomic"
)

// Atomic counters
var (
    urlsCreated     uint64
    redirectsServed uint64
)

// Increment increases the named counter
func Increment(name string) {
    switch name {
    case "urls_created":
        atomic.AddUint64(&urlsCreated, 1)
    case "redirects_served":
        atomic.AddUint64(&redirectsServed, 1)
    }
}

// GetMetrics returns all metrics, including active_links from the database
func GetMetrics(db *sql.DB) (map[string]uint64, error) {
    active, err := getActiveLinksCount(db)
    if err != nil {
        return nil, err
    }
    return map[string]uint64{
        "urls_created":     atomic.LoadUint64(&urlsCreated),
        "redirects_served": atomic.LoadUint64(&redirectsServed),
        "active_links":     active,
    }, nil
}

// getActiveLinksCount counts non-expired, non-revoked links
func getActiveLinksCount(db *sql.DB) (uint64, error) {
    var count uint64
    row := db.QueryRow(`
        SELECT COUNT(*) 
        FROM links 
        WHERE (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
          AND revoked = 0
    `)
    if err := row.Scan(&count); err != nil {
        return 0, err
    }
    return count, nil
}
