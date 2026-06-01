package worker

import (
	"context"
	"order-service/internal/config"
	"order-service/internal/model"
	"order-service/internal/publisher"
	"order-service/internal/repository"
	"time"
)

const MaxRetryCount = 5

type OutboxWorker struct {
	repo      *repository.OutboxRepository
	publisher *publisher.KafkaPublisher
	cfg       *config.Config
}

func NewOutboxWorker(
	repo *repository.OutboxRepository,
	publisher *publisher.KafkaPublisher,
	cfg *config.Config,
) *OutboxWorker {

	return &OutboxWorker{
		repo:      repo,
		publisher: publisher,
		cfg:       cfg,
	}
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

func (w *OutboxWorker) processEvent(
	ctx context.Context,
	event model.OutboxEvent,
) {

	// 1. retry limit check
	if event.RetryCount >= MaxRetryCount {

		// 2. publish to DLQ BEFORE marking dead
		dlqErr := w.publisher.PublishToTopic(
			w.cfg.KafkaDLQTopic, // "order.dlq"
			event.ID,
			[]byte(event.Payload),
		)

		if dlqErr != nil {
			// if DLQ fails → keep retrying later
			return
		}

		// 3. mark as dead after successful DLQ publish
		_ = w.repo.MarkAsDead(ctx, event.ID)
		return
	}

	// 4. normal publish
	err := w.publisher.PublishToTopic(
		event.EventType, // topic like order.created
		event.ID,
		[]byte(event.Payload),
	)

	if err != nil {
		_ = w.repo.MarkAsFailed(ctx, event.ID, err.Error())
		return
	}

	// 5. success
	_ = w.repo.MarkAsSent(ctx, event.ID)
}
