package neo4j

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	neo4jDSNEnv = "HEALTH_GO_N4J_DSN"
)

func TestNew(t *testing.T) {
	check := New(Config{
		DSN: getDSN(t),
	})

	err := check(context.Background())
	require.NoError(t, err)
}

func getDSN(t *testing.T) string {
	t.Helper()

	neo4jDSN, ok := os.LookupEnv(neo4jDSNEnv)
	require.True(t, ok)

	// "docker-compose port <service> <port>" returns 0.0.0.0:XXXX locally, change it to local port
	neo4jDSN = strings.Replace(neo4jDSN, "0.0.0.0:", "127.0.0.1:", 1)

	return neo4jDSN
}
