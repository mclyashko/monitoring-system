package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/mclyashko/monitoring-system/services/metrics-collector/adapters/db"
	"github.com/mclyashko/monitoring-system/services/metrics-collector/adapters/rest"
	"github.com/mclyashko/monitoring-system/services/metrics-collector/config"
	"github.com/mclyashko/monitoring-system/services/metrics-collector/core"
)

func main() {
	cfg := mustLoadConfig()

	log := mustMakeLogger(cfg.LogLevel)

	greetings(log)

	storage := mustMakeStorage(log, &cfg.DB)

	mustMakeMigrations(log, storage, cfg.DB.DBConnString)

	mux := mustMakeMux(log, storage)

	server := &http.Server{
		Addr:        cfg.AppAddress,
		ReadTimeout: cfg.ReadTimeout,
		Handler:     mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Debug("shutting down server")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Error("erroneous shutdown", slog.String("error", err.Error()))
		}
	}()

	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("server closed unexpectedly", slog.String("error", err.Error()))
			return
		}
	}
}

func mustLoadConfig() *config.Config {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()

	cfg := config.MustLoad(configPath)

	return cfg
}

func mustMakeLogger(logLevel string) *slog.Logger {
	var level slog.Level
	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}

func greetings(log *slog.Logger) {
	log.Info("starting server")
	log.Debug("debug messages are enabled")
}

func mustMakeStorage(log *slog.Logger, cfgDb *config.DB) *db.DB {
	log.Info("connecting to the database...")

	storage, err := db.New(log, cfgDb)
	if err != nil {
		log.Error("failed to initialize storage", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log.Info("database connection established successfully")

	return storage
}

func mustMakeMigrations(log *slog.Logger, db *db.DB, connString string) {
	if err := db.Migrate(connString); err != nil {
		log.Error("failed to apply migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func mustMakeMux(log *slog.Logger, repo core.MetricRepository) *http.ServeMux {
	mux := http.NewServeMux()

	metricService := core.NewMetricService(log, repo)

	mux.HandleFunc("GET /", rest.NewPingHandler(log))
	mux.HandleFunc("GET /metric", rest.NewGetMetricByMetricIdentityHandler(log, metricService))
	mux.HandleFunc("POST /metric", rest.NewCreateMetricHandler(log, metricService))

	log.Info("mux initialized with routes")

	return mux
}
