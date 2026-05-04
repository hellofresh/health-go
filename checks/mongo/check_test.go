package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const mgDSNEnv = "HEALTH_GO_MG_DSN"

func TestNewPingCheck(t *testing.T) {
	ctx := context.Background()

	t.Run("success check with valid client", func(t *testing.T) {
		dsn := getDSN(t)

		client, err := mongo.Connect(ctx, options.Client().ApplyURI(dsn))
		require.NoError(t, err)

		defer func() {
			errDisc := client.Disconnect(ctx)
			require.NoError(t, errDisc)
		}()

		check := NewPingCheck(client, 0)

		err = check(ctx)
		require.NoError(t, err)
	})

	t.Run("fails on nil client", func(t *testing.T) {
		check := NewPingCheck(nil, 5*time.Second)

		err := check(ctx)
		require.ErrorIs(t, err, errNilClient)
	})
}

func TestNew(t *testing.T) {
	check := New(Config{
		DSN: getDSN(t),
	})

	err := check(context.Background())
	require.NoError(t, err)
}

func getDSN(t *testing.T) string {
	t.Helper()

	mongoDSN, ok := os.LookupEnv(mgDSNEnv)
	require.True(t, ok)

	// "docker compose port <service> <port>" returns 0.0.0.0:XXXX locally, change it to local port
	mongoDSN = strings.Replace(mongoDSN, "0.0.0.0:", "127.0.0.1:", 1)

	return mongoDSN
}
