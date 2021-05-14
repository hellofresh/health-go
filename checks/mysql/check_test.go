package mysql

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mysqlDSNEnv = "HEALTH_GO_MS_DSN"

func TestNew(t *testing.T) {
	initDB(t)

	check := New(Config{
		DSN: getDSN(t),
	})

	err := check(context.Background())
	require.NoError(t, err)
}

func TestEnsureConnectionIsClosed(t *testing.T) {
	initDB(t)

	mysqlDSN := getDSN(t)

	db, err := sql.Open("mysql", mysqlDSN)
	require.NoError(t, err)

	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	var (
		varName            string
		initialConnections int
	)
	row := db.QueryRow(`SHOW STATUS WHERE variable_name = 'Threads_connected'`)
	err = row.Scan(&varName, &initialConnections)
	require.NoError(t, err)

	check := New(Config{
		DSN: mysqlDSN,
	})

	ctx := context.Background()
	for i := 0; i < 10; i++ {
		err := check(ctx)
		assert.NoError(t, err)
		time.Sleep(100 * time.Millisecond)
	}

	var currentConnections int
	row = db.QueryRow(`SHOW STATUS WHERE variable_name = 'Threads_connected'`)
	err = row.Scan(&varName, &currentConnections)
	require.NoError(t, err)

	assert.Equal(t, initialConnections, currentConnections)
}

func getDSN(t *testing.T) string {
	t.Helper()

	mysqlDSN, ok := os.LookupEnv(mysqlDSNEnv)
	require.True(t, ok)

	// "docker-compose port <service> <port>" returns 0.0.0.0:XXXX locally, change it to local port
	mysqlDSN = strings.Replace(mysqlDSN, "0.0.0.0:", "127.0.0.1:", 1)

	return mysqlDSN
}

var dbInit sync.Once

func initDB(t *testing.T) {
	t.Helper()

	dbInit.Do(func() {
		db, err := sql.Open("mysql", getDSN(t))
		require.NoError(t, err)

		defer func() {
			err := db.Close()
			assert.NoError(t, err)
		}()

		_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS test (
  id           INT NOT NULL AUTO_INCREMENT PRIMARY KEY ,
  secret       VARCHAR(256) NOT NULL,
  extra        VARCHAR(256) NOT NULL,
  redirect_uri VARCHAR(256) NOT NULL
);
`)
		require.NoError(t, err)
	})
}
