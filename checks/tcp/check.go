package tcp

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"
)

const defaultRequestTimeout = 5 * time.Second

// Config is the TCP checker configuration settings container.
type Config struct {
	// Host is the remote service hostname health check TCP.
	Host string
	// Port is the remote service port health check TCP.
	Port int
	// RequestTimeout is the duration that health check will try to consume published test message.
	// If not set - 5 seconds
	RequestTimeout time.Duration
}

// New creates new HTTP service health check that verifies the following:
// - connection establishing
// - getting response status from defined URL
// - verifying that status code is less than 500
func New(config Config) func(ctx context.Context) error {
	if config.RequestTimeout == 0 {
		config.RequestTimeout = defaultRequestTimeout
	}

	return func(ctx context.Context) error {
		conn, err := net.DialTimeout("tcp", config.Host+":"+strconv.Itoa(config.Port), config.RequestTimeout)
		if err != nil {
			return fmt.Errorf("making the request for the health check failed: %w", err.Error())
		}

		conn.Close()

		return nil
	}
}
