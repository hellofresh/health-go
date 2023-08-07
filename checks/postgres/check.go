package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/hellofresh/health-go/v5"

	_ "github.com/lib/pq" // import pg driver
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
func New(config Config) func(ctx context.Context) health.CheckResponse {
	return func(ctx context.Context) (checkResponse health.CheckResponse) {
		db, err := sql.Open("postgres", config.DSN)
		if err != nil {
			checkResponse.Error = fmt.Errorf("PostgreSQL health check failed on connect: %w", err)
			return
		}

		defer func() {
			// override checkResponse only if there were no other errors
			if err := db.Close(); err != nil && checkResponse.Error == nil {
				checkResponse.Error = fmt.Errorf("PostgreSQL health check failed on connection closing: %w", err)
			}
		}()

		err = db.PingContext(ctx)
		if err != nil {
			checkResponse.Error = fmt.Errorf("PostgreSQL health check failed on ping: %w", err)
			return
		}

		rows, err := db.QueryContext(ctx, `SELECT VERSION()`)
		if err != nil {
			checkResponse.Error = fmt.Errorf("PostgreSQL health check failed on select: %w", err)
			return
		}
		defer func() {
			// override checkResponse only if there were no other errors
			if err = rows.Close(); err != nil && checkResponse.Error == nil {
				checkResponse.Error = fmt.Errorf("PostgreSQL health check failed on rows closing: %w", err)
			}
		}()

		return
	}
}
