package checker

import (
	"context"
	"time"
)

type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
)

type CheckResult struct {
	Status    Status    `json:"status"`
	Message   string    `json:"message,omitempty"`
	LatencyMs int64     `json:"latency_ms"` // int64  — easier to threshold on and visualize in monitoring tools
	CheckedAt time.Time `json:"checked_at"`
}

type Checker interface {
	Name() string
	Check(ctx context.Context) CheckResult
}
