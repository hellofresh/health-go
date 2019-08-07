package rabbitmq

import (
	"os"
	"testing"

	"github.com/hellofresh/health-go/checks/http"
	"github.com/stretchr/testify/require"
)

const httpURLEnv = "HEALTH_GO_MQ_URL"

func TestAliveness(t *testing.T) {
	if os.Getenv(httpURLEnv) == "" {
		t.SkipNow()
	}

	check := http.New(http.Config{
		URL: os.Getenv(httpURLEnv),
	})

	err := check()
	require.NoError(t, err)
}
