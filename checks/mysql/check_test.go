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

func TestNew(t *testing.T) {
	if os.Getenv(mysqlDSNEnv) == "" {
		t.SkipNow()
	}

	check := New(Config{
		DSN: os.Getenv(mysqlDSNEnv),
	})

	err := check()
	require.NoError(t, err)
}

func TestEnsureConnectionIsClosed(t *testing.T) {
	if os.Getenv(mysqlDSNEnv) == "" {
		t.SkipNow()
	}

	db, err := sql.Open("mysql", os.Getenv(mysqlDSNEnv))
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
		DSN: os.Getenv(mysqlDSNEnv),
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
