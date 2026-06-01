package broker

import (
	"crypto/tls"
	"log"
	"order-service/internal/config"
	"time"

	"github.com/segmentio/kafka-go"
)

// ─────────────────────────────────────────
// WRITER (producer)
// ─────────────────────────────────────────

func NewKafkaWriter(cfg *config.Config) *kafka.Writer {

	return &kafka.Writer{

		Addr: kafka.TCP(cfg.KafkaBroker),
		//Topic:    "order.created",
		Balancer: &kafka.Hash{}, // same orderID → same partition

		// delivery guarantee
		RequiredAcks: kafka.RequireAll, // wait for all replicas
		MaxAttempts:  3,                // retry 3x on failure
		WriteTimeout: 10 * time.Second, // timeout for each write attempt
		Async:        false,            // synchronous — safe

		// batching for throughput
		BatchSize:    100,                   // max messages per batch
		BatchTimeout: 10 * time.Millisecond, // max wait before flush

		// TLS for production encryption
		Transport: kafkaTransport(cfg),

		// error logging
		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			log.Printf("[kafka-writer] ERROR: "+msg, args...)
		}),

		// completion hook — log every publish result
		Completion: func(messages []kafka.Message, err error) {
			if err != nil {
				log.Printf("[kafka-writer] failed to deliver %d messages: %v", len(messages), err)
			}

		},
	}
}

// ─────────────────────────────────────────e
// READER (consumer)
// ─────────────────────────────────────────

func NewKafkaReader(cfg *config.Config, groupID string) *kafka.Reader {

	return kafka.NewReader(kafka.ReaderConfig{

		Brokers: []string{cfg.KafkaBroker},
		//Topic:   "order.created",
		GroupID: groupID, // horizontal scaling — partitions shared across instances

		//batch size control
		MinBytes: 10e3, // wait until 10KB ready — fewer round trips
		MaxBytes: 10e6, // max 10MB per fetch — prevent memory spike

		//offset control
		StartOffset: kafka.FirstOffset, // replay from beginning if new consumer group

		//commit control — prevent duplicate processing
		CommitInterval: 1 * time.Second, // auto-commit offset every 1 second

		//timeouts
		MaxWait:         500 * time.Millisecond, // max time to wait for MinBytes
		ReadLagInterval: 10 * time.Second,       // how often to check consumer lag

		//TLS for production
		Dialer: kafkaDialer(cfg),

		//error logging
		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			log.Printf("[kafka-reader] ERROR: "+msg, args...)
		}),

		//info logging — log partition assignments
		Logger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			log.Printf("[kafka-reader] INFO: "+msg, args...)
		}),
	})

}

// ─────────────────────────────────────────
// TLS — encrypted connection (production)
// ─────────────────────────────────────────

func kafkaTransport(cfg *config.Config) *kafka.Transport {
	if !cfg.KafkaTLSEnabled {
		return nil // no TLS in local dev
	}
	return &kafka.Transport{
		TLS: &tls.Config{
			InsecureSkipVerify: false, // always verify in production
			MinVersion:         tls.VersionTLS12,
		},
	}
}

func kafkaDialer(cfg *config.Config) *kafka.Dialer {
	if !cfg.KafkaTLSEnabled {
		return kafka.DefaultDialer // plain connection for local dev
	}
	return &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
		TLS: &tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
		},
	}
}

// ─────────────────────────────────────────
// CLOSER — graceful shutdown
// ─────────────────────────────────────────

func CloseWriter(w *kafka.Writer) {
	if err := w.Close(); err != nil {
		log.Printf("[kafka-writer] error closing writer: %v", err)
	}
}

func CloseReader(r *kafka.Reader) {
	if err := r.Close(); err != nil {
		log.Printf("[kafka-reader] error closing reader: %v", err)
	}
}
