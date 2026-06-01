package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName         string
	Port            string
	MongoURI        string
	DatabaseName    string
	JWTSecret       string
	Environment     string
	KafkaBroker     string
	KafkaDLQTopic   string
	KafkaTLSEnabled bool
}

func LoadConfig() *Config {

	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found, using system environment")
	}

	return &Config{
		AppName:       getEnv("APP_NAME", "order-service"),
		Port:          getEnv("PORT", "8085"),
		MongoURI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DatabaseName:  getEnv("DATABASE_NAME", "orderdb"),
		JWTSecret:     getEnv("JWT_SECRET", ""),
		Environment:   getEnv("ENVIRONMENT", "development"),
		KafkaBroker:   getEnv("KAFKA_BROKER", "localhost:9092"),
		KafkaDLQTopic: getEnv("KAFKA_DLQ_TOPIC", "order.dlq"),
	}
}

func getEnv(key string, defaultValue string) string {

	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	return value
}
