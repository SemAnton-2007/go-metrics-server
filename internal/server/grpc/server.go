package grpc // Исправленное имя пакета

import (
	"context"
	"log"
	"net"

	"go-metrics-server/internal/models"
	"go-metrics-server/internal/server/config"
	"go-metrics-server/internal/server/repository"
	"go-metrics-server/internal/server/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "go-metrics-server/internal/proto"
)

type MetricsServer struct {
	pb.UnimplementedMetricsServiceServer
	service *service.MetricService
}

func NewMetricsServer(service *service.MetricService) *MetricsServer {
	return &MetricsServer{service: service}
}

func (s *MetricsServer) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	var metrics []models.Metrics
	for _, m := range req.Metrics {
		metric := models.Metrics{
			ID:    m.Id,
			MType: m.Type,
		}

		switch m.Type {
		case "counter":
			delta := m.Delta
			metric.Delta = &delta
		case "gauge":
			value := m.Value
			metric.Value = &value
		}
		metrics = append(metrics, metric)
	}

	if err := s.service.UpdateMetrics(ctx, metrics); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update metrics: %v", err)
	}

	return &pb.UpdateMetricsResponse{}, nil
}

func (s *MetricsServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{Success: true}, nil
}

func Start(cfg *config.Config, repo repository.MetricRepository) {
	metricService := service.NewMetricService(repo)
	server := NewMetricsServer(metricService)

	lis, err := net.Listen("tcp", cfg.GRPCAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMetricsServiceServer(grpcServer, server)

	log.Printf("gRPC server running on %s", cfg.GRPCAddress)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
