-- Adds expiry and revocation columns
ALTER TABLE links
ADD COLUMN expires_at TIMESTAMP NULL;

ALTER TABLE links
ADD COLUMN revoked BOOLEAN DEFAULT 0;
