package rest

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/mclyashko/monitoring-system/services/test-service-go/core"
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

type OrderDTO struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
	UserID    int `json:"user_id"`
}

func NewCreateOrderHandler(log *slog.Logger, service *core.OrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var orderDTO OrderDTO
		if err := json.NewDecoder(r.Body).Decode(&orderDTO); err != nil {
			log.Error("failed to parse request", slog.String("error", err.Error()))
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		order := core.Order{
			ProductID: orderDTO.ProductID,
			Quantity:  orderDTO.Quantity,
			UserID:    orderDTO.UserID,
		}

		id, err := service.CreateOrder(order)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrInvalidOrder):
				log.Warn("order validation failed", slog.String("error", err.Error()))
				http.Error(w, err.Error(), http.StatusBadRequest)
			case errors.Is(err, core.ErrSaveFailed):
				log.Error("failed to save order", slog.String("error", err.Error()))
				http.Error(w, "internal error", http.StatusInternalServerError)
			default:
				log.Error("unexpected error", slog.String("error", err.Error()))
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
			return
		}

		response := struct {
			ID int `json:"id"`
		}{ID: id}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response)
	}
}

func NewGetOrderByIDHandler(log *slog.Logger, service *core.OrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderIDStr := r.PathValue("id")
		if orderIDStr == "" {
			log.Warn("order ID is missing in request", slog.String("error", "missing order_id"))
			http.Error(w, "order_id is required", http.StatusBadRequest)
			return
		}

		orderID, err := strconv.Atoi(orderIDStr)
		if err != nil {
			log.Warn("invalid order ID format", slog.String("order_id", orderIDStr), slog.String("error", err.Error()))
			http.Error(w, "invalid order ID", http.StatusBadRequest)
			return
		}

		order, err := service.GetOrderByID(orderID)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrInvalidOrderID):
				log.Warn("invalid order ID", slog.Int("order_id", orderID), slog.String("error", err.Error()))
				http.Error(w, err.Error(), http.StatusBadRequest)
			case errors.Is(err, core.ErrOrderNotFound):
				log.Warn("order not found", slog.Int("order_id", orderID), slog.String("error", err.Error()))
				http.Error(w, err.Error(), http.StatusNotFound)
			default:
				log.Error("unexpected error", slog.String("error", err.Error()))
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
			return
		}

		response := struct {
			ID        int `json:"id"`
			ProductID int `json:"product_id"`
			Quantity  int `json:"quantity"`
			UserID    int `json:"user_id"`
		}{
			ID:        order.ID,
			ProductID: order.ProductID,
			Quantity:  order.Quantity,
			UserID:    order.UserID,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}
