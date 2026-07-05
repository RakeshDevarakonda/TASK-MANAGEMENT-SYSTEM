package middleware

import (
    "math"
    "net/http"
    "sync"
    "time"

    "task-system/internal/utils"

    "github.com/gin-gonic/gin"
)

type bucket struct {
    tokens     float64
    lastRefill time.Time
}

var (
    mu      sync.Mutex
    buckets = make(map[string]*bucket)
)

// RateLimit implements an in‑memory token bucket rate limiter.
// Unauthenticated users (identified by IP) get 5 tokens, refill 1 token per minute.
// Authenticated users (identified by JWT user ID) get 10 tokens, refill 1 token per minute.
func RateLimit() gin.HandlerFunc {
    return func(c *gin.Context) {
        key := ""
        capacity := 0.0
        // Try to get a valid JWT token to identify an authenticated user.
        if tokenStr, err := c.Cookie("accessToken"); err == nil {
            if claims, err := utils.ValidateAccessToken(tokenStr); err == nil && claims.UserID != "" {
                key = "user:" + claims.UserID
                capacity = 10
            }
        }
        if key == "" {
            // Fallback to IP‑based key for unauthenticated requests.
            key = "ip:" + c.ClientIP()
            capacity = 5
        }
        now := time.Now()
        mu.Lock()
        b, ok := buckets[key]
        if !ok {
            b = &bucket{tokens: capacity, lastRefill: now}
            buckets[key] = b
        }
        // Refill tokens based on whole minutes passed.
        minutes := int(now.Sub(b.lastRefill).Minutes())
        if minutes > 0 {
            b.tokens = math.Min(b.tokens+float64(minutes), capacity)
            b.lastRefill = b.lastRefill.Add(time.Duration(minutes) * time.Minute)
        }
        if b.tokens < 1 {
            mu.Unlock()
            utils.SendError(c, http.StatusTooManyRequests, "rate limit exceeded", nil)
            c.Abort()
            return
        }
        b.tokens -= 1
        mu.Unlock()
        c.Next()
    }
}
