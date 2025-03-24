package grpc

import (
	"context"
	"errors"
	"log/slog"
	"time"

	metricspb "github.com/mclyashko/monitoring-system/services/metrics-collector/adapters/grpc/proto"
	"github.com/mclyashko/monitoring-system/services/metrics-collector/core"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	metricspb.UnimplementedMetricsCollectorServer
	log     *slog.Logger
	service *core.MetricService
}

func NewServer(log *slog.Logger, service *core.MetricService) *Server {
	return &Server{log: log, service: service}
}

func (s *Server) Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	s.log.Info("Pong")
	return &emptypb.Empty{}, nil
}

func (s *Server) SendMetric(_ context.Context, req *metricspb.SendMetricRequest) (*metricspb.SendMetricResponse, error) {
	metric := core.Metric{
		MetricIdentity: core.MetricIdentity{
			Time:       time.Now().UTC(),
			ServiceURL: req.ServiceUrl,
			MetricName: req.MetricName,
			PodName:    req.PodName,
		},
		MetricValue: req.MetricValue,
	}

	identity, err := s.service.CreateMetric(metric)
	if err != nil {
		switch {
		case errors.Is(err, core.ErrInvalidMetric):
			s.log.Warn("metric validation failed", slog.String("error", err.Error()))
			return nil, status.Errorf(codes.InvalidArgument, "metric validation failed")
		case errors.Is(err, core.ErrSaveFailed):
			s.log.Error("failed to save metric", slog.String("error", err.Error()))
			return nil, status.Errorf(codes.Internal, "failed to save metric")
		default:
			s.log.Error("unexpected error", slog.String("error", err.Error()))
			return nil, status.Errorf(codes.Internal, "unexpected error")
		}
	}

	response := metricspb.SendMetricResponse{
		Time:       timestamppb.New(identity.Time),
		ServiceUrl: identity.ServiceURL,
		MetricName: identity.MetricName,
		PodName:    identity.PodName,
	}

	return &response, nil
}
