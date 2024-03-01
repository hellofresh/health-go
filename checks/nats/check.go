package nats

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
)

// Config is the NATS checker configuration settings container.
type Config struct {
	// DSN is the NATS instance connection DSN. Required.
	DSN string
}

// New creates new NATS health check that verifies the status of the connection.
func New(config Config) func(ctx context.Context) error {
	return func(context.Context) error {
		nc, err := nats.Connect(config.DSN)
		if err != nil {
			return fmt.Errorf("nats health check failed on client creation: %w", err)
		}
		defer nc.Close()

		status := nc.Status()
		if status != nats.CONNECTED {
			return fmt.Errorf("nats health check failed as connection status is %s", status)
		}

		return nil
	}
}
