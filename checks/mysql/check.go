package mysql

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql" // import mysql driver
	wErrors "github.com/pkg/errors"
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
func New(config Config) func() error {
	return func() (checkErr error) {
		db, err := sql.Open("mysql", config.DSN)
		if err != nil {
			checkErr = wErrors.Wrap(err, "MySQL health check failed on connect")
			return
		}

		defer func() {
			// override checkErr only if there were no other errors
			if err = db.Close(); err != nil && checkErr == nil {
				checkErr = wErrors.Wrap(err, "MySQL health check failed on connection closing")
			}
		}()

		err = db.Ping()
		if err != nil {
			checkErr = wErrors.Wrap(err, "MySQL health check failed on ping")
			return
		}

		rows, err := db.Query(`SELECT VERSION()`)
		if err != nil {
			checkErr = wErrors.Wrap(err, "MySQL health check failed on select")
			return
		}
		defer func() {
			// override checkErr only if there were no other errors
			if err = rows.Close(); err != nil && checkErr == nil {
				checkErr = wErrors.Wrap(err, "MySQL health check failed on rows closing")
			}
		}()

		return
	}
}
