package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"order-service/internal/domain"
	"order-service/internal/dto"
	"order-service/internal/model"
	"order-service/internal/repository"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type OrderService struct {
	orderRepo  *repository.OrderRepository
	outboxRepo *repository.OutboxRepository
}

func NewOrderService(orderRepo *repository.OrderRepository, outboxRepo *repository.OutboxRepository) *OrderService {
	return &OrderService{
		orderRepo:  orderRepo,
		outboxRepo: outboxRepo,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID string, req dto.CreateOrderRequest) (*model.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()

	if len(req.Items) == 0 {
		return nil, errors.New("items cannot be empty")
	}

	var totalAmount float64
	var orderItems []model.OrderItem

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return nil, errors.New("invalid quantity")
		}
		if item.Price <= 0 {
			return nil, errors.New("invalid price")
		}
		totalAmount += float64(item.Quantity) * item.Price
		orderItems = append(orderItems, model.OrderItem{
			ProductID: item.ProductID,
			Name:      item.Name,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}

	// idempotency key — use provided or generate one
	idempotencyKey := req.IdempotencyKey
	if idempotencyKey == "" {
		idempotencyKey = uuid.NewString()
	}

	// check for duplicate request
	existingOrder, err := s.orderRepo.GetByIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		return nil, err // real DB error
	}
	if existingOrder != nil {
		return existingOrder, nil // duplicate — return cached result
	}

	order := &model.Order{
		UserID:         userID,
		IdempotencyKey: idempotencyKey,
		Items:          orderItems,
		TotalAmount:    totalAmount,
		Status:         string(model.OrderPending),
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	orderBytes, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}

	outboxEvent := &model.OutboxEvent{
		EventType: "order.created",
		Payload:   string(orderBytes),
		Status:    model.OutboxStatusPending,
		CreatedAt: time.Now(),
	}

	if err := s.outboxRepo.Create(ctx, outboxEvent); err != nil {
		// don't fail the order — but this must be visible
		log.Printf("[OrderService] failed to create outbox event for order %s: %v", order.ID.Hex(), err)
	}

	return order, nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, id string) (*model.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, mongo.ErrNoDocuments
	}
	return order, nil
}

func (s *OrderService) GetOrderByUserID(ctx context.Context, userID string) ([]model.Order, error) {
	return s.orderRepo.GetUserByID(ctx, userID)
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID string, newStatus string) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return mongo.ErrNoDocuments
	}

	if err := domain.CanTransition(string(order.Status), newStatus); err != nil {
		return err
	}

	return s.orderRepo.UpdateStatus(ctx, orderID, newStatus)
}
