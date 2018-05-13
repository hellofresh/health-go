package grpc

import (
	"log"
	"net"
	"os"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

const (
	addr    = ":8080"
	service = "HealthTest"
)

var healthServer *health.Server

func TestMain(m *testing.M) {
	healthServer = health.NewServer()
	healthServer.SetServingStatus(service, healthpb.HealthCheckResponse_SERVING)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("error setting up tcp listener: %v", err)
	}

	server := grpc.NewServer()
	healthpb.RegisterHealthServer(server, healthServer)

	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	defer server.Stop()

	os.Exit(m.Run())
}

func TestNew_WithServingStatusServing(t *testing.T) {
	healthServer.SetServingStatus(service, healthpb.HealthCheckResponse_SERVING)

	check := New(Config{
		Target:  addr,
		Service: service,
		DialOptions: []grpc.DialOption{
			grpc.WithInsecure(),
		},
	})

	if err := check(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestNew_WithServingStatusUnknown(t *testing.T) {
	healthServer.SetServingStatus(service, healthpb.HealthCheckResponse_UNKNOWN)

	check := New(Config{
		Target:  addr,
		Service: service,
		DialOptions: []grpc.DialOption{
			grpc.WithInsecure(),
		},
	})

	if err := check(); err != errStatusUnhealthy {
		t.Fatalf("expected error: %v", errStatusUnhealthy)
	}
}

func TestNew_WithServingStatusNotServing(t *testing.T) {
	healthServer.SetServingStatus(service, healthpb.HealthCheckResponse_NOT_SERVING)

	check := New(Config{
		Target:  addr,
		Service: service,
		DialOptions: []grpc.DialOption{
			grpc.WithInsecure(),
		},
	})

	if err := check(); err != errStatusUnhealthy {
		t.Fatalf("expected error: %v", errStatusUnhealthy)
	}
}
