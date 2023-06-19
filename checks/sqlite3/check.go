package healthchecks

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// Config is the SQLite3 checker configuration settings container.
type Config struct {
	// DSN is the SQLite3 instance connection DSN. Required.
	DSN string
}

// New creates new SQLite3 health check that verifies the following:
// - connection establishing
// - doing the ping command
// - selecting postgres version
func New(config Config) func(ctx context.Context) error {
	return func(ctx context.Context) (checkErr error) {
		db, err := sql.Open("sqlite3", config.DSN)
		if err != nil {
			checkErr = fmt.Errorf("sqlite health check failed on connect: %w", err)
			return
		}

		defer func() {
			// override checkErr only if there were no other errors
			if err := db.Close(); err != nil && checkErr == nil {
				checkErr = fmt.Errorf("sqlite health check failed on connection closing: %w", err)
			}
		}()

		err = db.PingContext(ctx)
		if err != nil {
			checkErr = fmt.Errorf("sqlite health check failed on ping: %w", err)
			return
		}

		rows, err := db.QueryContext(ctx, `select sqlite_version()`)
		if err != nil {
			checkErr = fmt.Errorf("sqlite health check failed on select: %w", err)
			return
		}
		defer func() {
			// override checkErr only if there were no other errors
			if err = rows.Close(); err != nil && checkErr == nil {
				checkErr = fmt.Errorf("sqlite health check failed on rows closing: %w", err)
			}
		}()

		return
	}
}
