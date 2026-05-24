package main

import (
	"order-service/internal/broker"
	"order-service/internal/config"

	"log"

	"github.com/labstack/echo/v4/middleware"

	"github.com/labstack/echo/v4"
)

func main() {

	// Load configuration
	cfg := config.LoadConfig()

	// create writer
	kafkaWriter := broker.NewKafkaWriter(cfg)
	defer broker.CloseWriter(kafkaWriter)

	// create reader
	kafkaReader := broker.NewKafkaReader(cfg, "order-service-group")
	defer broker.CloseReader(kafkaReader)

	//  create echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	// allow frontend to call your API

	// e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
	// 	AllowOrigins: []string{"http://localhost:3000"},
	// 	AllowMethods: []string{"GET", "POST", "PATCH", "PUT", "DELETE"},
	// 	AllowHeaders: []string{"Authorization", "Content-Type"},
	// }))

	// health checl endpoint
	e.GET("/health", func(c echo.Context) error {

		return c.JSON(200, map[string]string{

			"status":  "healthy",
			"service": cfg.AppName,
		})
	})
	// start server

	log.Printf("%s running on port %s", cfg.AppName, cfg.Port)

	e.Logger.Fatal(e.Start(":" + cfg.Port))

}
