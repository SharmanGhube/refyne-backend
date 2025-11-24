package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/refynehq/refyne-backend/internal/config"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// Client wraps the Redis client
type Client struct {
	*redis.Client
	logger *zap.Logger
}

// NewRedisClient creates a new Redis client instance
func NewRedisClient(cfg *config.Config) (*Client, error) {
	logger := logging.GetComponentLogger("redis")
	
	addr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)
	
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})
	
	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("Failed to connect to Redis",
			zap.String("addr", addr),
			zap.Error(err),
		)
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}
	
	logger.Info("Redis connection established successfully",
		zap.String("addr", addr),
		zap.Int("db", cfg.Redis.DB),
	)
	
	return &Client{
		Client: rdb,
		logger: logger,
	}, nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	c.logger.Info("Closing Redis connection")
	return c.Client.Close()
}

// HealthCheck checks if Redis is reachable
func (c *Client) HealthCheck(ctx context.Context) error {
	return c.Ping(ctx).Err()
}
