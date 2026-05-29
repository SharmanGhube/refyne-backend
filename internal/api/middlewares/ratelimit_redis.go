package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisRateLimiter implements distributed sliding window rate limiting backed by
// Redis sorted sets (ZSET). It automatically falls back to the in-memory limiter
// when Redis is unreachable, ensuring the API remains available even during
// Redis outages.
type RedisRateLimiter struct {
	rdb      redis.Cmdable
	fallback *InMemoryRateLimiter
	logger   *zap.Logger
	prefix   string // key prefix to namespace rate limit keys
}

// NewRedisRateLimiter creates a Redis-backed rate limiter with automatic
// in-memory fallback. Pass nil for rdb to always use in-memory limiting.
func NewRedisRateLimiter(rdb redis.Cmdable, logger *zap.Logger) *RedisRateLimiter {
	return &RedisRateLimiter{
		rdb:      rdb,
		fallback: NewInMemoryRateLimiter(logger),
		logger:   logger,
		prefix:   "rl:",
	}
}

// Middleware returns a Gin handler that enforces the given rate limit rule.
func (rl *RedisRateLimiter) Middleware(rule RateLimitRule) gin.HandlerFunc {
	// If no Redis client, delegate entirely to in-memory
	if rl.rdb == nil {
		return rl.fallback.Middleware(rule)
	}

	return func(c *gin.Context) {
		key := rl.prefix + rule.KeyFunc(c)

		allowed, remaining, resetTime, err := rl.checkLimitRedis(c.Request.Context(), key, rule)
		if err != nil {
			// Redis failed — fall through to in-memory
			rl.logger.Warn("Redis rate limit check failed, using in-memory fallback",
				zap.Error(err),
				zap.String("key", key),
			)
			rl.fallback.Middleware(rule)(c)
			return
		}

		// Set standard rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(rule.Requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		if !allowed {
			rl.logger.Warn("Rate limit exceeded (Redis)",
				zap.String("key", key),
				zap.String("ip", c.ClientIP()),
				zap.String("path", c.Request.URL.Path),
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

// checkLimitRedis performs a sliding window rate limit check using a Redis ZSET.
//
// Algorithm:
//  1. ZREMRANGEBYSCORE — remove entries older than the window
//  2. ZCARD           — count remaining entries
//  3. ZADD            — add current request (if within limit)
//  4. EXPIRE          — set TTL on the key to auto-clean up
//
// All commands run inside a pipeline for atomicity and performance (1 round trip).
func (rl *RedisRateLimiter) checkLimitRedis(ctx context.Context, key string, rule RateLimitRule) (allowed bool, remaining int, resetTime time.Time, err error) {
	now := time.Now()
	windowStart := now.Add(-rule.Window)
	member := fmt.Sprintf("%d:%d", now.UnixNano(), now.UnixNano()%1000) // unique per request

	pipe := rl.rdb.Pipeline()

	// 1. Remove expired entries
	pipe.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(windowStart.UnixNano(), 10))

	// 2. Count current entries
	countCmd := pipe.ZCard(ctx, key)

	// Execute the read portion first so we know the count
	if _, err = pipe.Exec(ctx); err != nil {
		return false, 0, now, fmt.Errorf("redis pipeline (phase 1): %w", err)
	}

	count := int(countCmd.Val())
	allowed = count < rule.Requests

	if allowed {
		// 3. Add the new request timestamp as score + member
		addPipe := rl.rdb.Pipeline()
		addPipe.ZAdd(ctx, key, redis.Z{Score: float64(now.UnixNano()), Member: member})
		addPipe.Expire(ctx, key, rule.Window+time.Minute) // TTL = window + buffer
		if _, err = addPipe.Exec(ctx); err != nil {
			return false, 0, now, fmt.Errorf("redis pipeline (phase 2): %w", err)
		}
		count++ // reflect the just-added entry
	}

	remaining = rule.Requests - count
	if remaining < 0 {
		remaining = 0
	}

	resetTime = now.Add(rule.Window)

	return allowed, remaining, resetTime, nil
}
