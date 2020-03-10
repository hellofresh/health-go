package rabbitmq

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mqDSNEnv = "HEALTH_GO_MQ_DSN"

func getDSN(t *testing.T) string {
	if mqDSN, ok := os.LookupEnv(mqDSNEnv); ok {
		return mqDSN
	}

	t.Fatalf("required env variable missing: %s", mqDSNEnv)
	return ""
}

func TestNew(t *testing.T) {
	check := New(Config{
		DSN: getDSN(t),
	})

	err := check()
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
