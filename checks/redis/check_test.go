package redis

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const rdDSNEnv = "HEALTH_GO_RD_DSN"

var redisDSN string

func TestMain(m *testing.M) {
	var ok bool
	if redisDSN, ok = os.LookupEnv(rdDSNEnv); !ok {
		panic(fmt.Errorf("required env variable missing: %s", rdDSNEnv))
	}

	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	check := New(Config{
		DSN: redisDSN,
	})

	err := check()
	require.NoError(t, err)
}
