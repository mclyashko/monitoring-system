package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mclyashko/monitoring-system/services/test-service-go/config"
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
