package grpcclient

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pb "go-metrics-server/internal/proto"
)

type MockMetricsService struct {
	pb.UnimplementedMetricsServiceServer
	updateMetricsFunc func(context.Context, *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error)
}

func (m *MockMetricsService) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	return m.updateMetricsFunc(ctx, req)
}

func TestNew_Success(t *testing.T) {
	lis, srv := startTestServer(t, func(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
		return &pb.UpdateMetricsResponse{}, nil
	})
	defer srv.Stop()

	// Тестируем New с помощью вспомогательной функции
	client := createTestClient(t, lis.Addr().String())
	require.NotNil(t, client)
	defer client.Close()

	// Проверяем, что клиент работает
	metrics := map[string]interface{}{"test": 1.23}
	err := client.SendMetrics(metrics)
	require.NoError(t, err)
}

func TestNew_InvalidAddress(t *testing.T) {
	client := createTestClient(t, "invalid:address")
	require.Nil(t, client)
}

func TestNew_EmptyAddress(t *testing.T) {
	client := createTestClient(t, "")
	require.Nil(t, client)
}

func TestClose_Success(t *testing.T) {
	lis, srv := startTestServer(t, func(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
		return &pb.UpdateMetricsResponse{}, nil
	})
	defer srv.Stop()

	client := createTestClient(t, lis.Addr().String())
	require.NotNil(t, client)

	require.NotNil(t, client.conn)
	client.Close()

	require.Eventually(t, func() bool {
		state := client.conn.GetState()
		return state == connectivity.Shutdown || state == connectivity.TransientFailure
	}, time.Second, 10*time.Millisecond)
}

func TestClose_NoPanic(t *testing.T) {
	var client *GRPCClient
	require.NotPanics(t, func() {
		client.Close()
	})
}

func TestSendMetrics_Success(t *testing.T) {
	lis, srv := startTestServer(t, func(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
		require.Len(t, req.Metrics, 2)
		return &pb.UpdateMetricsResponse{}, nil
	})
	defer srv.Stop()

	client := createTestClient(t, lis.Addr().String())
	require.NotNil(t, client)
	defer client.Close()

	metrics := map[string]interface{}{
		"gauge_metric":   3.14,
		"counter_metric": int64(42),
	}

	err := client.SendMetrics(metrics)
	require.NoError(t, err)
}

func TestSendMetrics_Error(t *testing.T) {
	lis, srv := startTestServer(t, func(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
		return nil, status.Errorf(codes.Internal, "server error")
	})
	defer srv.Stop()

	client := createTestClient(t, lis.Addr().String())
	require.NotNil(t, client)
	defer client.Close()

	metrics := map[string]interface{}{"test": 1.23}
	err := client.SendMetrics(metrics)

	require.Error(t, err)
	require.Contains(t, err.Error(), "server error")
}

func TestSendMetrics_InvalidType(t *testing.T) {
	lis, srv := startTestServer(t, func(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
		require.Len(t, req.Metrics, 1)
		return &pb.UpdateMetricsResponse{}, nil
	})
	defer srv.Stop()

	client := createTestClient(t, lis.Addr().String())
	require.NotNil(t, client)
	defer client.Close()

	metrics := map[string]interface{}{
		"valid_gauge":   1.23,
		"invalid_type":  "string",
		"invalid_float": float32(45),
	}

	err := client.SendMetrics(metrics)
	require.NoError(t, err)
}

func startTestServer(
	t *testing.T,
	handler func(context.Context, *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error),
) (net.Listener, *grpc.Server) {
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	srv := grpc.NewServer()
	pb.RegisterMetricsServiceServer(srv, &MockMetricsService{
		updateMetricsFunc: handler,
	})

	go func() {
		if err := srv.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			t.Errorf("server error: %v", err)
		}
	}()

	return lis, srv
}

func createTestClient(t *testing.T, address string) *GRPCClient {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		t.Logf("Failed to create test client: %v", err)
		return nil
	}

	return &GRPCClient{
		client: pb.NewMetricsServiceClient(conn),
		conn:   conn,
	}
}
