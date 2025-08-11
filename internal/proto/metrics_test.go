package proto

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMessages(t *testing.T) {
	t.Run("UpdateMetricsRequest", func(t *testing.T) {
		req := &UpdateMetricsRequest{
			Metrics: []*Metric{{Id: "test"}},
		}

		assert.Len(t, req.GetMetrics(), 1)

		req.Reset()
		assert.Nil(t, req.GetMetrics())

		assert.NotNil(t, req.ProtoReflect())
	})

	t.Run("PingResponse", func(t *testing.T) {
		resp := &PingResponse{Success: true}

		assert.True(t, resp.GetSuccess())

		resp.Reset()
		assert.False(t, resp.GetSuccess())

		assert.NotNil(t, resp.ProtoReflect())
	})
}

func TestGRPCClient(t *testing.T) {
	t.Run("NewClient", func(t *testing.T) {
		mockConn := &mockClientConn{}
		client := NewMetricsServiceClient(mockConn)
		assert.NotNil(t, client)
	})

	t.Run("ClientMethods", func(t *testing.T) {
		calls := 0
		mockConn := &mockClientConn{
			invokeHandler: func(ctx context.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error {
				calls++
				switch method {
				case MetricsService_UpdateMetrics_FullMethodName:
					*reply.(*UpdateMetricsResponse) = UpdateMetricsResponse{}
				case MetricsService_Ping_FullMethodName:
					*reply.(*PingResponse) = PingResponse{Success: true}
				}
				return nil
			},
		}

		client := NewMetricsServiceClient(mockConn)

		t.Run("UpdateMetrics", func(t *testing.T) {
			resp, err := client.UpdateMetrics(context.Background(), &UpdateMetricsRequest{
				Metrics: []*Metric{{Id: "test"}},
			})
			assert.NoError(t, err)
			assert.NotNil(t, resp)
		})

		t.Run("Ping", func(t *testing.T) {
			resp, err := client.Ping(context.Background(), &PingRequest{})
			assert.NoError(t, err)
			assert.True(t, resp.Success)
		})

		assert.Equal(t, 2, calls)
	})
}

func TestGRPCServer(t *testing.T) {
	t.Run("UnimplementedServer", func(t *testing.T) {
		srv := &UnimplementedMetricsServiceServer{}
		ctx := context.Background()

		t.Run("UpdateMetrics", func(t *testing.T) {
			resp, err := srv.UpdateMetrics(ctx, &UpdateMetricsRequest{})
			assert.Nil(t, resp)
			assert.Error(t, err)
			assert.Equal(t, codes.Unimplemented, status.Code(err))
		})

		t.Run("Ping", func(t *testing.T) {
			resp, err := srv.Ping(ctx, &PingRequest{})
			assert.Nil(t, resp)
			assert.Error(t, err)
			assert.Equal(t, codes.Unimplemented, status.Code(err))
		})
	})

	t.Run("ImplementedServer", func(t *testing.T) {
		srv := &mockMetricsServer{
			updateMetricsFunc: func(ctx context.Context, req *UpdateMetricsRequest) (*UpdateMetricsResponse, error) {
				assert.NotNil(t, req)
				return &UpdateMetricsResponse{}, nil
			},
			pingFunc: func(ctx context.Context, req *PingRequest) (*PingResponse, error) {
				return &PingResponse{Success: true}, nil
			},
		}

		t.Run("UpdateMetrics", func(t *testing.T) {
			resp, err := srv.UpdateMetrics(context.Background(), &UpdateMetricsRequest{})
			assert.NoError(t, err)
			assert.NotNil(t, resp)
		})

		t.Run("Ping", func(t *testing.T) {
			resp, err := srv.Ping(context.Background(), &PingRequest{})
			assert.NoError(t, err)
			assert.True(t, resp.Success)
		})
	})

	t.Run("Registration", func(t *testing.T) {
		s := grpc.NewServer()
		RegisterMetricsServiceServer(s, &mockMetricsServer{})

		info := s.GetServiceInfo()
		require.Contains(t, info, "proto.MetricsService")
		assert.Len(t, info["proto.MetricsService"].Methods, 2)
	})
}

func TestHelpers(t *testing.T) {
	t.Run("FileDescriptor", func(t *testing.T) {
		assert.NotEmpty(t, file_proto_metrics_proto_rawDesc)
		assert.NotEmpty(t, file_proto_metrics_proto_rawDescGZIP())
	})

	t.Run("MethodNames", func(t *testing.T) {
		assert.Equal(t, "/proto.MetricsService/UpdateMetrics", MetricsService_UpdateMetrics_FullMethodName)
		assert.Equal(t, "/proto.MetricsService/Ping", MetricsService_Ping_FullMethodName)
	})
}

type mockClientConn struct {
	grpc.ClientConnInterface
	invokeHandler func(ctx context.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error
}

func (m *mockClientConn) Invoke(ctx context.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error {
	if m.invokeHandler != nil {
		return m.invokeHandler(ctx, method, args, reply, opts...)
	}
	return nil
}

type mockMetricsServer struct {
	UnimplementedMetricsServiceServer
	updateMetricsFunc func(context.Context, *UpdateMetricsRequest) (*UpdateMetricsResponse, error)
	pingFunc          func(context.Context, *PingRequest) (*PingResponse, error)
}

func (m *mockMetricsServer) UpdateMetrics(ctx context.Context, req *UpdateMetricsRequest) (*UpdateMetricsResponse, error) {
	if m.updateMetricsFunc != nil {
		return m.updateMetricsFunc(ctx, req)
	}
	return &UpdateMetricsResponse{}, nil
}

func (m *mockMetricsServer) Ping(ctx context.Context, req *PingRequest) (*PingResponse, error) {
	if m.pingFunc != nil {
		return m.pingFunc(ctx, req)
	}
	return &PingResponse{Success: true}, nil
}

func (m *mockMetricsServer) mustEmbedUnimplementedMetricsServiceServer() {}
