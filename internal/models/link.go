package models

import (
    "database/sql"
    "time"
)

type Link struct {
    ID         int
    Shortcode  string
    TargetURL  string
    CreatedAt  time.Time
    Hits       int
    ExpiresAt  sql.NullTime
    Revoked    bool
}
