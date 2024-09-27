package clickhouse

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ch "github.com/ClickHouse/clickhouse-go/v2"
)

const chDSNEnv = "HEALTH_GO_CLICKHOUSE_DSN"

func TestNew(t *testing.T) {
	initDB(t)

	check := New(Config{
		DSN: getDSN(t),
	})

	err := check(context.Background())
	require.NoError(t, err)
}

func TestNewWithError(t *testing.T) {
	check := New(Config{})

	err := check(context.Background())
	require.Error(t, err)
}

func getDSN(t *testing.T) string {
	t.Helper()

	chDSN, ok := os.LookupEnv(chDSNEnv)
	require.True(t, ok)

	return chDSN
}

var dbInit sync.Once

func initDB(t *testing.T) {
	t.Helper()

	dbInit.Do(func() {
		ctx := context.Background()

		// 1. Parse DSN
		opts, err := ch.ParseDSN(getDSN(t))
		require.NoError(t, err)

		// 2. Acquire connection
		conn, err := ch.Open(opts)
		require.NoError(t, err)

		defer func() {
			err := conn.Close()
			assert.NoError(t, err)
		}()

		// 3. Create test table
		err = conn.Exec(ctx, `
			CREATE TABLE IF NOT EXISTS test_clickhouse (
			    id        Int64,
			    timestamp DateTime
			)
			ENGINE = MergeTree()
			ORDER BY id;`)
		require.NoError(t, err)

	})
}
