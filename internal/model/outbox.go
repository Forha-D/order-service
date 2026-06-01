package model

import "time"

const (
	OutboxStatusPending = "PENDING"
	OutboxStatusSent    = "SENT"
	OutboxStatusDead    = "DEAD"
)

type OutboxEvent struct {
	ID        string `bson:"_id,omitempty"`
	EventType string `bson:"event_type"`
	Payload   string `bson:"payload"` // JSON string

	RetryCount int    `bson:"retry_count"`
	LastError  string `bson:"last_error,omitempty"`

	Status      string    `bson:"status"` // PENDING, SENT, DEAD
	CreatedAt   time.Time `bson:"created_at"`
	ProcessedAt time.Time `bson:"processed_at,omitempty"`
}
