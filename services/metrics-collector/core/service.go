package core

import "log/slog"

type MetricService struct {
	log  *slog.Logger
	repo MetricRepository
}

func NewMetricService(log *slog.Logger, repo MetricRepository) *MetricService {
	return &MetricService{
		log:  log,
		repo: repo,
	}
}

func (s *MetricService) CreateMetric(metric Metric) (*MetricIdentity, error) {
	if metric.ServiceURL == "" || metric.PodName == "" || metric.MetricName == "" {
		s.log.Warn("validation failed for metric, missing required params", slog.Any("metric", metric))
		return nil, ErrInvalidMetric
	}

	metricIdentity, err := s.repo.Save(metric)
	if err != nil {
		s.log.Error("failed to save metric", slog.String("error", err.Error()))
		return nil, ErrSaveFailed
	}

	s.log.Info("metric successfully created", slog.Any("metric_identity", *metricIdentity))
	return metricIdentity, nil
}

func (s *MetricService) GetMetricByMetricIdentity(metricIdentity MetricIdentity) (*Metric, error) {
	if metricIdentity.ServiceURL == "" || metricIdentity.PodName == "" || metricIdentity.MetricName == "" {
		s.log.Warn("invalid metric identity", slog.Any("metric_identity", metricIdentity))
		return nil, ErrInvalidMetricIdentity
	}

	metric, err := s.repo.FindByMetricIdentity(metricIdentity)
	if err != nil {
		s.log.Error("failed to find metric", slog.String("error", err.Error()))
		return nil, ErrMetricNotFound
	}

	s.log.Info("metric successfully retrieved", slog.Any("metric_identity", metricIdentity))
	return metric, nil
}
