package core

import "time"

type MetricIdentity struct {
	Time       time.Time
	ServiceURL string
	MetricName string
	PodName    string
}

type Metric struct {
	MetricIdentity
	MetricValue float64
}
