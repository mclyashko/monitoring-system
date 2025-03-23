package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mclyashko/monitoring-system/services/test-service-go/config"
	"github.com/mclyashko/monitoring-system/services/test-service-go/core"
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

func (db *DB) Save(order core.Order) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var id int
	query := `INSERT INTO "order" (product_id, quantity, user_id) VALUES ($1, $2, $3) RETURNING id`
	err := db.pool.QueryRow(ctx, query, order.ProductID, order.Quantity, order.UserID).Scan(&id)
	if err != nil {
		db.log.Error("failed to insert order", slog.String("error", err.Error()))
		return 0, fmt.Errorf("failed to insert order: %w", err)
	}

	db.log.Info("order saved successfully", slog.Int("order_id", id))
	return id, nil
}

func (db *DB) FindByID(orderID int) (*core.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var order core.Order
	query := `SELECT id, product_id, quantity, user_id FROM "order" WHERE id = $1`
	err := db.pool.QueryRow(ctx, query, orderID).Scan(&order.ID, &order.ProductID, &order.Quantity, &order.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			db.log.Warn("order not found", slog.Int("order_id", orderID))
			return nil, fmt.Errorf("order with id %d not found", orderID)
		}
		db.log.Error("failed to fetch order", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to fetch order: %w", err)
	}

	db.log.Info("order found successfully", slog.Int("order_id", order.ID))
	return &order, nil
}
