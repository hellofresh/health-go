package redis

import (
	"errors"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
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
	return func() (checkErr error) {
		pool := &redis.Pool{
			MaxIdle:     1,
			IdleTimeout: 10 * time.Second,
			Dial:        func() (redis.Conn, error) { return redis.DialURL(config.DSN) },
		}

		conn := pool.Get()
		defer func() {
			// override checkErr only if there were no other errors
			if err := conn.Close(); err != nil && checkErr == nil {
				checkErr = fmt.Errorf("Redis health check failed on connection closing: %w", err)
			}
		}()

		data, err := conn.Do("PING")
		if err != nil {
			checkErr = fmt.Errorf("Redis ping failed: %w", err)
			return
		}

		if data == nil {
			checkErr = errors.New("empty response for redis ping")
			return
		}

		if data != "PONG" {
			checkErr = fmt.Errorf("unexpected response for redis ping %s", data)
			return
		}

		return
	}
}
