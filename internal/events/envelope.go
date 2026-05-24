package events

import "time"

// EventEnvelope is the standard structure for ALL Kafka messages
type EventEnvelope struct {
	EventID        string      `json:"event_id"`
	EventType      string      `json:"event_type"`
	Version        string      `json:"version"`
	Source         string      `json:"source"`
	Timestamp      time.Time   `json:"timestamp"`
	CorrelationID  string      `json:"correlation_id"`
	IdempotencyKey string      `json:"idempotency_key"`
	Payload        interface{} `json:"payload"`
}
