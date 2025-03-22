package db

import (
	"embed"
	"log/slog"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func (d *DB) Migrate(connString string) error {
	d.log.Info("running migrations...")

	connString = strings.Replace(connString, "postgres://", "pgx5://", 1)

	migrationsSource, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		d.log.Error("failed to load migrations source", slog.String("error", err.Error()))
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", migrationsSource, connString)
	if err != nil {
		d.log.Error("failed to initialize migration instance", slog.String("error", err.Error()))
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		d.log.Error("failed to apply migrations", slog.String("error", err.Error()))
		return err
	} else if err == migrate.ErrNoChange {
		d.log.Info("no new migrations to apply")
	}

	d.log.Info("migrations applied successfully")

	return nil
}
