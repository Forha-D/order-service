package worker

import (
	"context"
	"order-service/internal/publisher"
	"order-service/internal/repository"
	"time"
)

type OutboxWorker struct {
	repo      *repository.OutboxRepository
	publisher *publisher.KafkaPublisher
}

func (w *OutboxWorker) Start(ctx context.Context) {

	for {
		events, _ := w.repo.GetPending(ctx)

		for _, event := range events {

			err := w.publisher.Publish(event.EventType, event.Payload)

			if err != nil {
				continue
			}

			w.repo.MarkAsSent(ctx, event.ID)
		}

		time.Sleep(2 * time.Second)
	}
}
