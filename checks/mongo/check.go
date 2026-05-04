package mongo

import (
	"context"
	"errors"
	"fmt"
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

var errNilClient = errors.New("mongoDB health check failed on ping: nil mongo client")

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

// NewPingCheck creates a MongoDB health check using an existing client.
// The caller is responsible for managing the client's lifecycle.
func NewPingCheck(client *mongo.Client, timeout time.Duration) func(ctx context.Context) error {
	if client == nil {
		return func(context.Context) error {
			return errNilClient
		}
	}

	if timeout == 0 {
		timeout = defaultTimeoutPing
	}

	return func(ctx context.Context) error {
		ctxPing, cancelPing := context.WithTimeout(ctx, timeout)
		defer cancelPing()

		err := client.Ping(ctxPing, readpref.Primary())
		if err != nil {
			return fmt.Errorf("mongoDB health check failed on ping: %w", err)
		}

		return nil
	}
}

// New creates new MongoDB health check that verifies the following:
// - connection establishing
// - doing the ping command
//
// Deprecated: use NewPingCheck instead. It uses an existing *mongo.Client
// to avoid connection churn.
func New(config Config) func(ctx context.Context) error {
	if config.TimeoutConnect == 0 {
		config.TimeoutConnect = defaultTimeoutConnect
	}

	if config.TimeoutDisconnect == 0 {
		config.TimeoutDisconnect = defaultTimeoutDisconnect
	}

	if config.TimeoutPing == 0 {
		config.TimeoutPing = defaultTimeoutPing
	}

	return func(ctx context.Context) (checkErr error) {
		client, err := mongo.NewClient(options.Client().ApplyURI(config.DSN))
		if err != nil {
			checkErr = fmt.Errorf("mongoDB health check failed on client creation: %w", err)
			return
		}

		ctxConn, cancelConn := context.WithTimeout(ctx, config.TimeoutConnect)
		defer cancelConn()

		err = client.Connect(ctxConn)
		if err != nil {
			checkErr = fmt.Errorf("mongoDB health check failed on connect: %w", err)
			return
		}

		defer func() {
			ctxDisc, cancelDisc := context.WithTimeout(ctx, config.TimeoutDisconnect)
			defer cancelDisc()

			// override checkErr only if there were no other errors
			if err := client.Disconnect(ctxDisc); err != nil && checkErr == nil {
				checkErr = fmt.Errorf("mongoDB health check failed on closing connection: %w", err)
			}
		}()

		ctxPing, cancelPing := context.WithTimeout(ctx, config.TimeoutPing)
		defer cancelPing()

		err = client.Ping(ctxPing, readpref.Primary())
		if err != nil {
			checkErr = fmt.Errorf("mongoDB health check failed on ping: %w", err)
			return
		}

		return
	}
}
