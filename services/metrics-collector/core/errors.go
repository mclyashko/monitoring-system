package core

import "errors"

var (
	ErrInvalidMetric = errors.New("invalid metric: no required params")
	ErrSaveFailed    = errors.New("failed to save metric")
)

var (
	ErrInvalidMetricIdentity = errors.New("invalid metric identity")
	ErrMetricNotFound        = errors.New("metric not found")
)
