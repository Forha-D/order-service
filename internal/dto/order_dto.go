package dto

type CreateOrderRequest struct {
	Items []CreateOrderItem `json:"items"`
	// IdempotencyKey is not part of the JSON body but will be set from the header in the handler
	IdempotencyKey string `json:"-"`
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
