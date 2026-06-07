package repository

import (
	"context"
	"time"

	"order-service/internal/model"
	"order-service/internal/mongoerr"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type OutboxRepository struct {
	collection *mongo.Collection
}

func NewOutboxRepository(db *mongo.Database) *OutboxRepository {
	return &OutboxRepository{
		collection: db.Collection("outbox"),
	}
}

func withOutboxContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		return context.WithTimeout(ctx, defaultDBTimeout)
	}
	return ctx, func() {}
}

func retryOutboxOperation(ctx context.Context, fn func(context.Context) error) error {
	var err error
	for attempt := 1; attempt <= maxRetryAttempts; attempt++ {
		err = fn(ctx)
		if err == nil || !mongoerr.IsTransient(err) {
			return err
		}
		if ctx.Err() != nil {
			return err
		}
		time.Sleep(time.Duration(attempt) * 200 * time.Millisecond)
	}
	return err
}

func (r *OutboxRepository) Create(ctx context.Context, event *model.OutboxEvent) error {
	ctx, cancel := withOutboxContext(ctx)
	defer cancel()

	return retryOutboxOperation(ctx, func(ctx context.Context) error {
		if event.ID.IsZero() {
			event.ID = primitive.NewObjectID()
		}
		if event.CreatedAt.IsZero() {
			event.CreatedAt = time.Now()
		}
		if event.Status == "" {
			event.Status = model.OutboxStatusPending
		}
		_, err := r.collection.InsertOne(ctx, event)
		return err
	})
}

func (r *OutboxRepository) GetPending(ctx context.Context) ([]model.OutboxEvent, error) {
	ctx, cancel := withOutboxContext(ctx)
	defer cancel()

	var cursor *mongo.Cursor
	err := retryOutboxOperation(ctx, func(ctx context.Context) error {
		var findErr error
		cursor, findErr = r.collection.Find(
			ctx,
			bson.M{"status": model.OutboxStatusPending},
			options.Find().SetLimit(100).SetSort(bson.D{{Key: "created_at", Value: 1}}),
		)
		return findErr
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	events := make([]model.OutboxEvent, 0)
	for cursor.Next(ctx) {
		var event model.OutboxEvent
		if err := cursor.Decode(&event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

func (r *OutboxRepository) MarkAsSent(ctx context.Context, id primitive.ObjectID) error {
	ctx, cancel := withOutboxContext(ctx)
	defer cancel()

	now := time.Now()
	return retryOutboxOperation(ctx, func(ctx context.Context) error {
		_, err := r.collection.UpdateOne(
			ctx,
			bson.M{"_id": id},
			bson.M{"$set": bson.M{
				"status":       model.OutboxStatusSent,
				"processed_at": &now,
			}},
		)
		return err
	})
}

func (r *OutboxRepository) MarkAsFailed(ctx context.Context, id primitive.ObjectID, errMsg string) error {
	ctx, cancel := withOutboxContext(ctx)
	defer cancel()

	return retryOutboxOperation(ctx, func(ctx context.Context) error {
		_, err := r.collection.UpdateOne(
			ctx,
			bson.M{"_id": id},
			bson.M{
				"$inc": bson.M{"retry_count": 1},
				"$set": bson.M{"last_error": errMsg},
			},
		)
		return err
	})
}

func (r *OutboxRepository) MarkAsDead(ctx context.Context, id primitive.ObjectID) error {
	ctx, cancel := withOutboxContext(ctx)
	defer cancel()

	return retryOutboxOperation(ctx, func(ctx context.Context) error {
		_, err := r.collection.UpdateOne(
			ctx,
			bson.M{"_id": id},
			bson.M{"$set": bson.M{"status": model.OutboxStatusDead}},
		)
		return err
	})
}
