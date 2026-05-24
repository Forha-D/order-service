package consumer

import (
	"context"
	"encoding/json"
	"log"
	"order-service/internal/events"
	"order-service/internal/service"

	"github.com/segmentio/kafka-go"
)

type OrderConsumer struct {
	reader  *kafka.Reader
	service *service.OrderService
}

func NewOrderConsumer(
	reader *kafka.Reader,
	service *service.OrderService,
) *OrderConsumer {

	return &OrderConsumer{
		reader:  reader,
		service: service,
	}
}

func (c *OrderConsumer) Start(ctx context.Context) {

	log.Println("[consumer] order consumer started")

	for {
		msg, err := c.reader.ReadMessage(ctx)

		if err != nil {
			log.Printf("[consumer] read error: %v", err)

			if ctx.Err() != nil {
				log.Println("[consumer] shutting down")
				return
			}
			continue
		}

		err = c.handleMessage(ctx, msg)
		if err != nil {
			log.Printf("[consumer] handle error: %v", err)
		}
	}
}

func (c *OrderConsumer) handleMessage(
	ctx context.Context,
	msg kafka.Message,
) error {

	var envelope events.EventEnvelope

	err := json.Unmarshal(msg.Value, &envelope)
	if err != nil {
		return err
	}

	switch envelope.EventType {

	case "payment.success":
		return c.handlePaymentSuccess(ctx, envelope)

	case "payment.failed":
		return c.handlePaymentFailed(ctx, envelope)

	default:
		log.Printf("[consumer] unknown event: %s", envelope.EventType)
	}

	return nil
}

func (c *OrderConsumer) handlePaymentSuccess(
	ctx context.Context,
	envelope events.EventEnvelope,
) error {

	data, err := json.Marshal(envelope.Payload)
	if err != nil {
		return err
	}

	var payload events.PaymentSuccessPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	log.Printf("[consumer] payment success orderID=%s", payload.OrderID)

	return c.service.UpdateOrderStatus(
		ctx,
		payload.OrderID,
		"PAID",
	)
}

func (c *OrderConsumer) handlePaymentFailed(
	ctx context.Context,
	envelope events.EventEnvelope,
) error {

	data, err := json.Marshal(envelope.Payload)
	if err != nil {
		return err
	}

	var payload events.PaymentFailedPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	log.Printf("[consumer] payment failed orderID=%s", payload.OrderID)

	return c.service.UpdateOrderStatus(
		ctx,
		payload.OrderID,
		"CANCELLED",
	)
}
