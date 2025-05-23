package limiter

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestRateLimiting(t *testing.T) {
    limiter := NewIPRateLimiter(2) // 2 RPS
    dummy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })
    handler := limiter.Middleware(dummy)

    req, _ := http.NewRequest("GET", "/", nil)
    req.RemoteAddr = "192.168.1.1"
    rr := httptest.NewRecorder()

    // First 2 requests succeed
    for i := 0; i < 2; i++ {
        handler.ServeHTTP(rr, req)
        if rr.Code != http.StatusOK {
            t.Errorf("Expected 200, got %d", rr.Code)
        }
        rr = httptest.NewRecorder()
    }

    // Third should fail
    handler.ServeHTTP(rr, req)
    if rr.Code != http.StatusTooManyRequests {
        t.Errorf("Expected 429, got %d", rr.Code)
    }
}
