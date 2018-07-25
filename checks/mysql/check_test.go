package mysql

import (
	"os"
	"testing"
)

const mysqlDSNEnv = "HEALTH_GO_MS_DSN"

func TestNew(t *testing.T) {
	if os.Getenv(mysqlDSNEnv) == "" {
		t.SkipNow()
	}

	check := New(Config{
		DSN: os.Getenv(mysqlDSNEnv),
	})

	if err := check(); err != nil {
		t.Fatalf("MySQL check failed: %s", err.Error())
	}
}
