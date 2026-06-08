package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderStatus string

const (
	OrderPending    OrderStatus = "PENDING"
	OrderPaid       OrderStatus = "PAID"
	OrderProcessing OrderStatus = "PROCESSING"
	OrderShipped    OrderStatus = "SHIPPED"
	OrderDelivered  OrderStatus = "DELIVERED"
	OrderCancelled  OrderStatus = "CANCELLED"
)

type OrderItem struct {
	ProductID string  `json:"product_id"   bson:"product_id"`
	Name      string  `json:"name"         bson:"name"`
	Quantity  int     `json:"quantity"     bson:"quantity"`
	Price     float64 `json:"price"        bson:"price"`
}

type Order struct {
	ID     primitive.ObjectID `json:"id"                 bson:"_id,omitempty"`
	UserID string             `json:"user_id"            bson:"user_id"`
	// IdempotencyKey string             `json:"idempotency_key"    bson:"idempotency_key,omitempty"`
	Items       []OrderItem `json:"items"              bson:"items"`
	TotalAmount float64     `json:"total_amount"       bson:"total_amount"`
	Status      string      `json:"status"             bson:"status"`
	CreatedAt   time.Time   `json:"created_at"         bson:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"         bson:"updated_at"`
}
