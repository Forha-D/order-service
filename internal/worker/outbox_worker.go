package worker

import (
	"context"
	"log"
	"time"

	"order-service/internal/config"
	"order-service/internal/model"
	"order-service/internal/publisher"
	"order-service/internal/repository"
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
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	log.Println("[OutboxWorker] started")

	for {
		select {
		case <-ctx.Done():
			log.Println("[OutboxWorker] shutting down")
			return
		case <-ticker.C:
			w.poll(ctx)
		}
	}
}

func (w *OutboxWorker) poll(ctx context.Context) {
	events, err := w.repo.GetPending(ctx)
	if err != nil {
		log.Printf("[OutboxWorker] failed to get pending events: %v", err)
		return
	}
	for _, event := range events {
		w.processEvent(ctx, event)
	}
}

func (w *OutboxWorker) processEvent(ctx context.Context, event model.OutboxEvent) {
	if event.RetryCount >= MaxRetryCount {
		dlqErr := w.publisher.PublishToTopic(
			w.cfg.KafkaDLQTopic,
			event.ID.Hex(),
			[]byte(event.Payload),
		)
		if dlqErr != nil {
			log.Printf("[OutboxWorker] failed to publish event %s to DLQ: %v", event.ID.Hex(), dlqErr)
			return
		}
		_ = w.repo.MarkAsDead(ctx, event.ID)
		return
	}

	err := w.publisher.PublishToTopic(
		event.EventType,
		event.ID.Hex(),
		[]byte(event.Payload),
	)
	if err != nil {
		log.Printf("[OutboxWorker] failed to publish event %s: %v", event.ID.Hex(), err)
		_ = w.repo.MarkAsFailed(ctx, event.ID, err.Error())
		return
	}

	_ = w.repo.MarkAsSent(ctx, event.ID)
}
