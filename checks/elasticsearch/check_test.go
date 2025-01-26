package elasticsearch

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const esDSNEnv = "HEALTH_GO_ES_DSN"

func TestNew(t *testing.T) {
	t.Parallel()

	check := New(getConfig(t))

	err := check(context.Background())
	require.NoError(t, err)
}

func getConfig(t *testing.T) Config {
	t.Helper()

	elasticSearchDSN, ok := os.LookupEnv(esDSNEnv)
	require.True(t, ok, "HEALTH_GO_ES_DSN environment variable not set")

	// "docker-compose port <service> <port>" returns 0.0.0.0:XXXX locally, change it to local port
	elasticSearchDSN = strings.Replace(elasticSearchDSN, "0.0.0.0:", "127.0.0.1:", 1)

	return Config{
		DSN:      elasticSearchDSN,
		Password: "test", // Set in docker-compose.yml
	}
}
