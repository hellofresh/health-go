package rabbitmq

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mqDSNEnv = "HEALTH_GO_MQ_DSN"

var mqDSN string

func TestMain(m *testing.M) {
	var ok bool
	if mqDSN, ok = os.LookupEnv(mqDSNEnv); !ok {
		panic(fmt.Errorf("required env variable missing: %s", mqDSNEnv))
	}

	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	check := New(Config{
		DSN: mqDSN,
	})

	err := check()
	require.NoError(t, err)
}

func TestConfig(t *testing.T) {
	conf := &Config{
		DSN: mqDSN,
	}

	conf.defaults()

	assert.Equal(t, defaultExchange, conf.Exchange)
	assert.Equal(t, defaultConsumeTimeout, conf.ConsumeTimeout)
}
