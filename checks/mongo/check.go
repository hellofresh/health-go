package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	defaultTimeoutConnect    = 5 * time.Second
	defaultTimeoutDisconnect = 5 * time.Second
	defaultTimeoutPing       = 5 * time.Second
)

// Config is the MongoDB checker configuration settings container.
type Config struct {
	// DSN is the MongoDB instance connection DSN. Required.
	DSN string
	// LogFunc is the callback function for errors logging during check.
	// If not set logging is skipped.
	LogFunc func(err error, details string, extra ...interface{})

	// TimeoutConnect defines timeout for establishing mongo connection, if not set - default value is used
	TimeoutConnect time.Duration
	// TimeoutDisconnect defines timeout for closing connection, if not set - default value is used
	TimeoutDisconnect time.Duration
	// TimeoutDisconnect defines timeout for making ping request, if not set - default value is used
	TimeoutPing time.Duration
}

// New creates new MongoDB health check that verifies the following:
// - connection establishing
// - doing the ping command
func New(config Config) func() error {
	if config.LogFunc == nil {
		config.LogFunc = func(err error, details string, extra ...interface{}) {}
	}

	if config.TimeoutConnect == 0 {
		config.TimeoutConnect = defaultTimeoutConnect
	}

	if config.TimeoutDisconnect == 0 {
		config.TimeoutDisconnect = defaultTimeoutDisconnect
	}

	if config.TimeoutPing == 0 {
		config.TimeoutPing = defaultTimeoutPing
	}

	return func() error {
		var ctx context.Context
		var cancel context.CancelFunc

		client, err := mongo.NewClient(options.Client().ApplyURI(config.DSN))
		if err != nil {
			config.LogFunc(err, "MongoDB health check failed on client creation")
			return err
		}

		ctx, cancel = context.WithTimeout(context.Background(), config.TimeoutConnect)
		defer cancel()
		err = client.Connect(ctx)

		if err != nil {
			config.LogFunc(err, "MongoDB health check failed on connect")
			return err
		}

		defer func() {
			ctx, cancel = context.WithTimeout(context.Background(), config.TimeoutDisconnect)
			defer cancel()
			if err := client.Disconnect(ctx); err != nil {
				config.LogFunc(err, "MongoDB health check failed on closing connection")
			}
		}()

		ctx, cancel = context.WithTimeout(context.Background(), config.TimeoutPing)
		defer cancel()
		err = client.Ping(ctx, readpref.Primary())

		if err != nil {
			config.LogFunc(err, "MongoDB health check failed during ping")
			return err
		}

		return nil
	}
}
