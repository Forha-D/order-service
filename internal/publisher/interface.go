package publisher

import (
	"order-service/internal/model"
)

type EventPublisher interface {
	PublishOrderCreated(order *model.Order) error
}
