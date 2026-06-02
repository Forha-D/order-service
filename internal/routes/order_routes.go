package routes

import (
	"order-service/internal/handler"
	"order-service/internal/middleware"

	"github.com/labstack/echo/v4"
)

func RegisterOrderRoutes(
	e *echo.Echo,
	orderHandler *handler.OrderHandler,
) {

	orderGroup := e.Group("/api/v1/orders")

	// Kong-authenticated routes
	orderGroup.Use(middleware.AuthMiddleware)

	orderGroup.POST("", orderHandler.CreateOrder)

	orderGroup.GET("/:id", orderHandler.GetOrderByID)

	orderGroup.GET("/my-orders", orderHandler.GetOrdersByUser)

	orderGroup.PATCH("/:id/status", orderHandler.UpdateOrderStatus)
}
