package main

import (
	"context"
	"log"

	"order-service/internal/broker"
	"order-service/internal/config"
	"order-service/internal/handler"
	"order-service/internal/publisher"
	"order-service/internal/repository"
	"order-service/internal/routes"
	"order-service/internal/service"

	"order-service/internal/worker"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {

	// Load Config
	cfg := config.LoadConfig()

	// MongoDB connection
	client, err := mongo.Connect(
		options.Client().ApplyURI(cfg.MongoURI),
	)

	if err != nil {
		log.Fatal(err)
	}

	// DB instance
	db := client.Database(cfg.DatabaseName)

	// Repositories (DB-level injection style)
	orderRepo := repository.NewOrderRepository(db)
	outboxRepo := repository.NewOutboxRepository(db)

	// Services
	orderService := service.NewOrderService(
		orderRepo,
		outboxRepo,
	)

	// Handlers
	orderHandler := handler.NewOrderHandler(orderService)

	// Echo instance
	e := echo.New()

	// Routes
	routes.RegisterOrderRoutes(e, orderHandler)

	// Kafka Writer (for outbox events)
	kafkaWriter := broker.NewKafkaWriter(cfg)
	defer broker.CloseWriter(kafkaWriter)

	// Kafka Publisher (for outbox events)
	kafkaPublisher := publisher.NewKafkaPublisher(kafkaWriter)

	outboxWorker := worker.NewOutboxWorker(
		outboxRepo,
		kafkaPublisher,
		cfg,
	)
	go outboxWorker.Start(context.Background())

	// Start server
	log.Printf("starting %s on port %s", cfg.AppName, cfg.Port)

	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
