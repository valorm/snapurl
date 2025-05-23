package limiter

import (
    "net/http"
    "sync"
    "time"

    "golang.org/x/time/rate"
)

// IPRateLimiter manages rate limits per IP
type IPRateLimiter struct {
    limiterMap map[string]*rate.Limiter
    mu         sync.RWMutex
    rps        int
}

// NewIPRateLimiter initializes the limiter
func NewIPRateLimiter(rps int) *IPRateLimiter {
    return &IPRateLimiter{
        limiterMap: make(map[string]*rate.Limiter),
        rps:        rps,
    }
}

// GetLimiter returns or creates a rate limiter for the IP
func (l *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
    l.mu.Lock()
    defer l.mu.Unlock()

    if limiter, exists := l.limiterMap[ip]; exists {
        return limiter
    }

    limiter := rate.NewLimiter(rate.Every(time.Second/time.Duration(l.rps)), l.rps)
    l.limiterMap[ip] = limiter
    return limiter
}

// Middleware wraps a handler to enforce rate limits
func (l *IPRateLimiter) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip := r.RemoteAddr // Simplified: for production, parse headers or use real IP resolver

        limiter := l.GetLimiter(ip)
        if !limiter.Allow() {
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }

        next.ServeHTTP(w, r)
    })
}
