package http

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const httpURLEnv = "HEALTH_GO_HTTP_URL"

func getURL(t *testing.T) string {
	if httpURL, ok := os.LookupEnv(httpURLEnv); ok {
		return httpURL
	}

	t.Fatalf("required env variable missing: %s", httpURLEnv)
	return ""
}

func TestNew(t *testing.T) {
	check := New(Config{
		URL: getURL(t),
	})

	err := check()
	require.NoError(t, err)
}
