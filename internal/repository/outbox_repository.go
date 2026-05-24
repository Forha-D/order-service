package repository

import (
	"context"
	"order-service/internal/model"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type OutboxRepository struct {
	collection *mongo.Collection
}

func (r *OutboxRepository) Create(ctx context.Context, event *model.OutboxEvent) error {
	_, err := r.collection.InsertOne(ctx, event)
	return err
}

func (r *OutboxRepository) GetPending(ctx context.Context) ([]model.OutboxEvent, error) {

	cursor, err := r.collection.Find(ctx, bson.M{
		"status": "PENDING",
	})

	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []model.OutboxEvent

	for cursor.Next(ctx) {
		var event model.OutboxEvent
		if err := cursor.Decode(&event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

func (r *OutboxRepository) MarkAsSent(ctx context.Context, id string) error {

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"status":       "SENT",
				"processed_at": time.Now(),
			},
		},
	)

	return err
}
