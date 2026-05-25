package worker

import (
	"context"
	"order-service/internal/model"
	"order-service/internal/publisher"
	"order-service/internal/repository"
	"time"
)

const MaxRetryCount = 5

type OutboxWorker struct {
	repo      *repository.OutboxRepository
	publisher *publisher.KafkaPublisher
}

func (w *OutboxWorker) Start(ctx context.Context) {

	for {
		events, err := w.repo.GetPending(ctx)

		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		for _, event := range events {

			//err := w.publisher.Publish(event.EventType, event.Payload)
			w.processEvent(
				ctx,
				event,
			)
		}

		//w.repo.MarkAsSent(ctx, event.ID)
		time.Sleep(2 * time.Second)
	}

}

// retry logic

func (w *OutboxWorker) processEvent(ctx context.Context, event model.OutboxEvent) {

	if event.RetryCount >= MaxRetryCount {

		_ = w.repo.MarkAsDead(

			ctx,
			event.ID,
		)
		return
	}

	err := w.publisher.Publish(event.EventType, event.Payload)

	if err != nil {
		_ = w.repo.MarkAsFailed(
			ctx,
			event.ID,
			err.Error(),
		)
		return
	}

	_ = w.repo.MarkAsSent(ctx, event.ID)
}
