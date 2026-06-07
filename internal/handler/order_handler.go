package handler

import (
	"net/http"

	"order-service/internal/dto"
	"order-service/internal/service"

	"github.com/labstack/echo/v4"
)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) CreateOrder(c echo.Context) error {
	userID, ok := c.Get("userID").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
	}

	var req dto.CreateOrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	// idempotency key comes from header into the DTO — service handles the logic
	req.IdempotencyKey = c.Request().Header.Get("Idempotency-Key")

	order, err := h.orderService.CreateOrder(c.Request().Context(), userID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, dto.OrderResponse{
		ID:          order.ID.Hex(),
		UserID:      order.UserID,
		Status:      order.Status,
		TotalAmount: order.TotalAmount,
	})
}

func (h *OrderHandler) GetOrderByID(c echo.Context) error {
	orderID := c.Param("id")
	if orderID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "missing order id",
		})
	}

	order, err := h.orderService.GetOrderByID(c.Request().Context(), orderID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "order not found",
		})
	}

	return c.JSON(http.StatusOK, dto.OrderResponse{
		ID:          order.ID.Hex(),
		UserID:      order.UserID,
		Status:      order.Status,
		TotalAmount: order.TotalAmount,
	})
}

func (h *OrderHandler) GetOrdersByUser(c echo.Context) error {
	userID, ok := c.Get("userID").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
	}

	orders, err := h.orderService.GetOrderByUserID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to fetch orders",
		})
	}

	response := make([]dto.OrderResponse, 0, len(orders))
	for _, order := range orders {
		response = append(response, dto.OrderResponse{
			ID:          order.ID.Hex(),
			UserID:      order.UserID,
			Status:      order.Status,
			TotalAmount: order.TotalAmount,
		})
	}

	return c.JSON(http.StatusOK, response)
}

func (h *OrderHandler) UpdateOrderStatus(c echo.Context) error {
	orderID := c.Param("id")
	if orderID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "missing order id",
		})
	}

	var req dto.UpdateOrderStatusRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	if req.Status == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "status is required",
		})
	}

	if err := h.orderService.UpdateOrderStatus(c.Request().Context(), orderID, req.Status); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "order status updated successfully",
	})
}
