package redis

import (
	"fmt"
	"strings"

	"github.com/go-redis/redis/v7"
)

// Config is the Redis checker configuration settings container.
type Config struct {
	// DSN is the Redis instance connection DSN. Required.
	DSN string
}

// New creates new Redis health check that verifies the following:
// - connection establishing
// - doing the PING command and verifying the response
func New(config Config) func() error {
	// support all DSN formats (for backward compatibility) - with and w/out schema and path part:
	// - redis://localhost:1234/
	// - localhost:1234
	redisDSN := strings.TrimPrefix(config.DSN, "redis://")
	redisDSN = strings.TrimSuffix(redisDSN, "/")

	return func() error {
		rdb := redis.NewClient(&redis.Options{
			Addr: redisDSN,
		})
		defer rdb.Close()

		pong, err := rdb.Ping().Result()
		if err != nil {
			return fmt.Errorf("redis ping failed: %w", err)
		}

		if pong != "PONG" {
			return fmt.Errorf("unexpected response for redis ping: %q", pong)
		}

		return nil
	}
}
