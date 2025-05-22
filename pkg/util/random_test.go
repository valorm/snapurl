package util

import (
    "testing"
)

func TestGenerateCode(t *testing.T) {
    const codeLength = 8
    seen := make(map[string]bool)

    for i := 0; i < 100; i++ {
        code, err := GenerateCode(codeLength)
        if err != nil {
            t.Fatalf("Test %d: unexpected error: %v", i, err)
        }

        if len(code) != codeLength {
            t.Errorf("Test %d: expected length %d, got %d", i, codeLength, len(code))
        }

        if seen[code] {
            t.Errorf("Test %d: duplicate code generated: %s", i, code)
        }
        seen[code] = true
    }
}
