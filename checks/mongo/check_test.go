package mongo

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const mgDSNEnv = "HEALTH_GO_MG_DSN"

func getDSN(t *testing.T) string {
	if mongoDSN, ok := os.LookupEnv(mgDSNEnv); ok {
		return mongoDSN
	}

	t.Fatalf("required env variable missing: %s", mgDSNEnv)
	return ""
}

func TestNew(t *testing.T) {
	check := New(Config{
		DSN: getDSN(t),
	})

	err := check()
	require.NoError(t, err)
}
