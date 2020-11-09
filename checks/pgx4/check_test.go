package pgx4

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const pgDSNEnv = "HEALTH_GO_PG_PGX4_DSN"

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

	pgDSN := getDSN(t)
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, getDSN(t))
	require.NoError(t, err)

	defer func() {
		err := conn.Close(ctx)
		assert.NoError(t, err)
	}()

	var initialConnections int
	row := conn.QueryRow(ctx, `SELECT sum(numbackends) FROM pg_stat_database`)
	err = row.Scan(&initialConnections)
	require.NoError(t, err)

	check := New(Config{
		DSN: pgDSN,
	})

	for i := 0; i < 10; i++ {
		err := check(ctx)
		assert.NoError(t, err)
		time.Sleep(100 * time.Millisecond)
	}

	var currentConnections int
	row = conn.QueryRow(ctx, `SELECT sum(numbackends) FROM pg_stat_database`)
	err = row.Scan(&currentConnections)
	require.NoError(t, err)

	assert.Equal(t, initialConnections, currentConnections)
}

func getDSN(t *testing.T) string {
	t.Helper()

	pgDSN, ok := os.LookupEnv(pgDSNEnv)
	require.True(t, ok)

	return pgDSN
}

var dbInit sync.Once

func initDB(t *testing.T) {
	t.Helper()

	dbInit.Do(func() {
		ctx := context.Background()

		conn, err := pgx.Connect(ctx, getDSN(t))
		require.NoError(t, err)

		defer func() {
			err := conn.Close(ctx)
			assert.NoError(t, err)
		}()

		_, err = conn.Exec(ctx, `
CREATE TABLE IF NOT EXISTS test_pgx4 (
  id           TEXT NOT NULL PRIMARY KEY,
  secret       TEXT NOT NULL,
  extra        TEXT NOT NULL,
  redirect_uri TEXT NOT NULL
);
`)
		require.NoError(t, err)
	})
}
