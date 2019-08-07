package redis

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const rdDSNEnv = "HEALTH_GO_RD_DSN"

func TestNew(t *testing.T) {
	if os.Getenv(rdDSNEnv) == "" {
		t.SkipNow()
	}

	check := New(Config{
		DSN: os.Getenv(rdDSNEnv),
	})

	err := check()
	require.NoError(t, err)
}
