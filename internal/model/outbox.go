package model

import "time"

type OutboxEvent struct {
	ID          string    `bson:"_id,omitempty"`
	EventType   string    `bson:"event_type"`
	Payload     string    `bson:"payload"` // JSON string
	Status      string    `bson:"status"`  // PENDING, SENT, FAILED
	CreatedAt   time.Time `bson:"created_at"`
	ProcessedAt time.Time `bson:"processed_at,omitempty"`
}
