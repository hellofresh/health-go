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

const sqlServerDSNEnv = "HEALTH_GO_SS_DSN"

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

	sqlDSN := getDSN(t)

	db, err := sql.Open(azuread.DriverName, sqlDSN)
	require.NoError(t, err)

	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	var initialConnections int
	row := db.QueryRow(`SELECT COUNT(session_id) FROM sys.dm_exec_sessions`)
	err = row.Scan(&initialConnections)
	require.NoError(t, err)

	check := New(Config{
		DSN: sqlDSN,
	})

	ctx := context.Background()
	for i := 0; i < 10; i++ {
		err := check(ctx)
		assert.NoError(t, err)
		time.Sleep(100 * time.Millisecond)
	}

	var currentConnections int
	row = db.QueryRow(`SELECT COUNT(session_id) FROM sys.dm_exec_sessions`)
	err = row.Scan(&currentConnections)
	require.NoError(t, err)

	assert.Equal(t, initialConnections, currentConnections)
}

func getDSN(t *testing.T) string {
	t.Helper()

	//get env
	sqlServerDSN := os.Getenv(sqlServerDSNEnv)
	require.NotEmpty(t, sqlServerDSN)

	return sqlServerDSN
}

var dbInit sync.Once

func initDB(t *testing.T) {
	t.Helper()

	dbInit.Do(func() {
		db, err := sql.Open(azuread.DriverName, getDSN(t))
		require.NoError(t, err)

		defer func() {
			err := db.Close()
			assert.NoError(t, err)
		}()

		_, err = db.Exec(`
IF OBJECT_ID('dbo.test_mssql', 'U') IS NULL
BEGIN
CREATE TABLE test_mssql (
  id           NVARCHAR(255) NOT NULL PRIMARY KEY,
  secret       NVARCHAR(255) NOT NULL,
  extra        NVARCHAR(255) NOT NULL,
  redirect_uri NVARCHAR(255) NOT NULL
);
END
`)
		require.NoError(t, err)
	})
}
