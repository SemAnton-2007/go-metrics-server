package grpcclient

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	"go-metrics-server/internal/agent/config"
	pb "go-metrics-server/internal/proto"
)

type GRPCClient struct {
	client pb.MetricsServiceClient
	conn   *grpc.ClientConn
}

func New(cfg *config.Config) *GRPCClient {
	conn, err := grpc.NewClient(
		cfg.GRPCAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("Failed to create gRPC client: %v", err)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if !waitForReady(ctx, conn) {
		log.Printf("Failed to establish connection to %s", cfg.GRPCAddress)
		return nil
	}

	return &GRPCClient{
		client: pb.NewMetricsServiceClient(conn),
		conn:   conn,
	}
}

func waitForReady(ctx context.Context, conn *grpc.ClientConn) bool {
	for {
		select {
		case <-ctx.Done():
			return false
		default:
			if conn.WaitForStateChange(ctx, conn.GetState()) {
				if conn.GetState() == connectivity.Ready {
					return true
				}
			}
		}
	}
}

func (c *GRPCClient) SendMetrics(metrics map[string]interface{}) error {
	var pbMetrics []*pb.Metric
	for name, value := range metrics {
		metric := &pb.Metric{Id: name}
		switch v := value.(type) {
		case float64:
			metric.Type = "gauge"
			metric.Value = v
		case int64:
			metric.Type = "counter"
			metric.Delta = v
		default:
			continue
		}
		pbMetrics = append(pbMetrics, metric)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.client.UpdateMetrics(ctx, &pb.UpdateMetricsRequest{
		Metrics: pbMetrics,
	})
	return err
}

func (c *GRPCClient) Close() {
	if c == nil || c.conn == nil {
		return
	}
	c.conn.Close()
}
