package mongo

import (
	"context"
	"time"

	wErrors "github.com/pkg/errors"
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
	if config.TimeoutConnect == 0 {
		config.TimeoutConnect = defaultTimeoutConnect
	}

	if config.TimeoutDisconnect == 0 {
		config.TimeoutDisconnect = defaultTimeoutDisconnect
	}

	if config.TimeoutPing == 0 {
		config.TimeoutPing = defaultTimeoutPing
	}

	return func() (checkErr error) {
		var ctx context.Context
		var cancel context.CancelFunc

		client, err := mongo.NewClient(options.Client().ApplyURI(config.DSN))
		if err != nil {
			checkErr = wErrors.Wrap(err, "mongoDB health check failed on client creation")
			return
		}

		ctx, cancel = context.WithTimeout(context.Background(), config.TimeoutConnect)
		defer cancel()

		err = client.Connect(ctx)
		if err != nil {
			checkErr = wErrors.Wrap(err, "mongoDB health check failed on connect")
			return
		}

		defer func() {
			ctx, cancel = context.WithTimeout(context.Background(), config.TimeoutDisconnect)
			defer cancel()

			// override checkErr only if there were no other errors
			if err := client.Disconnect(ctx); err != nil && checkErr == nil {
				checkErr = wErrors.Wrap(err, "mongoDB health check failed on closing connection")
			}
		}()

		ctx, cancel = context.WithTimeout(context.Background(), config.TimeoutPing)
		defer cancel()

		err = client.Ping(ctx, readpref.Primary())
		if err != nil {
			checkErr = wErrors.Wrap(err, "mongoDB health check failed on ping")
			return
		}

		return
	}
}
