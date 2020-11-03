package redis

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const rdDSNEnv = "HEALTH_GO_RD_DSN"

func TestNew(t *testing.T) {
	check := New(Config{
		DSN: getDSN(t),
	})

	err := check(context.Background())
	require.NoError(t, err)
}

func getDSN(t *testing.T) string {
	t.Helper()

	redisDSN, ok := os.LookupEnv(rdDSNEnv)
	require.True(t, ok)

	return redisDSN
}
