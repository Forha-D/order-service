package service

import (
	"context"
	"encoding/json"
	"errors"

	// "log"
	"order-service/internal/domain"
	"order-service/internal/dto"
	"order-service/internal/model"
	"order-service/internal/repository"
	"time"
)

type OrderService struct {
	orderRepo *repository.OrderRepository
	// publisher publisher.EventPublisher
	outboxRepo *repository.OutboxRepository
}

func NewOrderService(orderRepo *repository.OrderRepository, outboxRepo *repository.OutboxRepository) *OrderService {
	return &OrderService{
		orderRepo:  orderRepo,
		outboxRepo: outboxRepo,
		// publisher: publisher,

	}

}

func (s *OrderService) CreateOrder(ctx context.Context, userID string, req dto.CreateOrderRequest) (*model.Order, error) {

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

	order := &model.Order{
		UserID:      userID,
		Items:       orderItems,
		TotalAmount: totalAmount,
		Status:      string(model.OrderPending),
	}

	err := s.orderRepo.Create(ctx, order)
	if err != nil {
		return nil, err
	}

	orderBytes, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}

	outboxEvent := model.OutboxEvent{
		EventType: "order.created",
		Payload:   string(orderBytes),
		Status:    "PENDING",
		CreatedAt: time.Now(),
	}

	_ = s.outboxRepo.Create(ctx, &outboxEvent)

	return order, nil

	// err = s.publisher.PublishOrderCreated(order)
	// if err != nil {
	// 	// IMPORTANT: DO NOT FAIL ORDER CREATION
	// 	// Kafka is eventual delivery system
	// 	log.Printf("[OrderService] failed to publish order.created event: %v", err)
	// }

	// return order, nil

}

func (s *OrderService) GetOrderByID(ctx context.Context, id string) (*model.Order, error) {
	return s.orderRepo.GetByID(ctx, id)
}

func (s *OrderService) GetOrderByUserID(ctx context.Context, userID string) ([]model.Order, error) {
	return s.orderRepo.GetUserByID(ctx, userID)
}

func (s *OrderService) UpdateOrderStatus(
	ctx context.Context,
	orderID string,
	newStatus string,
) error {

	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	err = domain.CanTransition(string(order.Status), newStatus)
	if err != nil {
		return err
	}

	return s.orderRepo.UpdateStatus(ctx, orderID, newStatus)
}
