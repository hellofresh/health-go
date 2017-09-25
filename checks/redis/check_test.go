package redis

import (
	"os"
	"testing"
)

const rdDSNEnv = "HEALTH_GO_RD_DSN"

func TestNew(t *testing.T) {
	if os.Getenv(rdDSNEnv) == "" {
		t.SkipNow()
	}

	check := New(Config{
		DSN: os.Getenv(rdDSNEnv),
	})

	if err := check(); err != nil {
		t.Fatalf("Redis check failed: %s", err.Error())
	}
}
