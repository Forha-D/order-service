package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IdempotencyStatus string

const (
	IdempotencyInProgress IdempotencyStatus = "IN_PROGRESS"
	IdempotencyCompleted  IdempotencyStatus = "COMPLETED"
	IdempotencyFailed     IdempotencyStatus = "FAILED"
)

type IdempotencyRecord struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	// Client supplied idempotency key
	Key string `bson:"key" json:"key"`

	// SHA256 hash of request body
	RequestHash string `bson:"request_hash" json:"request_hash"`

	// Processing state
	Status IdempotencyStatus `bson:"status" json:"status"`

	// Cached response returned to client
	Response string `bson:"response,omitempty" json:"response,omitempty"`

	// Optional error message
	Error string `bson:"error,omitempty" json:"error,omitempty"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`

	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`

	// TTL cleanup field
	ExpiresAt time.Time `bson:"expires_at" json:"expires_at"`
}
