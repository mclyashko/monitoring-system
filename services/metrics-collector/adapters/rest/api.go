package rest

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/mclyashko/monitoring-system/services/metrics-collector/core"
)

func NewPingHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("received ping request", slog.String("method", r.Method), slog.String("url", r.URL.String()))

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Pong"))

		if err != nil {
			log.Error("failed to send Pong response", slog.String("error", err.Error()))
		} else {
			log.Info("sent response: Pong")
		}
	}
}

type MetricDTO struct {
	ServiceURL  string  `json:"service_url"`
	MetricName  string  `json:"metric_name"`
	PodName     string  `json:"pod_name"`
	MetricValue float64 `json:"metric_value"`
}

func NewCreateMetricHandler(log *slog.Logger, service *core.MetricService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metricDTO MetricDTO
		if err := json.NewDecoder(r.Body).Decode(&metricDTO); err != nil {
			log.Error("failed to parse request", slog.String("error", err.Error()))
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		metric := core.Metric{
			MetricIdentity: core.MetricIdentity{
				Time:       time.Now().UTC(),
				ServiceURL: metricDTO.ServiceURL,
				MetricName: metricDTO.MetricName,
				PodName:    metricDTO.PodName,
			},
			MetricValue: metricDTO.MetricValue,
		}

		identity, err := service.CreateMetric(metric)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrInvalidMetric):
				log.Warn("metric validation failed", slog.String("error", err.Error()))
				http.Error(w, err.Error(), http.StatusBadRequest)
			case errors.Is(err, core.ErrSaveFailed):
				log.Error("failed to save metric", slog.String("error", err.Error()))
				http.Error(w, "internal error", http.StatusInternalServerError)
			default:
				log.Error("unexpected error", slog.String("error", err.Error()))
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
			return
		}

		response := struct {
			Time       time.Time `json:"time"`
			ServiceURL string    `json:"service_url"`
			MetricName string    `json:"metric_name"`
			PodName    string    `json:"pod_name"`
		}{
			Time:       identity.Time,
			ServiceURL: identity.ServiceURL,
			MetricName: identity.MetricName,
			PodName:    identity.PodName,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response)
	}
}

type MetricIdentity struct {
	Time       time.Time `json:"time"`
	ServiceURL string    `json:"service_url"`
	MetricName string    `json:"metric_name"`
	PodName    string    `json:"pod_name"`
}

func NewGetMetricByMetricIdentityHandler(log *slog.Logger, service *core.MetricService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		timeStr := query.Get("time")
		serviceUrl := query.Get("service_url")
		metricName := query.Get("metric_name")
		podName := query.Get("pod_name")

		parsedTime, err := time.Parse(time.RFC3339Nano, timeStr)
		if err != nil {
			log.Warn("invalid time format", slog.String("time", timeStr), slog.String("error", err.Error()))
			http.Error(w, "invalid time format", http.StatusBadRequest)
			return
		}

		metricIdentity := core.MetricIdentity{
			Time:       parsedTime,
			ServiceURL: serviceUrl,
			MetricName: metricName,
			PodName:    podName,
		}

		metric, err := service.GetMetricByMetricIdentity(metricIdentity)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrInvalidMetricIdentity):
				log.Warn("invalid metric identity", slog.Any("metric_identity", metricIdentity), slog.String("error", err.Error()))
				http.Error(w, err.Error(), http.StatusBadRequest)
			case errors.Is(err, core.ErrMetricNotFound):
				log.Warn("metric not found", slog.Any("metric_identity", metricIdentity), slog.String("error", err.Error()))
				http.Error(w, err.Error(), http.StatusNotFound)
			default:
				log.Error("unexpected error", slog.String("error", err.Error()))
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
			return
		}

		response := struct {
			Time        time.Time `json:"time"`
			ServiceURL  string    `json:"service_url"`
			MetricName  string    `json:"metric_name"`
			PodName     string    `json:"pod_name"`
			MetricValue float64   `json:"metric_value"`
		}{
			Time:        metric.Time,
			ServiceURL:  metric.ServiceURL,
			MetricName:  metric.MetricName,
			PodName:     metric.PodName,
			MetricValue: metric.MetricValue,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}
