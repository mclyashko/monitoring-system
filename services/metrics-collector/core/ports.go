package core

type MetricRepository interface {
	Save(metric Metric) (*MetricIdentity, error)
	FindByMetricIdentity(metricIdentity MetricIdentity) (*Metric, error)
}
