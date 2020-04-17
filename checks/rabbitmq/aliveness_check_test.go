package rabbitmq

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hellofresh/health-go/v3/checks/http"
)

const httpURLEnv = "HEALTH_GO_MQ_URL"

func TestAliveness(t *testing.T) {
	check := http.New(http.Config{
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
