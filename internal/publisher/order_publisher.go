package publisher

import (
	"encoding/json"
	"log"
	"time"

	"order-service/internal/events"
	"order-service/internal/model"

	"github.com/google/uuid"
)

func (p *KafkaPublisher) PublishOrderCreated(
	order *model.Order,
) error {

	var items []events.OrderItem

	for _, item := range order.Items {

		items = append(items, events.OrderItem{
			ProductID: item.ProductID,
			Name:      item.Name,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}

	payload := events.OrderCreatedPayload{
		OrderID:     order.ID.Hex(),
		UserID:      order.UserID,
		Status:      string(order.Status),
		TotalAmount: order.TotalAmount,
		Items:       items,
	}

	event := events.EventEnvelope{
		EventID:        uuid.NewString(),
		EventType:      events.OrderCreated,
		Version:        "v1",
		Source:         "order-service",
		Timestamp:      time.Now(),
		CorrelationID:  order.ID.Hex(),
		IdempotencyKey: order.ID.Hex(),
		Payload:        payload,
	}

	//convert to JSON (IMPORTANT)
	body, err := json.Marshal(event)
	if err != nil {
		log.Printf("[publisher] marshal failed: %v", err)
		return err
	}

	return p.PublishToTopic(
		"order.created",
		order.ID.Hex(),
		body,
	)
}
