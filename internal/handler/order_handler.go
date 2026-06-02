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

	return &OrderHandler{
		orderService: orderService,
	}
}

func (h *OrderHandler) CreateOrder(c echo.Context) error {

	// Get user from middleware (Echo context)
	userID := c.Get("userID").(string)

	var req dto.CreateOrderRequest

	// Bind request body
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	// Call service
	order, err := h.orderService.CreateOrder(
		c.Request().Context(),
		userID,
		req,
	)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// Response DTO
	response := dto.OrderResponse{
		ID:          order.ID.Hex(),
		UserID:      order.UserID,
		Status:      order.Status,
		TotalAmount: order.TotalAmount,
	}

	return c.JSON(http.StatusCreated, response)
}

func (h *OrderHandler) GetOrderByID(c echo.Context) error {

	// Get order ID from URL param
	orderID := c.Param("id")

	if orderID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "missing order id",
		})
	}

	// Call service
	order, err := h.orderService.GetOrderByID(
		c.Request().Context(),
		orderID,
	)

	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "order not found",
		})
	}

	// Response DTO
	response := dto.OrderResponse{
		ID:          order.ID.Hex(),
		UserID:      order.UserID,
		Status:      order.Status,
		TotalAmount: order.TotalAmount,
	}

	return c.JSON(http.StatusOK, response)
}

func (h *OrderHandler) GetOrdersByUser(c echo.Context) error {

	// 1. Get user from middleware (Kong → Echo context)
	userID := c.Get("userID")
	if userID == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "unauthorized user",
		})
	}

	uid := userID.(string)

	// 2. Call service
	orders, err := h.orderService.GetOrderByUserID(
		c.Request().Context(),
		uid,
	)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to fetch orders",
		})
	}

	// 3. Map response
	var response []dto.OrderResponse

	for _, order := range orders {
		response = append(response, dto.OrderResponse{
			ID:          order.ID.Hex(),
			UserID:      order.UserID,
			Status:      order.Status,
			TotalAmount: order.TotalAmount,
		})
	}

	// 4. Return JSON
	return c.JSON(http.StatusOK, response)
}

func (h *OrderHandler) UpdateOrderStatus(c echo.Context) error {

	// 1. Get order ID from path
	orderID := c.Param("id")
	if orderID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "missing order id",
		})
	}

	// 2. Request DTO
	var req dto.UpdateOrderStatusRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	// 3. Basic validation
	if req.Status == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "status is required",
		})
	}

	// 4. Call service layer
	err := h.orderService.UpdateOrderStatus(
		c.Request().Context(),
		orderID,
		req.Status,
	)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// 5. Success response
	return c.JSON(http.StatusOK, map[string]string{
		"message": "order status updated successfully",
	})
}
