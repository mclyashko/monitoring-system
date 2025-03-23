package core

type OrderRepository interface {
	Save(order Order) (int, error)
	FindByID(orderId int) (*Order, error)
}
