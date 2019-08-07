package postgres

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const pgDSNEnv = "HEALTH_GO_PG_DSN"

var pgDSN string

func TestMain(m *testing.M) {
	var ok bool
	if pgDSN, ok = os.LookupEnv(pgDSNEnv); !ok {
		panic(fmt.Errorf("required env variable missing: %s", pgDSNEnv))
	}

	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	check := New(Config{
		DSN: pgDSN,
	})

	err := check()
	require.NoError(t, err)
}

func TestEnsureConnectionIsClosed(t *testing.T) {
	db, err := sql.Open("postgres", pgDSN)
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
		DSN: pgDSN,
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
