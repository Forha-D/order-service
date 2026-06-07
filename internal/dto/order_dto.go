package dto

type CreateOrderRequest struct {
	Items          []CreateOrderItem `json:"items"`
	IdempotencyKey string            `json:"-"` // set from header, not request body
}

// Update status
type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}

type CreateOrderItem struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}
