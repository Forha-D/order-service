package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"order-service/internal/domain"
	"order-service/internal/dto"
	"order-service/internal/model"
	"order-service/internal/repository"
	"order-service/internal/utils"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type OrderService struct {
	orderRepo       *repository.OrderRepository
	outboxRepo      *repository.OutboxRepository
	idempotencyRepo *repository.IdempotencyRepository
}

func NewOrderService(orderRepo *repository.OrderRepository, outboxRepo *repository.OutboxRepository, idempotencyRepo *repository.IdempotencyRepository) *OrderService {
	return &OrderService{
		orderRepo:       orderRepo,
		outboxRepo:      outboxRepo,
		idempotencyRepo: idempotencyRepo,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID string, req dto.CreateOrderRequest) (*model.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()

	req.IdempotencyKey = strings.TrimSpace(req.IdempotencyKey)
	if req.IdempotencyKey == "" {
		return nil, errors.New("idempotency key required")
	}

	if len(req.Items) == 0 {
		return nil, errors.New("items cannot be empty")
	}

	var totalAmount float64
	orderItems := make([]model.OrderItem, 0, len(req.Items))

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

	// STEP 1: request hash (include user scope, not the raw idempotency key itself)
	requestPayload := struct {
		UserID string                `json:"user_id"`
		Items  []dto.CreateOrderItem `json:"items"`
	}{
		UserID: userID,
		Items:  req.Items,
	}

	requestBytes, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, err
	}

	requestHash := utils.SHA256(string(requestBytes))
	scopedKey := utils.BuildIdempotencyKey(userID, "create_order", req.IdempotencyKey)

	// STEP 2: claim idempotency FIRST
	record, err := s.idempotencyRepo.Claim(
		ctx,
		scopedKey,
		requestHash,
		24*time.Hour,
	)
	if err != nil {
		return nil, err
	}

	// STEP 3: payload mismatch protection
	if record.RequestHash != requestHash {
		return nil, errors.New("idempotency key reused with different payload")
	}

	// STEP 4: replay case
	if record.Status == model.IdempotencyCompleted {
		var cached model.Order
		if err := json.Unmarshal([]byte(record.Response), &cached); err != nil {
			return nil, err
		}
		return &cached, nil
	}

	if record.Status == model.IdempotencyInProgress {
		return nil, errors.New("request already in progress")
	}

	// STEP 5: build order (ONLY NOW)
	order := &model.Order{
		UserID: userID,
		// IdempotencyKey: req.IdempotencyKey,
		Items:       orderItems,
		TotalAmount: totalAmount,
		Status:      string(model.OrderPending),
	}

	// STEP 6: create order
	if err := s.orderRepo.Create(ctx, order); err != nil {
		_ = s.idempotencyRepo.MarkFailed(ctx, scopedKey, err.Error())
		return nil, err
	}

	// STEP 7: outbox event
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
		_ = s.idempotencyRepo.MarkFailed(ctx, scopedKey, "outbox creation failed")
	}

	// STEP 8: mark idempotency COMPLETE ONCE
	if err := s.idempotencyRepo.MarkCompleted(
		ctx,
		scopedKey,
		string(orderBytes),
	); err != nil {
		log.Printf("failed to mark completed: %v", err)
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
