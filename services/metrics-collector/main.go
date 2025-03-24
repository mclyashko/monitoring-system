package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/mclyashko/monitoring-system/services/metrics-collector/adapters/db"
	"github.com/mclyashko/monitoring-system/services/metrics-collector/adapters/rest"
	"github.com/mclyashko/monitoring-system/services/metrics-collector/config"
	"github.com/mclyashko/monitoring-system/services/metrics-collector/core"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	metricsgrpc "github.com/mclyashko/monitoring-system/services/metrics-collector/adapters/grpc"
	metricspb "github.com/mclyashko/monitoring-system/services/metrics-collector/adapters/grpc/proto"
)

func main() {
	cfg := mustLoadConfig()
	log := mustMakeLogger(cfg.LogLevel)
	greetings(log)

	storage := mustMakeStorage(log, &cfg.DB)
	mustMakeMigrations(log, storage, cfg.DB.DBConnString)

	metricService := core.NewMetricService(log, storage)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	grpcServerGracefulStop := mustStartGRPCServer(log, ctx, cfg.GRPCAddress, metricService)
	restServerGracefulStop := mustStartRESTServer(log, ctx, cfg, metricService)

	<-ctx.Done()

	grpcServerGracefulStop()
	restServerGracefulStop()
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

func mustStartGRPCServer(log *slog.Logger, ctx context.Context, grpcAddress string, metricService *core.MetricService) func() {
	lis, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		log.Error("failed to listen gRPC", slog.String("error", err.Error()))
		os.Exit(1)
	}

	s := grpc.NewServer()
	metricspb.RegisterMetricsCollectorServer(s, metricsgrpc.NewServer(log, metricService))
	reflection.Register(s)

	go func() {
		log.Info("gRPC server started", slog.String("address", grpcAddress))
		if err := s.Serve(lis); err != nil {
			log.Error("gRPC server failed", slog.String("error", err.Error()))
		}
	}()

	return func() {
		log.Debug("stopping gRPC server gracefully")
		s.GracefulStop()
		log.Info("gRPC server stopped")
	}
}

func mustMakeMux(log *slog.Logger, metricService *core.MetricService) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", rest.NewPingHandler(log))
	mux.HandleFunc("GET /metric", rest.NewGetMetricByMetricIdentityHandler(log, metricService))
	mux.HandleFunc("POST /metric", rest.NewCreateMetricHandler(log, metricService))

	log.Info("mux initialized with routes")

	return mux
}

func mustStartRESTServer(log *slog.Logger, ctx context.Context, cfg *config.Config, metricService *core.MetricService) func() {
	mux := mustMakeMux(log, metricService)
	server := &http.Server{
		Addr:        cfg.AppAddress,
		ReadTimeout: cfg.ReadTimeout,
		Handler:     mux,
	}

	go func() {
		log.Info("REST server started", slog.String("address", cfg.AppAddress))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("REST server closed unexpectedly", slog.String("error", err.Error()))
		}
	}()

	return func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

		log.Debug("stopping REST server gracefully")
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error("REST server shutdown error", slog.String("error", err.Error()))
		}
		log.Info("REST server stopped")
	}
}
