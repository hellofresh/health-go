package mongo

import (
	"context"
	"os"
	"strings"
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

	// "docker-compose port <service> <port>" returns 0.0.0.0:XXXX locally, change it to local port
	mongoDSN = strings.Replace(mongoDSN, "0.0.0.0:", "127.0.0.1:", 1)

	return mongoDSN
}
