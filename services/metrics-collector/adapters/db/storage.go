package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mclyashko/monitoring-system/services/metrics-collector/config"
	"github.com/mclyashko/monitoring-system/services/metrics-collector/core"
)

type DB struct {
	log  *slog.Logger
	pool *pgxpool.Pool
}

func New(log *slog.Logger, cfgDb *config.DB) (*DB, error) {
	config, err := pgxpool.ParseConfig(cfgDb.DBConnString)
	if err != nil {
		log.Error("unable to parse connection string", slog.String("error", err.Error()))
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}

	config.MinConns = cfgDb.PoolMinConns

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Error("unable to connect to database", slog.String("error", err.Error()))
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return &DB{
		log:  log,
		pool: pool,
	}, nil
}

func (db *DB) Save(metric core.Metric) (*core.MetricIdentity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var metricIdentity core.MetricIdentity
	query := `
		INSERT INTO metric (time, service_url, metric_name, pod_name, metric_value) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING time, service_url, metric_name, pod_name
	`
	err := db.pool.
		QueryRow(ctx, query, metric.Time, metric.ServiceURL, metric.MetricName, metric.PodName, metric.MetricValue).
		Scan(&metricIdentity.Time, &metricIdentity.ServiceURL, &metricIdentity.MetricName, &metricIdentity.PodName)
	if err != nil {
		db.log.Error("failed to insert metric", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to insert metric: %w", err)
	}

	db.log.Info("metric saved successfully", slog.Any("metric_identity", metricIdentity))
	return &metricIdentity, nil
}

func (db *DB) FindByMetricIdentity(metricIdentity core.MetricIdentity) (*core.Metric, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var metric core.Metric
	query := `
		SELECT time, service_url, metric_name, pod_name, metric_value
		FROM metric 
		WHERE time = $1 AND service_url = $2 AND metric_name = $3 AND pod_name = $4
	`
	err := db.pool.
		QueryRow(ctx, query, metricIdentity.Time, metricIdentity.ServiceURL, metricIdentity.MetricName, metricIdentity.PodName).
		Scan(&metric.Time, &metric.ServiceURL, &metric.MetricName, &metric.PodName, &metric.MetricValue)
	if err != nil {
		if err == pgx.ErrNoRows {
			db.log.Warn("metric not found", slog.Any("metric_identity", metricIdentity))
			return nil, fmt.Errorf("metric with identity %v not found", metricIdentity)
		}
		db.log.Error("failed to fetch metric", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to fetch metric: %w", err)
	}

	db.log.Info("metric found successfully", slog.Any("metric_identity", metricIdentity))
	return &metric, nil
}
