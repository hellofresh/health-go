package mysql

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql" // import mysql driver
)

// Config is the MySQL checker configuration settings container.
type Config struct {
	// DSN is the MySQL instance connection DSN. Required.
	DSN string

	// LogFunc is the callback function for errors logging during check.
	// If not set logging is skipped.
	LogFunc func(err error, details string, extra ...interface{})
}

// New creates new MySQL health check that verifies the following:
// - connection establishing
// - doing the ping command
// - selecting mysql version
func New(config Config) func() error {
	if config.LogFunc == nil {
		config.LogFunc = func(err error, details string, extra ...interface{}) {}
	}

	return func() error {
		db, err := sql.Open("mysql", config.DSN)
		if err != nil {
			config.LogFunc(err, "MySQL health check failed during connect")
			return err
		}

		defer func() {
			if err = db.Close(); err != nil {
				config.LogFunc(err, "MySQL health check failed during connection closing")
			}
		}()

		err = db.Ping()
		if err != nil {
			config.LogFunc(err, "MySQL health check failed during ping")
			return err
		}

		rows, err := db.Query(`SELECT VERSION()`)
		if err != nil {
			config.LogFunc(err, "MySQL health check failed during select")
			return err
		}
		defer func() {
			if err = rows.Close(); err != nil {
				config.LogFunc(err, "MySQL health check failed during rows closing")
			}
		}()

		return nil
	}
}
