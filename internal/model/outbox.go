package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	OutboxStatusPending = "PENDING"
	OutboxStatusSent    = "SENT"
	OutboxStatusDead    = "DEAD"
)

type OutboxEvent struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	EventType   string             `bson:"event_type"`
	Payload     string             `bson:"payload"`
	RetryCount  int                `bson:"retry_count"`
	LastError   string             `bson:"last_error,omitempty"`
	Status      string             `bson:"status"`
	CreatedAt   time.Time          `bson:"created_at"`
	ProcessedAt *time.Time         `bson:"processed_at,omitempty"` // pointer — omits when nil
}
