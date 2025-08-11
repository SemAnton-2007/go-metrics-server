package grpcclient

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go-metrics-server/internal/agent/config"
	pb "go-metrics-server/internal/proto"
)

type GRPCClient struct {
	client pb.MetricsServiceClient
	conn   *grpc.ClientConn
}

func New(cfg *config.Config) *GRPCClient {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		cfg.GRPCAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Printf("Failed to connect to gRPC server: %v", err)
		return nil
	}

	return &GRPCClient{
		client: pb.NewMetricsServiceClient(conn),
		conn:   conn,
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
