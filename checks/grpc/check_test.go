package grpc

import (
	"context"
	"log"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

const (
	addr    = ":8080"
	service = "HealthTest"
)

var healthServer *health.Server

func TestMain(m *testing.M) {
	healthServer = health.NewServer()
	healthServer.SetServingStatus(service, grpc_health_v1.HealthCheckResponse_SERVING)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("error setting up tcp listener: %v", err)
	}

	server := grpc.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)

	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	defer server.Stop()

	os.Exit(m.Run())
}

func TestNew_WithServingStatusServing(t *testing.T) {
	healthServer.SetServingStatus(service, grpc_health_v1.HealthCheckResponse_SERVING)

	check := New(Config{
		Target:  addr,
		Service: service,
		DialOptions: []grpc.DialOption{
			grpc.WithInsecure(),
		},
	})

	err := check(context.Background())
	require.NoError(t, err)
}

func TestNew_WithServingStatusUnknown(t *testing.T) {
	healthServer.SetServingStatus(service, grpc_health_v1.HealthCheckResponse_UNKNOWN)

	check := New(Config{
		Target:  addr,
		Service: service,
		DialOptions: []grpc.DialOption{
			grpc.WithInsecure(),
		},
	})

	err := check(context.Background())
	require.Error(t, err)
}

func TestNew_WithServingStatusNotServing(t *testing.T) {
	healthServer.SetServingStatus(service, grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	check := New(Config{
		Target:  addr,
		Service: service,
		DialOptions: []grpc.DialOption{
			grpc.WithInsecure(),
		},
	})

	err := check(context.Background())
	require.Error(t, err)
}
