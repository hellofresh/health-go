package nats

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const natsDSNEnv = "HEALTH_GO_NATS_DSN"

func TestNew(t *testing.T) {
	check := New(Config{
		DSN: getDSN(t),
	})

	err := check(context.Background())
	require.NoError(t, err.Error)
}

func getDSN(t *testing.T) string {
	t.Helper()

	dsn, ok := os.LookupEnv(natsDSNEnv)
	require.True(t, ok)

	return dsn
}
