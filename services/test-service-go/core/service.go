package core

import "log/slog"

type OrderService struct {
	log  *slog.Logger
	repo OrderRepository
}

func NewOrderService(log *slog.Logger, repo OrderRepository) *OrderService {
	return &OrderService{
		log:  log,
		repo: repo,
	}
}

func (s *OrderService) CreateOrder(order Order) (int, error) {
	if order.ProductID <= 0 || order.Quantity <= 0 || order.UserID <= 0 {
		s.log.Warn("validation failed for order", slog.Any("order", order))
		return 0, ErrInvalidOrder
	}

	id, err := s.repo.Save(order)
	if err != nil {
		s.log.Error("failed to save order", slog.String("error", err.Error()))
		return 0, ErrSaveFailed
	}

	s.log.Info("order successfully created", slog.Int("order_id", id))
	return id, nil
}

func (s *OrderService) GetOrderByID(orderID int) (*Order, error) {
	if orderID <= 0 {
		s.log.Warn("invalid order ID", slog.Int("order_id", orderID))
		return nil, ErrInvalidOrderID
	}

	order, err := s.repo.FindByID(orderID)
	if err != nil {
		s.log.Error("failed to find order", slog.String("error", err.Error()))
		return nil, ErrOrderNotFound
	}

	s.log.Info("order successfully retrieved", slog.Int("order_id", orderID))
	return order, nil
}
