package rabbitmq

import (
	"os"
	"testing"

	"github.com/hellofresh/health-go/checks/http"
	"github.com/stretchr/testify/require"
)

const httpURLEnv = "HEALTH_GO_MQ_URL"

func getURL(t *testing.T) string {
	if httpURL, ok := os.LookupEnv(httpURLEnv); ok {
		return httpURL
	}

	t.Fatalf("required env variable missing: %s", httpURLEnv)
	return ""
}

func TestAliveness(t *testing.T) {
	check := http.New(http.Config{
		URL: getURL(t),
	})

	err := check()
	require.NoError(t, err)
}
