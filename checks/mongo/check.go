package mongo

import (
	"context"
	"time"

	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/readpref"
)

// Config is the MongoDB checker configuration settings container.
type Config struct {
	// DSN is the MongoDB instance connection DSN. Required.
	DSN string
	// LogFunc is the callback function for errors logging during check.
	// If not set logging is skipped.
	LogFunc func(err error, details string, extra ...interface{})
}

// New creates new MongoDB health check that verifies the following:
// - connection establishing
// - doing the ping command
func New(config Config) func() error {
	if config.LogFunc == nil {
		config.LogFunc = func(err error, details string, extra ...interface{}) {}
	}

	return func() error {
		var ctx context.Context
		var cancel context.CancelFunc

		client, err := mongo.NewClient(config.DSN)
		if err != nil {
			config.LogFunc(err, "MongoDB health check failed on client creation")
			return err
		}

		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = client.Connect(ctx)

		if err != nil {
			config.LogFunc(err, "MongoDB health check failed on connect")
			return err
		}

		defer func() {
			ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := client.Disconnect(ctx); err != nil {
				config.LogFunc(err, "MongoDB health check failed on closing connection")
			}
		}()

		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = client.Ping(ctx, readpref.Primary())

		if err != nil {
			config.LogFunc(err, "MongoDB health check failed during ping")
			return err
		}

		return nil
	}
}
