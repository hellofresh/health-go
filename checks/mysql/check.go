package mysql

import (
	"database/sql"
	"fmt"
	"strings"
	"errors"

	// MySQL database driver
	_ "github.com/go-sql-driver/mysql"
)

// Config is the MySQL checker configuration settings container.
type Config struct {
	// DSN is the MySQL instance connection DSN. Required.
	DSN string

	// Table is the table name used for testing, must already exist in the DB and
	// has insert/select/delete privileges for the connection user. Required.
	Table string

	// IDColumn is the primary column for the table used for testing. Required.
	IDColumn string

	// InsertColumnsFunc is the callback function that returns map[<column>]<value>
	// for testing insert operation. Required.
	InsertColumnsFunc func() map[string]interface{}

	// LogFunc is the callback function for errors logging during check.
	// If not set logging is skipped.
	LogFunc func(err error, details string, extra ...interface{})
}

// New creates new MySQL health check that verifies the following:
// - connection establishing
// - inserting a row into defined table
// - selecting inserted row
// - deleting inserted row
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

		columns := config.InsertColumnsFunc()
		columnNames := []string{}
		columnPlaceholders := []string{}
		columnValues := []interface{}{}
		i := 1
		for column, value := range columns {
			columnNames = append(columnNames, column)
			columnPlaceholders = append(columnPlaceholders, "?")
			columnValues = append(columnValues, value)

			i++
		}

		insertQuery := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s)",
			config.Table,
			strings.Join(columnNames, ", "),
			strings.Join(columnPlaceholders, ", "),
		)

		result, err := db.Exec(insertQuery, columnValues...)
		if err != nil {
			config.LogFunc(err, "MySQL health check failed during insert")
			return err
		}

		idValue, err := result.LastInsertId()
		if err != nil {
			config.LogFunc(err, "MySQL health check failed during retrieval of last inserted id")
			return err
		}

		selectQuery := fmt.Sprintf(
			"SELECT %s FROM %s WHERE %s = ?",
			strings.Join(columnNames, ", "),
			config.Table,
			config.IDColumn,
		)
		selectRows, err := db.Query(selectQuery, idValue)
		if err != nil {
			config.LogFunc(err, "MySQL health check failed during select")
			return err
		}

		if !selectRows.Next() {
			config.LogFunc(err, "MySQL health check failed during checking select result rows")
			return errors.New("looks like select result has 0 rows")
		}
		err = selectRows.Close()
		if err != nil {
			config.LogFunc(err, "MySQL health check failed during closing select result")
			return err
		}

		deleteQuery := fmt.Sprintf(
			"DELETE FROM %s WHERE %s = ?",
			config.Table,
			config.IDColumn,
		)
		deleteResult, err := db.Exec(deleteQuery, idValue)
		if err != nil {
			config.LogFunc(err, "MySQL health check failed during delete")
			return err
		}
		deleted, err := deleteResult.RowsAffected()
		if err != nil {
			config.LogFunc(err, "MySQL health check failed during extracting delete result")
			return err
		}
		if deleted < 1 {
			config.LogFunc(err, "MySQL health check failed during checking delete result")
			return errors.New("looks like delete removed 0 rows")
		}

		return nil
	}
}
