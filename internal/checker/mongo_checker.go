package checker

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

const (
	defaultMongoDegradedThreshold = 200 * time.Millisecond
	defaultMongoUnhealthyTimeout  = 2 * time.Second
)

type MongoChecker struct {
	client            *mongo.Client
	name              string
	degradedThreshold time.Duration
	unhealthyTimeout  time.Duration
}

type MongoCheckerOption func(*MongoChecker)

func WithMongoDegradedThreshold(d time.Duration) MongoCheckerOption {
	return func(c *MongoChecker) { c.degradedThreshold = d }
}

func WithMongoUnhealthyTimeout(d time.Duration) MongoCheckerOption {
	return func(c *MongoChecker) { c.unhealthyTimeout = d }
}

func NewMongoChecker(client *mongo.Client, opts ...MongoCheckerOption) *MongoChecker {
	c := &MongoChecker{
		client:            client,
		name:              "mongodb",
		degradedThreshold: defaultMongoDegradedThreshold,
		unhealthyTimeout:  defaultMongoUnhealthyTimeout,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (m *MongoChecker) Name() string { return m.name }

func (m *MongoChecker) Check(ctx context.Context) CheckResult {
	pingCtx, cancel := context.WithTimeout(ctx, m.unhealthyTimeout)
	defer cancel()

	start := time.Now()
	err := m.client.Ping(pingCtx, nil)
	latencyMs := time.Since(start).Milliseconds()

	if err != nil {
		return CheckResult{
			Status:    StatusUnhealthy,
			Message:   err.Error(),
			LatencyMs: latencyMs,
			CheckedAt: time.Now().UTC(),
		}
	}

	if time.Duration(latencyMs)*time.Millisecond > m.degradedThreshold {
		return CheckResult{
			Status:    StatusDegraded,
			Message:   "ping latency above threshold",
			LatencyMs: latencyMs,
			CheckedAt: time.Now().UTC(),
		}
	}

	return CheckResult{
		Status:    StatusHealthy,
		LatencyMs: latencyMs,
		CheckedAt: time.Now().UTC(),
	}
}
