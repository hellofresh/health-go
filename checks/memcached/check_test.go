package memcached

import (
	"context"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

const rdDSNEnv = "HEALTH_GO_MD_DSN"

func TestNew(t *testing.T) {
	check := New(Config{
		DSN: getDSN(t),
	})

	err := check(context.Background())
	require.NoError(t, err)
}

func TestNewError(t *testing.T) {
	check := New(Config{
		DSN: "",
	})

	err := check(context.Background())
	require.Error(t, err)
}

func getDSN(t *testing.T) string {
	t.Helper()

	redisDSN, ok := os.LookupEnv(rdDSNEnv)
	require.True(t, ok)

	return redisDSN
}
