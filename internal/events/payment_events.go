package events

type PaymentSuccessPayload struct {
	OrderID       string `json:"order_id"`
	PaymentID     string `json:"payment_id"`
	TransactionID string `json:"transaction_id"`
}

type PaymentFailedPayload struct {
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}
