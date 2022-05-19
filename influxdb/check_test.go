package influxdb

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const InfluxDbUrlEnv = "HEALTH_GO_INFLUXDB_URL"

func TestNew(t *testing.T) {
	check := New(Config{
		URL: getURL(t),
	})

	err := check(context.Background())
	require.NoError(t, err)
}

func TestNewWithError(t *testing.T) {
	check := New(Config{
		URL: "",
	})

	err := check(context.Background())
	require.Error(t, err)
}

func getURL(t *testing.T) string {
	t.Helper()

	url, ok := os.LookupEnv(InfluxDbUrlEnv)

	require.True(t, ok)

	return url
}
