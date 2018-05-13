package grpc

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type Config struct {
	// Target is the address of the gRPC server
	Target string
	// Service is the name of the gRPC service
	Service string
	// DialOptions configure how we set up the connection
	DialOptions []grpc.DialOption
	// LogFunc is the callback function for errors logging during check.
	// If not set logging is skipped.
	LogFunc func(err error, details string, extra ...interface{})
}

var (
	errStatusUnhealthy = errors.New("remote service is not available at the moment")
	defaultLogFunc     = func(err error, details string, extra ...interface{}) {}
)

// New creates new gRPC health check
func New(config Config) func() error {
	if config.LogFunc == nil {
		config.LogFunc = defaultLogFunc
	}

	return func() error {
		// Set up a connection to the gRPC server
		conn, err := grpc.Dial(config.Target, config.DialOptions...)
		if err != nil {
			config.LogFunc(err, "gRPC health check failed during connect")
			return err
		}

		defer func() {
			if err = conn.Close(); err != nil {
				config.LogFunc(err, "gRPC health check failed during connection closing")
			}
		}()

		client := healthpb.NewHealthClient(conn)

		res, err := client.Check(context.Background(), &healthpb.HealthCheckRequest{
			Service: config.Service,
		})
		if err != nil {
			config.LogFunc(err, "gRPC health check failed")
			return err
		}

		if res.GetStatus() != healthpb.HealthCheckResponse_SERVING {
			config.LogFunc(errStatusUnhealthy, fmt.Sprintf(
				"gRPC health check response status %s is unhealthy",
				res.GetStatus(),
			))
			return errStatusUnhealthy
		}

		return nil
	}
}
