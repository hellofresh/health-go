package postgres

import (
	"os"
	"testing"
	"time"
)

const pgDSNEnv = "HEALTH_GO_PG_DSN"

func TestNew(t *testing.T) {
	if os.Getenv(pgDSNEnv) == "" {
		t.SkipNow()
	}

	check := New(Config{
		DSN:      os.Getenv(pgDSNEnv),
		Table:    "client",
		IDColumn: "id",
		InsertColumnsFunc: func() map[string]interface{} {
			return map[string]interface{}{
				"id":           time.Now().Format(time.RFC3339Nano),
				"secret":       time.Now().Format(time.RFC3339Nano),
				"extra":        time.Now().Format(time.RFC3339Nano),
				"redirect_uri": "http://localhost",
			}
		},
	})

	if err := check(); err != nil {
		t.Fatalf("PostgreSQL check failed: %s", err.Error())
	}
}
