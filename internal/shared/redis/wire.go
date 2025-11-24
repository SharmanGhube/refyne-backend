package redis

import (
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

// ProvideRedisClient extracts the underlying redis.Client from our wrapped Client
func ProvideRedisClient(c *Client) *redis.Client {
	return c.Client
}

// ProviderSet provides Redis dependencies
var ProviderSet = wire.NewSet(
	NewRedisClient,
	ProvideRedisClient,
)
