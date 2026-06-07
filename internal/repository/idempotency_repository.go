package repository

import (
	"context"
	"order-service/internal/model"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type IdempotencyRepository struct {
	collection *mongo.Collection
}

func NewIdempotencyRepository(db *mongo.Database) *IdempotencyRepository {
	return &IdempotencyRepository{
		collection: db.Collection("idempotency_records"),
	}
}

func (r *IdempotencyRepository) GetByKey(
	ctx context.Context,
	key string,
) (*model.IdempotencyRecord, error) {

	var record model.IdempotencyRecord

	err := r.collection.
		FindOne(ctx, bson.M{"key": key}).
		Decode(&record)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &record, nil
}

func (r *IdempotencyRepository) Claim(
	ctx context.Context,
	key string,
	requestHash string,
	ttl time.Duration,
) (*model.IdempotencyRecord, error) {

	now := time.Now()

	record := &model.IdempotencyRecord{
		Key:         key,
		RequestHash: requestHash,
		Status:      model.IdempotencyInProgress,
		CreatedAt:   now,
		UpdatedAt:   now,
		ExpiresAt:   now.Add(ttl),
	}

	filter := bson.M{
		"key": key,
	}

	update := bson.M{
		"$setOnInsert": record,
	}

	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var result model.IdempotencyRecord

	err := r.collection.
		FindOneAndUpdate(ctx, filter, update, opts).
		Decode(&result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *IdempotencyRepository) MarkCompleted(
	ctx context.Context,
	key string,
	response string,
) error {

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"key": key},
		bson.M{
			"$set": bson.M{
				"status":     model.IdempotencyCompleted,
				"response":   response,
				"updated_at": time.Now(),
			},
		},
	)

	return err
}

func (r *IdempotencyRepository) MarkFailed(
	ctx context.Context,
	key string,
	message string,
) error {

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"key": key},
		bson.M{
			"$set": bson.M{
				"status":     model.IdempotencyFailed,
				"error":      message,
				"updated_at": time.Now(),
			},
		},
	)

	return err
}

func (r *IdempotencyRepository) EnsureIndexes(
	ctx context.Context,
) error {

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "key", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetName("idx_idempotency_key"),
		},
		{
			Keys: bson.D{
				{Key: "expires_at", Value: 1},
			},
			Options: options.Index().
				SetExpireAfterSeconds(0).
				SetName("idx_idempotency_ttl"),
		},
	}

	_, err := r.collection.Indexes().
		CreateMany(ctx, indexes)

	return err
}
