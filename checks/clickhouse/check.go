package clickhouse

import (
	"context"
	"fmt"

	ch "github.com/ClickHouse/clickhouse-go/v2"
)

// Config is the ClickHouse checker configuration settings container.
type Config struct {
	// DSN is the ClickHouse instance connection DSN. Required.
	DSN string
}

// New creates a new ClickHouse health check that verifies the following:
// - connection establishing
// - doing the ping command
// - selecting clickhouse version
func New(config Config) func(ctx context.Context) error {
	return func(ctx context.Context) (checkErr error) {
		opts, err := ch.ParseDSN(config.DSN)
		if err != nil {
			return fmt.Errorf("ClickHouse health check failed to parse DSN: %w", err)
		}

		conn, err := ch.Open(opts)
		if err != nil {
			return fmt.Errorf("ClickHouse health check failed on connect: %w", err)
		}

		defer func() {
			if err := conn.Close(); err != nil && checkErr == nil {
				checkErr = fmt.Errorf("clickhouse health check failed on connection closing: %w", err)
			}
		}()

		err = conn.Ping(ctx)
		if err != nil {
			checkErr = fmt.Errorf("ClickHouse health check failed on ping: %w", err)
			return
		}

		rows, err := conn.Query(ctx, `SELECT version()`)
		if err != nil {
			checkErr = fmt.Errorf("ClickHouse health check failed on select: %w", err)
			return
		}
		defer rows.Close()

		return
	}
}
