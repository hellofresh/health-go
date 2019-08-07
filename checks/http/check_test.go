package http

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const httpURLEnv = "HEALTH_GO_HTTP_URL"

func TestNew(t *testing.T) {
	if os.Getenv(httpURLEnv) == "" {
		t.SkipNow()
	}

	check := New(Config{
		URL: os.Getenv(httpURLEnv),
	})

	err := check()
	require.NoError(t, err)
}
