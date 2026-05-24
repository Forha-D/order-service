package dto

type CreateOrderRequest struct {
    Items []CreateOrderItem      `json:"items"` 
}


type CreateOrderItem struct {
ProductID   string       `json:"product_id"`
Name        string       `json:"name"`
Quantity    int          `json:"quantity"`
Price       float64      `json:"price"`
}
