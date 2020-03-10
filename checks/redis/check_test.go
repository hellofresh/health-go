package redis

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const rdDSNEnv = "HEALTH_GO_RD_DSN"

func getDSN(t *testing.T) string {
	if redisDSN, ok := os.LookupEnv(rdDSNEnv); ok {
		return redisDSN
	}

	t.Fatalf("required env variable missing: %s", rdDSNEnv)
	return ""
}

func TestNew(t *testing.T) {
	check := New(Config{
		DSN: getDSN(t),
	})

	err := check()
	require.NoError(t, err)
}
