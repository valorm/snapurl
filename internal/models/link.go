package models

import (
    "database/sql"
    "time"
)

type Link struct {
    ID        int            `json:"id"`
    Shortcode string         `json:"shortcode"`
    TargetURL string         `json:"target_url"`
    CreatedAt time.Time      `json:"created_at"`
    Hits      int            `json:"hits"`
    ExpiresAt sql.NullTime   `json:"expires_at"`
    Revoked   bool           `json:"revoked"`
}
