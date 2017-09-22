package redis

import (
	"errors"
	"time"

	"github.com/garyburd/redigo/redis"
)

// Config is the Redis checker configuration settings container.
type Config struct {
	// DSN is the Redis instance connection DSN. Required.
	DSN string
	// LogFunc is the callback function for errors logging during check.
	// If not set logging is skipped.
	LogFunc func(err error, details string, extra ...interface{})
}

// New creates new Redis health check that verifies the following:
// - connection establishing
// - doing the PING command and verifying the response
func New(config Config) func() error {
	return func() error {
		if config.LogFunc == nil {
			config.LogFunc = func(err error, details string, extra ...interface{}) {}
		}

		pool := &redis.Pool{
			MaxIdle:     1,
			IdleTimeout: 10 * time.Second,
			Dial:        func() (redis.Conn, error) { return redis.DialURL(config.DSN) },
		}

		conn := pool.Get()
		defer conn.Close()

		data, err := conn.Do("PING")
		if err != nil {
			config.LogFunc(err, "Redis ping failed")
			return err
		}

		if data == nil {
			config.LogFunc(nil, "Empty response for redis ping")
			return errors.New("empty response for redis ping")
		}

		if data != "PONG" {
			config.LogFunc(nil, "Unexpected response for redis ping", data)
			return errors.New("unexpected response for redis ping")
		}

		return nil
	}
}
