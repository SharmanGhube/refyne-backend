package middlewares

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// InMemoryRateLimiter implements a simple sliding window rate limiter using in-memory storage
// Note: This is suitable for single-instance deployments. For production scaling across
// multiple servers, use Redis-backed rate limiting (see ratelimit.go)
type InMemoryRateLimiter struct {
	mu      sync.RWMutex
	buckets map[string]*rateBucket
	logger  *zap.Logger
}

type rateBucket struct {
	requests []time.Time
	mu       sync.Mutex
}

type RateLimitRule struct {
	Requests int           // Maximum requests allowed
	Window   time.Duration // Time window
	KeyFunc  KeyGenFunc    // Function to generate rate limit key
}

type KeyGenFunc func(*gin.Context) string

// NewInMemoryRateLimiter creates a new in-memory rate limiter
func NewInMemoryRateLimiter(logger *zap.Logger) *InMemoryRateLimiter {
	rl := &InMemoryRateLimiter{
		buckets: make(map[string]*rateBucket),
		logger:  logger,
	}

	// Start cleanup routine
	go rl.cleanup()

	return rl
}

// Middleware creates a rate limiting middleware with the given rule
func (rl *InMemoryRateLimiter) Middleware(rule RateLimitRule) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := rule.KeyFunc(c)

		allowed, remaining, resetTime := rl.checkLimit(key, rule)

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rule.Requests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

		if !allowed {
			rl.logger.Warn("Rate limit exceeded",
				zap.String("key", key),
				zap.String("ip", c.ClientIP()),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
			)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "RATE_LIMIT_EXCEEDED",
				"message":     "Too many requests. Please try again later.",
				"retry_after": resetTime.Unix(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkLimit checks if the request is within rate limits
func (rl *InMemoryRateLimiter) checkLimit(key string, rule RateLimitRule) (bool, int, time.Time) {
	rl.mu.Lock()
	bucket, exists := rl.buckets[key]
	if !exists {
		bucket = &rateBucket{
			requests: make([]time.Time, 0),
		}
		rl.buckets[key] = bucket
	}
	rl.mu.Unlock()

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rule.Window)

	// Remove old requests outside the time window
	validRequests := make([]time.Time, 0)
	for _, reqTime := range bucket.requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}
	bucket.requests = validRequests

	// Check if within limit
	count := len(bucket.requests)
	allowed := count < rule.Requests

	if allowed {
		bucket.requests = append(bucket.requests, now)
		count++
	}

	remaining := rule.Requests - count
	if remaining < 0 {
		remaining = 0
	}

	resetTime := now.Add(rule.Window)
	if len(bucket.requests) > 0 {
		// Reset time is when the oldest request in the window expires
		oldestRequest := bucket.requests[0]
		resetTime = oldestRequest.Add(rule.Window)
	}

	return allowed, remaining, resetTime
}

// cleanup periodically removes stale buckets
func (rl *InMemoryRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, bucket := range rl.buckets {
			bucket.mu.Lock()
			// If no requests in the last hour, remove the bucket
			if len(bucket.requests) == 0 || now.Sub(bucket.requests[len(bucket.requests)-1]) > time.Hour {
				delete(rl.buckets, key)
			}
			bucket.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// Key generation functions
func IPKey(c *gin.Context) string {
	route := c.FullPath()
	if route == "" {
		route = c.Request.URL.Path
	}
	return fmt.Sprintf("ip:%s:%s", c.ClientIP(), route)
}

func UserKey(c *gin.Context) string {
	route := c.FullPath()
	if route == "" {
		route = c.Request.URL.Path
	}

	// Try to get user ID from context (set by auth middleware)
	if userID, exists := c.Get("userID"); exists {
		return fmt.Sprintf("user:%v:%s", userID, route)
	}

	// Fallback to IP if no user
	return IPKey(c)
}

func EmailKey(field string) KeyGenFunc {
	return func(c *gin.Context) string {
		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		// Try to extract email from JSON body
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err == nil {
			if email, ok := body[field].(string); ok {
				// Rebind the body for the next handler
				c.Set("_ratelimit_body", body)
				return fmt.Sprintf("email:%s:%s", email, route)
			}
		}

		// Fallback to IP
		return IPKey(c)
	}
}

// Predefined rate limit rules
var (
	RegisterLimit = RateLimitRule{
		Requests: 3,
		Window:   time.Hour,
		KeyFunc:  IPKey,
	}

	LoginLimit = RateLimitRule{
		Requests: 10,
		Window:   time.Hour,
		KeyFunc:  IPKey,
	}

	RefreshLimit = RateLimitRule{
		Requests: 20,
		Window:   time.Hour,
		KeyFunc:  IPKey,
	}

	OTPRequestLimit = RateLimitRule{
		Requests: 5,
		Window:   15 * time.Minute,
		KeyFunc:  IPKey,
	}

	PasswordResetLimit = RateLimitRule{
		Requests: 3,
		Window:   time.Hour,
		KeyFunc:  IPKey,
	}

	VerificationResendLimit = RateLimitRule{
		Requests: 5,
		Window:   time.Hour,
		KeyFunc:  IPKey,
	}

	ProtectedEndpointLimit = RateLimitRule{
		Requests: 100,
		Window:   time.Minute,
		KeyFunc:  UserKey,
	}

	GeneralAPILimit = RateLimitRule{
		Requests: 60,
		Window:   time.Minute,
		KeyFunc:  IPKey,
	}
)
