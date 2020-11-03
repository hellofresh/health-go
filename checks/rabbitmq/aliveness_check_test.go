package rabbitmq

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hellofresh/health-go/v4/checks/http"
)

const httpURLEnv = "HEALTH_GO_MQ_URL"

func TestAliveness(t *testing.T) {
	check := http.New(http.Config{
		URL: getURL(t),
	})

	err := check(context.Background())
	require.NoError(t, err)
}

func getURL(t *testing.T) string {
	t.Helper()

	httpURL, ok := os.LookupEnv(httpURLEnv)
	require.True(t, ok)

	return httpURL
}
