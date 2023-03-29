package sqlserver

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/microsoft/go-mssqldb/azuread"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sqlServerDSNEnv = "HEALTH_SS_DSN"

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

	sqlServerDSN := getDSN(t)

	db, err := sql.Open(azuread.DriverName, sqlServerDSN)
	require.NoError(t, err)

	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	var (
		initialConnections int
	)
	row := db.QueryRow(`SELECT cntr_value FROM sys.dm_os_performance_counters WHERE counter_name = 'User Connections' AND instance_name = 'Total'`)
	err = row.Scan(&initialConnections)
	require.NoError(t, err)

	check := New(Config{
		DSN: sqlServerDSN,
	})

	ctx := context.Background()
	for i := 0; i < 10; i++ {
		err := check(ctx)
		assert.NoError(t, err)
		time.Sleep(100 * time.Millisecond)
	}

	var currentConnections int
	row = db.QueryRow(`SELECT cntr_value FROM sys.dm_os_performance_counters WHERE counter_name = 'User Connections' AND instance_name = 'Total'`)
	err = row.Scan(&currentConnections)
	require.NoError(t, err)

	assert.Equal(t, initialConnections, currentConnections)
}

func getDSN(t *testing.T) string {
	t.Helper()

	sqlServerDSN, ok := os.LookupEnv(sqlServerDSNEnv)
	require.True(t, ok)

	return sqlServerDSN
}

var dbInit sync.Once

func initDB(t *testing.T) {
	t.Helper()

	dbInit.Do(func() {
		db, err := sql.Open("sqlserver", getDSN(t))
		require.NoError(t, err)

		defer func() {
			err := db.Close()
			assert.NoError(t, err)
		}()

		_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS test (
  id           INT NOT NULL IDENTITY(1,1) PRIMARY KEY ,
  secret       VARCHAR(256) NOT NULL,
  extra        VARCHAR(256) NOT NULL,
  redirect_uri VARCHAR(256) NOT NULL
);
`)
		require.NoError(t, err)
	})
}
