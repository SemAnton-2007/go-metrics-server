package grpc

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	pb "go-metrics-server/internal/proto"
	"go-metrics-server/internal/server/repository"
	"go-metrics-server/internal/server/service"
)

func startTestServer(t *testing.T) (*grpc.Server, *bufconn.Listener) {
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()

	repo := repository.NewMemoryRepository()
	service := service.NewMetricService(repo)
	pb.RegisterMetricsServiceServer(srv, NewMetricsServer(service))

	go func() {
		if err := srv.Serve(lis); err != nil {
			t.Logf("Server exited: %v", err)
		}
	}()

	t.Cleanup(func() {
		srv.Stop()
	})

	return srv, lis
}

func createTestClient(t *testing.T, lis *bufconn.Listener) pb.MetricsServiceClient {
	conn, err := grpc.NewClient(
		"",
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	return pb.NewMetricsServiceClient(conn)
}

func TestGRPCServer(t *testing.T) {
	_, lis := startTestServer(t)
	client := createTestClient(t, lis)

	t.Run("Ping", func(t *testing.T) {
		resp, err := client.Ping(context.Background(), &pb.PingRequest{})
		require.NoError(t, err)
		require.True(t, resp.GetSuccess())
	})

	t.Run("UpdateMetrics", func(t *testing.T) {
		req := &pb.UpdateMetricsRequest{
			Metrics: []*pb.Metric{
				{Id: "test", Type: "gauge", Value: 1.23},
				{Id: "counter", Type: "counter", Delta: 10},
			},
		}

		resp, err := client.UpdateMetrics(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp)
	})
}
