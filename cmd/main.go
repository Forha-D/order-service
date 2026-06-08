package main

import (
	"context"
	"log"

	"order-service/internal/broker"
	"order-service/internal/checker"
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

	// Repositories
	orderRepo := repository.NewOrderRepository(db)
	outboxRepo := repository.NewOutboxRepository(db)
	idempotencyRepo := repository.NewIdempotencyRepository(db)

	// ── Ensure MongoDB indexes ─────────────────────────────────
	// ── Cancellable context for background workers ─────────────
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := orderRepo.EnsureIndexes(ctx); err != nil {
		log.Fatalf("failed to ensure order indexes: %v", err)
	}
	if err := idempotencyRepo.EnsureIndexes(ctx); err != nil {
		log.Fatalf("failed to ensure idempotency indexes: %v", err)
	}
	log.Println("mongodb indexes ensured")

	// Services
	orderService := service.NewOrderService(orderRepo, outboxRepo, idempotencyRepo)

	// Handlers
	orderHandler := handler.NewOrderHandler(orderService)

	// ── App server (port 8080) — Kong routes this ──────────────
	app := echo.New()
	app.HideBanner = true
	routes.RegisterOrderRoutes(app, orderHandler)

	// ── Health server (port 9090) — K8s direct, Kong never sees this ──
	healthServer := echo.New()
	healthServer.HideBanner = true

	healthHandler := handler.NewHealthHandler(
		cfg.AppVersion,
		cfg.AppName,
		checker.NewMongoChecker(client),
		checker.NewKafkaChecker(cfg.KafkaBroker),
	)
	healthHandler.RegisterRoutes(healthServer) // internal group for detailed health, separate from public probes

	// Kafka
	kafkaWriter := broker.NewKafkaWriter(cfg)
	defer broker.CloseWriter(kafkaWriter)

	kafkaPublisher := publisher.NewKafkaPublisher(kafkaWriter)

	outboxWorker := worker.NewOutboxWorker(outboxRepo, kafkaPublisher, cfg)
	go outboxWorker.Start(ctx)

	// ── Start health server in background ─────────────────────
	go func() {
		log.Printf("health server starting on :%s", cfg.HealthPort)
		if err := healthServer.Start(":" + cfg.HealthPort); err != nil {
			healthServer.Logger.Fatal(err)
		}
	}()

	// ── Start app server (blocking) ───────────────────────────
	log.Printf("starting %s on port %s", cfg.AppName, cfg.Port)
	app.Logger.Fatal(app.Start(":" + cfg.Port))
}
