package rabbitmq

import (
	"os"
	"testing"
)

const mqDSNEnv = "HEALTH_GO_MQ_DSN"

func TestNew(t *testing.T) {
	if os.Getenv(mqDSNEnv) == "" {
		t.SkipNow()
	}

	check := New(Config{
		DSN: os.Getenv(mqDSNEnv),
	})

	if err := check(); err != nil {
		t.Fatalf("RabbitMQ check failed: %s", err.Error())
	}
}
