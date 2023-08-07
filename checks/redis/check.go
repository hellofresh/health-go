package redis

import (
	"context"
	"fmt"
	"github.com/hellofresh/health-go/v5"
	"strings"

	"github.com/go-redis/redis/v9"
)

// Config is the Redis checker configuration settings container.
type Config struct {
	// DSN is the Redis instance connection DSN. Required.
	DSN string
}

// New creates new Redis health check that verifies the following:
// - connection establishing
// - doing the PING command and verifying the response
func New(config Config) func(ctx context.Context) health.CheckResponse {
	// support all DSN formats (for backward compatibility) - with and w/out schema and path part:
	// - redis://localhost:1234/
	// - rediss://localhost:1234/
	// - localhost:1234
	redisDSN := config.DSN
	if !strings.HasPrefix(redisDSN, "redis://") && !strings.HasPrefix(redisDSN, "rediss://") {
		redisDSN = fmt.Sprintf("redis://%s", redisDSN)
	}
	redisOptions, _ := redis.ParseURL(redisDSN)

	return func(ctx context.Context) (checkResponse health.CheckResponse) {
		rdb := redis.NewClient(redisOptions)
		defer rdb.Close()

		pong, err := rdb.Ping(ctx).Result()
		if err != nil {
			checkResponse.Error = fmt.Errorf("redis ping failed: %w", err)
			return
		}

		if pong != "PONG" {
			checkResponse.Error = fmt.Errorf("unexpected response for redis ping: %q", pong)
			return
		}

		return
	}
}
