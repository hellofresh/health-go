package rabbitmq

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mqDSNEnv = "HEALTH_GO_MQ_DSN"

func TestNew(t *testing.T) {
	check := New(Config{
		DSN: getDSN(t),
	})

	err := check(context.Background())
	require.NoError(t, err)
}

func TestConfig(t *testing.T) {
	conf := &Config{
		DSN: getDSN(t),
	}

	conf.defaults()

	assert.Equal(t, defaultExchange, conf.Exchange)
	assert.Equal(t, defaultConsumeTimeout, conf.ConsumeTimeout)
}

func getDSN(t *testing.T) string {
	t.Helper()

	mqDSN, ok := os.LookupEnv(mqDSNEnv)
	require.True(t, ok)

	// "docker-compose port <service> <port>" returns 0.0.0.0:XXXX locally, change it to local port
	mqDSN = strings.Replace(mqDSN, "0.0.0.0:", "127.0.0.1:", 1)

	return mqDSN
}
