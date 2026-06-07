package checker

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	defaultKafkaDegradedThreshold = 300 * time.Millisecond // higher than mongo — network hop
	defaultKafkaUnhealthyTimeout  = 3 * time.Second
)

type KafkaChecker struct {
	brokerAddr        string
	name              string
	degradedThreshold time.Duration
	unhealthyTimeout  time.Duration
}

type KafkaCheckerOption func(*KafkaChecker)

func WithKafkaDegradedThreshold(d time.Duration) KafkaCheckerOption {
	return func(c *KafkaChecker) { c.degradedThreshold = d }
}

func WithKafkaUnhealthyTimeout(d time.Duration) KafkaCheckerOption {
	return func(c *KafkaChecker) { c.unhealthyTimeout = d }
}

func NewKafkaChecker(brokerAddr string, opts ...KafkaCheckerOption) *KafkaChecker {
	c := &KafkaChecker{
		brokerAddr:        brokerAddr,
		name:              "kafka",
		degradedThreshold: defaultKafkaDegradedThreshold,
		unhealthyTimeout:  defaultKafkaUnhealthyTimeout,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (k *KafkaChecker) Name() string { return k.name }

func (k *KafkaChecker) Check(ctx context.Context) CheckResult {
	pingCtx, cancel := context.WithTimeout(ctx, k.unhealthyTimeout)
	defer cancel()

	start := time.Now()

	// real broker check — sends a metadata request, not just a TCP dial
	conn, err := kafka.DialContext(pingCtx, "tcp", k.brokerAddr)
	if err != nil {
		return CheckResult{
			Status:    StatusUnhealthy,
			Message:   err.Error(),
			LatencyMs: time.Since(start).Milliseconds(),
			CheckedAt: time.Now().UTC(),
		}
	}
	defer conn.Close()

	// ReadPartitions sends a real metadata request to the broker
	// if the broker is up but not functional, this will fail
	_, err = conn.ReadPartitions()
	latencyMs := time.Since(start).Milliseconds()

	if err != nil {
		return CheckResult{
			Status:    StatusUnhealthy,
			Message:   err.Error(),
			LatencyMs: latencyMs,
			CheckedAt: time.Now().UTC(),
		}
	}

	if time.Duration(latencyMs)*time.Millisecond > k.degradedThreshold {
		return CheckResult{
			Status:    StatusDegraded,
			Message:   "broker response latency above threshold",
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
