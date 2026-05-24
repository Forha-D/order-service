package repository

import (
	"context"
	"log"
	"time"

	"order-service/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type OrderRepository struct {
	collection *mongo.Collection
}

func NewOrderRepository(db *mongo.Database) *OrderRepository {

	return &OrderRepository{

		collection: db.Collection("orders"),
	}

}

func (r *OrderRepository) Create(ctx context.Context, order *model.Order) error {

	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, order)

	return err
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*model.Order, error) {

	ObjectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		log.Printf("Invalid ObjectID: %v", err)
	}

	var order model.Order

	err = r.collection.FindOne(ctx, bson.M{"_id": ObjectID}).Decode(&order)

	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (r *OrderRepository) GetUserByID(ctx context.Context, userID string) ([]model.Order, error) {

	cursor, err := r.collection.Find(ctx, bson.M{"user_ID": userID})

	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var orders []model.Order

	for cursor.Next(ctx) {
		var order model.Order

		err := cursor.Decode(&order)

		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (r *OrderRepository) UpdateStatus(
	ctx context.Context,
	orderID string,
	status string,
) error {

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": orderID},
		bson.M{
			"$set": bson.M{
				"status":     status,
				"updated_at": time.Now(),
			},
		},
	)

	return err
}
