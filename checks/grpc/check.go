package grpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

const defaultCheckTimeout = 5 * time.Second

// Config is the gRPC checker configuration settings container.
type Config struct {
	// Target is the address of the gRPC server
	Target string
	// Service is the name of the gRPC service
	Service string
	// DialOptions configure how we set up the connection
	DialOptions []grpc.DialOption
	// CheckTimeout is the duration that health check will try to get gRPC service health status.
	// If not set - 5 seconds
	CheckTimeout time.Duration
}

// New creates new gRPC health check
func New(config Config) func(ctx context.Context) error {
	if config.CheckTimeout == 0 {
		config.CheckTimeout = defaultCheckTimeout
	}

	return func(ctx context.Context) error {
		// Set up a connection to the gRPC server
		conn, err := grpc.Dial(config.Target, config.DialOptions...)
		if err != nil {
			return fmt.Errorf("gRPC health check failed on connect: %w", err)
		}
		defer conn.Close()

		healthClient := grpc_health_v1.NewHealthClient(conn)

		ctx, cancel := context.WithTimeout(ctx, config.CheckTimeout)
		defer cancel()

		res, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{
			Service: config.Service,
		})
		if err != nil {
			return fmt.Errorf("gRPC health check failed on check call: %w", err)
		}

		if res.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
			return fmt.Errorf("gRPC service reported as non-serving: %q", res.GetStatus().String())
		}

		return nil
	}
}
