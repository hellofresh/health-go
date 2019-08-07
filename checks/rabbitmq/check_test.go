package rabbitmq

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mqDSNEnv = "HEALTH_GO_MQ_DSN"

func TestNew(t *testing.T) {
	if os.Getenv(mqDSNEnv) == "" {
		t.SkipNow()
	}

	check := New(Config{
		DSN: os.Getenv(mqDSNEnv),
	})

	err := check()
	require.NoError(t, err)
}

func TestConfig(t *testing.T) {
	conf := &Config{
		DSN: os.Getenv(mqDSNEnv),
	}

	conf.defaults()

	assert.Equal(t, defaultExchange, conf.Exchange)
	assert.Equal(t, defaultConsumeTimeout, conf.ConsumeTimeout)
}
