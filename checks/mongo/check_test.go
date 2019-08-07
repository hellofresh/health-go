package mongo

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const mgDSNEnv = "HEALTH_GO_MG_DSN"

var mongoDSN string

func TestMain(m *testing.M) {
	var ok bool
	if mongoDSN, ok = os.LookupEnv(mgDSNEnv); !ok {
		panic(fmt.Errorf("required env variable missing: %s", mgDSNEnv))
	}

	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	check := New(Config{
		DSN: mongoDSN,
	})

	err := check()
	require.NoError(t, err)
}
