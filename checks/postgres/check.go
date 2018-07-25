package postgres

import (
	"database/sql"
	_ "github.com/lib/pq"
)

// Config is the PostgreSQL checker configuration settings container.
type Config struct {
	// DSN is the PostgreSQL instance connection DSN. Required.
	DSN string
	// If not set logging is skipped.
	LogFunc func(err error, details string, extra ...interface{})
}

// New creates new PostgreSQL health check that verifies the following:
// - connection establishing
// - doing the ping command
func New(config Config) func() error {
	if config.LogFunc == nil {
		config.LogFunc = func(err error, details string, extra ...interface{}) {}
	}

	return func() error {
		db, err := sql.Open("postgres", config.DSN)
		if err != nil {
			config.LogFunc(err, "PostgreSQL health check failed during connect")
			return err
		}

		defer func() {
			if err = db.Close(); err != nil {
				config.LogFunc(err, "PostgreSQL health check failed during connection closing")
			}
		}()

		err = db.Ping()
		if err != nil {
			config.LogFunc(err, "PostgreSQL health check failed during ping")
			return err
		}

		return nil
	}
}
