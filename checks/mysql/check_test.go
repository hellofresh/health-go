package mysql

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mysqlDSNEnv = "HEALTH_GO_MS_DSN"

func getDSN(t *testing.T) string {
	if mysqlDSN, ok := os.LookupEnv(mysqlDSNEnv); ok {
		return mysqlDSN
	}

	t.Fatalf("required env variable missing: %s", mysqlDSNEnv)
	return ""
}

func TestNew(t *testing.T) {
	check := New(Config{
		DSN: getDSN(t),
	})

	err := check()
	require.NoError(t, err)
}

func TestEnsureConnectionIsClosed(t *testing.T) {
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

	for i := 0; i < 10; i++ {
		err := check()
		assert.NoError(t, err)
		time.Sleep(100 * time.Millisecond)
	}

	var currentConnections int
	row = db.QueryRow(`SHOW STATUS WHERE variable_name = 'Threads_connected'`)
	err = row.Scan(&varName, &currentConnections)
	require.NoError(t, err)

	assert.Equal(t, initialConnections, currentConnections)
}
