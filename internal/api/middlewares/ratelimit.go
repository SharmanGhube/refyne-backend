package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimiter manages rate limiting using Redis
type RateLimiter struct {
	redis  *redis.Client
	logger *zap.Logger
}

// RateLimitConfig defines rate limiting rules
type RateLimitConfig struct {
	Requests int           // Maximum requests
	Window   time.Duration // Time window
	KeyType  string        // "ip", "user", or "ip+user"
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter(redisClient *redis.Client, logger *zap.Logger) *RateLimiter {
	return &RateLimiter{
		redis:  redisClient,
		logger: logger,
	}
}

// Middleware creates a rate limiting middleware with the given config
func (rl *RateLimiter) Middleware(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate rate limit key based on config
		key := rl.generateKey(c, config.KeyType)
		
		// Check rate limit
		allowed, remaining, resetTime, err := rl.checkLimit(c.Request.Context(), key, config)
		
		// If Redis is down, allow the request (graceful degradation)
		if err != nil {
			rl.logger.Error("Rate limit check failed, allowing request",
				zap.Error(err),
				zap.String("key", key),
			)
			c.Next()
			return
		}
		
		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.Requests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))
		
		if !allowed {
			rl.logger.Warn("Rate limit exceeded",
				zap.String("key", key),
				zap.String("ip", c.ClientIP()),
				zap.String("path", c.Request.URL.Path),
			)
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "RATE_LIMIT_EXCEEDED",
				"message": "Too many requests. Please try again later.",
				"retry_after": resetTime.Unix(),
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// generateKey creates a rate limit key based on the type
func (rl *RateLimiter) generateKey(c *gin.Context, keyType string) string {
	route := c.FullPath()
	if route == "" {
		route = c.Request.URL.Path
	}
	
	switch keyType {
	case "user":
		// Extract user ID from context (set by auth middleware)
		if userID, exists := c.Get("userID"); exists {
			return fmt.Sprintf("ratelimit:user:%s:%s", userID, route)
		}
		// Fallback to IP if no user ID
		return fmt.Sprintf("ratelimit:ip:%s:%s", c.ClientIP(), route)
	
	case "ip+user":
		if userID, exists := c.Get("userID"); exists {
			return fmt.Sprintf("ratelimit:combined:%s:%s:%s", c.ClientIP(), userID, route)
		}
		return fmt.Sprintf("ratelimit:ip:%s:%s", c.ClientIP(), route)
	
	default: // "ip"
		return fmt.Sprintf("ratelimit:ip:%s:%s", c.ClientIP(), route)
	}
}

// checkLimit checks if the request is within rate limits
func (rl *RateLimiter) checkLimit(ctx context.Context, key string, config RateLimitConfig) (bool, int, time.Time, error) {
	now := time.Now()
	windowStart := now.Add(-config.Window)
	
	// Use Redis sorted set to track requests within the time window
	pipe := rl.redis.Pipeline()
	
	// Remove old entries outside the time window
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.UnixNano()))
	
	// Count current requests in window
	zCountCmd := pipe.ZCount(ctx, key, fmt.Sprintf("%d", windowStart.UnixNano()), fmt.Sprintf("%d", now.UnixNano()))
	
	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.UnixNano()),
		Member: fmt.Sprintf("%d", now.UnixNano()),
	})
	
	// Set expiration on the key
	pipe.Expire(ctx, key, config.Window+time.Minute)
	
	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, time.Time{}, err
	}
	
	// Get the count
	count, err := zCountCmd.Result()
	if err != nil {
		return false, 0, time.Time{}, err
	}
	
	// Calculate remaining requests and reset time
	remaining := config.Requests - int(count)
	if remaining < 0 {
		remaining = 0
	}
	
	resetTime := now.Add(config.Window)
	allowed := count <= int64(config.Requests)
	
	return allowed, remaining, resetTime, nil
}

// GetRateLimitInfo returns current rate limit status for a key
func (rl *RateLimiter) GetRateLimitInfo(ctx context.Context, key string, config RateLimitConfig) (int, time.Time, error) {
	now := time.Now()
	windowStart := now.Add(-config.Window)
	
	// Count current requests in window
	count, err := rl.redis.ZCount(ctx, key, fmt.Sprintf("%d", windowStart.UnixNano()), fmt.Sprintf("%d", now.UnixNano())).Result()
	if err != nil {
		return 0, time.Time{}, err
	}
	
	remaining := config.Requests - int(count)
	if remaining < 0 {
		remaining = 0
	}
	
	resetTime := now.Add(config.Window)
	
	return remaining, resetTime, nil
}

// ResetLimit clears rate limit for a specific key (admin function)
func (rl *RateLimiter) ResetLimit(ctx context.Context, key string) error {
	return rl.redis.Del(ctx, key).Err()
}

// Predefined rate limit configurations
var (
	// Auth endpoints
	RegisterRateLimit = RateLimitConfig{
		Requests: 3,
		Window:   time.Hour,
		KeyType:  "ip",
	}
	
	LoginRateLimit = RateLimitConfig{
		Requests: 10,
		Window:   time.Hour,
		KeyType:  "ip",
	}
	
	OTPRequestRateLimit = RateLimitConfig{
		Requests: 5,
		Window:   15 * time.Minute,
		KeyType:  "ip",
	}
	
	PasswordResetRequestRateLimit = RateLimitConfig{
		Requests: 3,
		Window:   time.Hour,
		KeyType:  "ip",
	}
	
	VerificationResendRateLimit = RateLimitConfig{
		Requests: 5,
		Window:   time.Hour,
		KeyType:  "ip",
	}
	
	// Protected endpoints (requires auth)
	ProtectedEndpointRateLimit = RateLimitConfig{
		Requests: 100,
		Window:   time.Minute,
		KeyType:  "user",
	}
	
	// General API rate limit
	GeneralAPIRateLimit = RateLimitConfig{
		Requests: 60,
		Window:   time.Minute,
		KeyType:  "ip",
	}
)
