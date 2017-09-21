package redis

import (
	"errors"
	"time"

	"github.com/garyburd/redigo/redis"
	log "github.com/sirupsen/logrus"
)

// New creates new Redis health check that verifies the following:
// - connection establishing
// - doing the PING command and verifying the response
func New(dsn string) func() error {
	return func() error {
		pool := &redis.Pool{
			MaxIdle:     1,
			IdleTimeout: 10 * time.Second,
			Dial:        func() (redis.Conn, error) { return redis.DialURL(dsn) },
		}

		conn := pool.Get()
		defer conn.Close()

		data, err := conn.Do("PING")
		if err != nil {
			log.WithError(err).Error("Redis ping failed")
			return err
		}

		if data == nil {
			log.Error("Empty response for redis ping")
			return errors.New("empty response for redis ping")
		}

		if data != "PONG" {
			log.WithField("data", data).Error("Unexpected response for redis ping")
			return errors.New("unexpected response for redis ping")
		}

		return nil
	}
}
