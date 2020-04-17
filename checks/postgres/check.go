package postgres

import (
	"database/sql"

	_ "github.com/lib/pq" // import pg driver
	wErrors "github.com/pkg/errors"
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
func New(config Config) func() error {
	return func() (checkErr error) {
		db, err := sql.Open("postgres", config.DSN)
		if err != nil {
			checkErr = wErrors.Wrap(err, "PostgreSQL health check failed on connect")
			return
		}

		defer func() {
			// override checkErr only if there were no other errors
			if err := db.Close(); err != nil && checkErr == nil {
				checkErr = wErrors.Wrap(err, "PostgreSQL health check failed on connection closing")
			}
		}()

		err = db.Ping()
		if err != nil {
			checkErr = wErrors.Wrap(err, "PostgreSQL health check failed on ping")
			return
		}

		rows, err := db.Query(`SELECT VERSION()`)
		if err != nil {
			checkErr = wErrors.Wrap(err, "PostgreSQL health check failed on select")
			return
		}
		defer func() {
			// override checkErr only if there were no other errors
			if err = rows.Close(); err != nil && checkErr == nil {
				checkErr = wErrors.Wrap(err, "PostgreSQL health check failed on rows closing")
			}
		}()

		return
	}
}
