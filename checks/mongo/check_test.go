package mongo

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const mgDSNEnv = "HEALTH_GO_MG_DSN"

func TestNew(t *testing.T) {
	if os.Getenv(mgDSNEnv) == "" {
		t.SkipNow()
	}

	check := New(Config{
		DSN: os.Getenv(mgDSNEnv),
	})

	err := check()
	require.NoError(t, err)
}
