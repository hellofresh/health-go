package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

// Config is the PostgreSQL checker configuration settings container.
type Config struct {
	// DSN is the PostgreSQL instance connection DSN. Required.
	DSN string
	// Table is the table name used for testing, must already exist in the DB and has insert/select/delete
	// privileges for the connection user. Required.
	Table string
	// IDColumn is the primary column for the table used for testing. Required.
	IDColumn string
	// InsertColumnsFunc is the callback function that returns map[<column>]<value> for testing insert operation.
	// Required.
	InsertColumnsFunc func() map[string]interface{}
	// LogFunc is the callback function for errors logging during check.
	// If not set logging is skipped.
	LogFunc func(err error, details string, extra ...interface{})
}

// New creates new PostgreSQL health check that verifies the following:
// - connection establishing
// - inserting a row into defined table
// - selecting inserted row
// - deleting inserted row
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

		columns := config.InsertColumnsFunc()
		columnNames := []string{}
		columnPlaceholders := []string{}
		columnValues := []interface{}{}
		i := 1
		for column, value := range columns {
			columnNames = append(columnNames, column)
			columnPlaceholders = append(columnPlaceholders, fmt.Sprintf("$%d", i))
			columnValues = append(columnValues, value)

			i++
		}

		insertQuery := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s) RETURNING %s",
			config.Table,
			strings.Join(columnNames, ", "),
			strings.Join(columnPlaceholders, ", "),
			config.IDColumn,
		)

		var idValue interface{}
		err = db.QueryRow(insertQuery, columnValues...).Scan(&idValue)
		if err != nil {
			config.LogFunc(err, "PostgreSQL health check failed during insert and scan")
			return err
		}

		selectQuery := fmt.Sprintf(
			"SELECT %s FROM %s WHERE %s = $1",
			strings.Join(columnNames, ", "),
			config.Table,
			config.IDColumn,
		)
		selectRows, err := db.Query(selectQuery, idValue)
		if err != nil {
			config.LogFunc(err, "PostgreSQL health check failed during select")
			return err
		}
		if !selectRows.Next() {
			config.LogFunc(err, "PostgreSQL health check failed during checking select result rows")
			return errors.New("looks like select result has 0 rows")
		}
		err = selectRows.Close()
		if err != nil {
			config.LogFunc(err, "PostgreSQL health check failed during closing select result")
			return err
		}

		deleteQuery := fmt.Sprintf(
			"DELETE FROM %s WHERE %s = $1",
			config.Table,
			config.IDColumn,
		)
		deleteResult, err := db.Exec(deleteQuery, idValue)
		if err != nil {
			config.LogFunc(err, "PostgreSQL health check failed during delete")
			return err
		}
		deleted, err := deleteResult.RowsAffected()
		if err != nil {
			config.LogFunc(err, "PostgreSQL health check failed during extracting delete result")
			return err
		}
		if deleted < 1 {
			config.LogFunc(err, "PostgreSQL health check failed during checking delete result")
			return errors.New("looks like delete removed 0 rows")
		}

		return nil
	}
}
