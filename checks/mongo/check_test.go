package mongo

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const mgDSNEnv = "HEALTH_GO_MG_DSN"

func TestNew(t *testing.T) {
	check := New(Config{
		DSN: getDSN(t),
	})

	err := check(context.Background())
	require.NoError(t, err)
}

func getDSN(t *testing.T) string {
	t.Helper()

	mongoDSN, ok := os.LookupEnv(mgDSNEnv)
	require.True(t, ok)

	return mongoDSN
}
