package http

import (
	"os"
	"testing"
)

const httpURLEnv = "HEALTH_GO_HTTP_URL"

func TestNew(t *testing.T) {
	if os.Getenv(httpURLEnv) == "" {
		t.SkipNow()
	}

	check := New(Config{
		URL: os.Getenv(httpURLEnv),
	})

	if err := check(); err != nil {
		t.Fatalf("HTTP check failed: %s", err.Error())
	}
}
