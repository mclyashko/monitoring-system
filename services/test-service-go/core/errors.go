package core

import "errors"

var (
	ErrInvalidOrder = errors.New("invalid order: product ID, quantity, and user ID must be positive")
	ErrSaveFailed   = errors.New("failed to save order")
)

var (
	ErrOrderNotFound  = errors.New("order not found")
	ErrInvalidOrderID = errors.New("invalid order ID")
)
