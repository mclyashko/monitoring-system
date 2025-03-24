package rest

import (
	"log/slog"
	"net/http"
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
