package mysql

import (
	"os"
	"testing"
	"time"
)

const mysqlDSNEnv = "HEALTH_GO_MS_DSN"

func TestNew(t *testing.T) {
	if os.Getenv(mysqlDSNEnv) == "" {
		t.SkipNow()
	}

	check := New(Config{
		DSN:      os.Getenv(mysqlDSNEnv),
		Table:    "test",
		IDColumn: "id",
		InsertColumnsFunc: func() map[string]interface{} {
			return map[string]interface{}{
				"id":           nil,
				"secret":       time.Now().Format(time.RFC3339Nano),
				"extra":        time.Now().Format(time.RFC3339Nano),
				"redirect_uri": "http://localhost",
			}
		},
	})

	if err := check(); err != nil {
		t.Fatalf("MySQL check failed: %s", err.Error())
	}
}
