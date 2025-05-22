package util

import (
    "crypto/rand"
    "fmt"
)

const (
    base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
    base62Size  = 62
)

// GenerateCode creates a secure, URL-safe base62 code of specified length.
func GenerateCode(length int) (string, error) {
    if length <= 0 {
        return "", fmt.Errorf("length must be > 0")
    }

    // Calculate how many random bytes needed (6 bits per character)
    bits := length * 6
    bytesNeeded := (bits + 7) / 8 // Round up to full bytes

    // Generate random bytes
    b := make([]byte, bytesNeeded)
    if _, err := rand.Read(b); err != nil {
        return "", fmt.Errorf("failed to read random bytes: %w", err)
    }

    var result []byte
    for i := 0; len(result) < length; i++ {
        // Cycle through bytes, mask low 6 bits
        idx := int(b[i%len(b)] & 0x3F) // 0–63
        if idx < base62Size {
            result = append(result, base62Chars[idx])
        }
        // If idx ≥ 62, skip to avoid bias
    }

    return string(result), nil
}
