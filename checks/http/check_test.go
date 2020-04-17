package http

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const httpURLEnv = "HEALTH_GO_HTTP_URL"

func TestNew(t *testing.T) {
	check := New(Config{
		URL: getURL(t),
	})

	err := check()
	require.NoError(t, err)
}

func getURL(t *testing.T) string {
	t.Helper()

	httpURL, ok := os.LookupEnv(httpURLEnv)
	require.True(t, ok)

	return httpURL
}
