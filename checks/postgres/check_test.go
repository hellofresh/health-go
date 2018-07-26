package postgres

import (
	"os"
	"testing"
)

const pgDSNEnv = "HEALTH_GO_PG_DSN"

func TestNew(t *testing.T) {
	if os.Getenv(pgDSNEnv) == "" {
		t.SkipNow()
	}

	check := New(Config{
		DSN: os.Getenv(pgDSNEnv),
	})

	if err := check(); err != nil {
		t.Fatalf("PostgreSQL check failed: %s", err.Error())
	}
}
