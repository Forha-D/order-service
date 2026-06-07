package repository

import (
	"context"
	"errors"
	"time"

	"order-service/internal/model"
	"order-service/internal/mongoerr"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const (
	defaultDBTimeout = 5 * time.Second
	maxRetryAttempts = 3
)

type OrderRepository struct {
	collection *mongo.Collection
}

func NewOrderRepository(db *mongo.Database) *OrderRepository {
	return &OrderRepository{
		collection: db.Collection("orders"),
	}
}

func withDatabaseContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		return context.WithTimeout(ctx, defaultDBTimeout)
	}
	return ctx, func() {}
}

func retryOnTransientFailure(ctx context.Context, fn func(context.Context) error) error {
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

func (r *OrderRepository) Create(ctx context.Context, order *model.Order) error {
	ctx, cancel := withDatabaseContext(ctx)
	defer cancel()

	return retryOnTransientFailure(ctx, func(ctx context.Context) error {
		if order.ID.IsZero() {
			order.ID = primitive.NewObjectID()
		}
		order.CreatedAt = time.Now()
		order.UpdatedAt = time.Now()
		_, err := r.collection.InsertOne(ctx, order)
		return err
	})
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*model.Order, error) {
	ctx, cancel := withDatabaseContext(ctx)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var order model.Order
	err = retryOnTransientFailure(ctx, func(ctx context.Context) error {
		return r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&order)
	})
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) GetUserByID(ctx context.Context, userID string) ([]model.Order, error) {
	ctx, cancel := withDatabaseContext(ctx)
	defer cancel()

	var cursor *mongo.Cursor
	err := retryOnTransientFailure(ctx, func(ctx context.Context) error {
		var findErr error
		cursor, findErr = r.collection.Find(ctx, bson.M{"user_id": userID})
		return findErr
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	orders := make([]model.Order, 0)
	for cursor.Next(ctx) {
		var order model.Order
		if err := cursor.Decode(&order); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID string, status string) error {
	ctx, cancel := withDatabaseContext(ctx)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return err
	}

	return retryOnTransientFailure(ctx, func(ctx context.Context) error {
		_, err := r.collection.UpdateOne(
			ctx,
			bson.M{"_id": objectID},
			bson.M{"$set": bson.M{
				"status":     status,
				"updated_at": time.Now(),
			}},
		)
		return err
	})
}
