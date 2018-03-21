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

func TestConfig(t *testing.T) {
	conf := &Config{
		DSN: os.Getenv(mqDSNEnv),
	}

	conf.defaults()

	if conf.Exchange != defaultExchange {
		t.Fatal("Invalid default conf exchange value")
	}

	if conf.ConsumeTimeout != defaultConsumeTimeout {
		t.Fatal("Invalid default conf exchange value")
	}
}
