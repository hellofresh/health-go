package redis

import (
	"time"

	"github.com/gomodule/redigo/redis"
	wErrors "github.com/pkg/errors"
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
				checkErr = wErrors.Wrap(err, "Redis health check failed on connection closing")
			}
		}()

		data, err := conn.Do("PING")
		if err != nil {
			checkErr = wErrors.Wrap(err, "Redis ping failed")
			return
		}

		if data == nil {
			checkErr = wErrors.New("empty response for redis ping")
			return
		}

		if data != "PONG" {
			checkErr = wErrors.Errorf("unexpected response for redis ping %s", data)
			return
		}

		return
	}
}
