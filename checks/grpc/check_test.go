package grpc

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func TestNew(t *testing.T) {
	for name, tc := range map[string]struct {
		servingStatus grpc_health_v1.HealthCheckResponse_ServingStatus
		requireError  bool
	}{
		"serving": {
			servingStatus: grpc_health_v1.HealthCheckResponse_SERVING,
			requireError:  false,
		},
		"unknown": {
			servingStatus: grpc_health_v1.HealthCheckResponse_UNKNOWN,
			requireError:  true,
		},
		"not serving": {
			servingStatus: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
			requireError:  true,
		},
	} {
		servingStatus := tc.servingStatus
		requireError := tc.requireError

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			const service = "HealthTest"

			healthServer := health.NewServer()
			healthServer.SetServingStatus(service, servingStatus)

			lis, err := net.Listen("tcp", "localhost:0")
			require.NoError(t, err)

			server := grpc.NewServer()
			grpc_health_v1.RegisterHealthServer(server, healthServer)

			go func() {
				if err := server.Serve(lis); err != nil {
					t.Log("Failed to serve GRPC", err)
				}
			}()
			defer server.Stop()

			check := New(Config{
				Target:  lis.Addr().String(),
				Service: service,
				DialOptions: []grpc.DialOption{
					grpc.WithInsecure(),
				},
			})

			err = check(context.Background())

			if requireError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
