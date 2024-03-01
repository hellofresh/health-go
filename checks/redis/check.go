package redis

import (
	"context"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
)

// Config is the Redis checker configuration settings container.
type Config struct {
	// DSN is the Redis instance connection DSN. Required.
	DSN string
}

// New creates new Redis health check that verifies the following:
// - connection establishing
// - doing the PING command and verifying the response
func New(config Config) func(ctx context.Context) error {
	// support all DSN formats (for backward compatibility) - with and w/out schema and path part:
	// - redis://localhost:1234/
	// - rediss://localhost:1234/
	// - localhost:1234
	redisDSN := config.DSN
	if !strings.HasPrefix(redisDSN, "redis://") && !strings.HasPrefix(redisDSN, "rediss://") {
		redisDSN = fmt.Sprintf("redis://%s", redisDSN)
	}
	redisOptions, _ := redis.ParseURL(redisDSN)

	return func(ctx context.Context) error {
		rdb := redis.NewClient(redisOptions)
		defer rdb.Close()

		pong, err := rdb.Ping(ctx).Result()
		if err != nil {
			return fmt.Errorf("redis ping failed: %w", err)
		}

		if pong != "PONG" {
			return fmt.Errorf("unexpected response for redis ping: %q", pong)
		}

		return nil
	}
}
