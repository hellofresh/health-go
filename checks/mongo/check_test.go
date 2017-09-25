package mongo

import (
	"os"
	"testing"
)

const mgDSNEnv = "HEALTH_GO_MG_DSN"

func TestNew(t *testing.T) {
	if os.Getenv(mgDSNEnv) == "" {
		t.SkipNow()
	}

	check := New(Config{
		DSN: os.Getenv(mgDSNEnv),
	})

	if err := check(); err != nil {
		t.Fatalf("MongoDB check failed: %s", err.Error())
	}
}
