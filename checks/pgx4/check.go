package pgx4

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
)

// Config is the PostgreSQL checker configuration settings container.
type Config struct {
	// DSN is the PostgreSQL instance connection DSN. Required.
	DSN string
}

// New creates new PostgreSQL health check that verifies the following:
// - connection establishing
// - doing the ping command
// - selecting postgres version
func New(config Config) func(ctx context.Context) error {
	return func(ctx context.Context) (checkErr error) {
		conn, err := pgx.Connect(ctx, config.DSN)
		if err != nil {
			checkErr = fmt.Errorf("PostgreSQL health check failed on connect: %w", err)
			return
		}

		defer func() {
			// override checkErr only if there were no other errors
			if err := conn.Close(ctx); err != nil && checkErr == nil {
				checkErr = fmt.Errorf("PostgreSQL health check failed on connection closing: %w", err)
			}
		}()

		err = conn.Ping(ctx)
		if err != nil {
			checkErr = fmt.Errorf("PostgreSQL health check failed on ping: %w", err)
			return
		}

		rows, err := conn.Query(ctx, `SELECT VERSION()`)
		if err != nil {
			checkErr = fmt.Errorf("PostgreSQL health check failed on select: %w", err)
			return
		}
		defer func() {
			rows.Close()
		}()

		return
	}
}
