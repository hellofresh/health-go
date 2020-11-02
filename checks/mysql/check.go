package mysql

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql" // import mysql driver
)

// Config is the MySQL checker configuration settings container.
type Config struct {
	// DSN is the MySQL instance connection DSN. Required.
	DSN string
}

// New creates new MySQL health check that verifies the following:
// - connection establishing
// - doing the ping command
// - selecting mysql version
func New(config Config) func(ctx context.Context) error {
	return func(ctx context.Context) (checkErr error) {
		db, err := sql.Open("mysql", config.DSN)
		if err != nil {
			checkErr = fmt.Errorf("MySQL health check failed on connect: %w", err)
			return
		}

		defer func() {
			// override checkErr only if there were no other errors
			if err = db.Close(); err != nil && checkErr == nil {
				checkErr = fmt.Errorf("MySQL health check failed on connection closing: %w", err)
			}
		}()

		err = db.PingContext(ctx)
		if err != nil {
			checkErr = fmt.Errorf("MySQL health check failed on ping: %w", err)
			return
		}

		rows, err := db.QueryContext(ctx, `SELECT VERSION()`)
		if err != nil {
			checkErr = fmt.Errorf("MySQL health check failed on select: %w", err)
			return
		}
		defer func() {
			// override checkErr only if there were no other errors
			if err = rows.Close(); err != nil && checkErr == nil {
				checkErr = fmt.Errorf("MySQL health check failed on rows closing: %w", err)
			}
		}()

		return
	}
}
