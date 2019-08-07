package postgres

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const pgDSNEnv = "HEALTH_GO_PG_DSN"

func TestNew(t *testing.T) {
	if os.Getenv(pgDSNEnv) == "" {
		t.SkipNow()
	}

	check := New(Config{
		DSN: os.Getenv(pgDSNEnv),
	})

	err := check()
	require.NoError(t, err)
}

func TestEnsureConnectionIsClosed(t *testing.T) {
	if os.Getenv(pgDSNEnv) == "" {
		t.SkipNow()
	}

	db, err := sql.Open("postgres", os.Getenv(pgDSNEnv))
	require.NoError(t, err)

	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	var initialConnections int
	row := db.QueryRow(`SELECT sum(numbackends) FROM pg_stat_database`)
	err = row.Scan(&initialConnections)
	require.NoError(t, err)

	check := New(Config{
		DSN: os.Getenv(pgDSNEnv),
	})

	for i := 0; i < 10; i++ {
		err := check()
		assert.NoError(t, err)
		time.Sleep(100 * time.Millisecond)
	}

	var currentConnections int
	row = db.QueryRow(`SELECT sum(numbackends) FROM pg_stat_database`)
	err = row.Scan(&currentConnections)
	require.NoError(t, err)

	assert.Equal(t, initialConnections, currentConnections)
}
