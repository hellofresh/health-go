package http

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const httpURLEnv = "HEALTH_GO_HTTP_URL"

var httpURL string

func TestMain(m *testing.M) {
	var ok bool
	if httpURL, ok = os.LookupEnv(httpURLEnv); !ok {
		panic(fmt.Errorf("required env variable missing: %s", httpURLEnv))
	}

	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	check := New(Config{
		URL: httpURL,
	})

	err := check()
	require.NoError(t, err)
}
