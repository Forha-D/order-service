package handler

import (
	"context"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/labstack/echo/v4"

	"order-service/internal/checker"
)

type HealthHandler struct {
	checkers  []checker.Checker
	startTime time.Time
	version   string
	service   string
}

// PublicReadyResponse — safe to expose externally, no dep details
type PublicReadyResponse struct {
	Status    checker.Status `json:"status"`
	Service   string         `json:"service"`
	Timestamp string         `json:"timestamp"`
}

// DetailedHealthResponse — internal/ops only, full dep breakdown
type DetailedHealthResponse struct {
	Status    checker.Status                 `json:"status"`
	Service   string                         `json:"service"`
	Version   string                         `json:"version"`
	Uptime    string                         `json:"uptime"`
	Timestamp string                         `json:"timestamp"`
	Checks    map[string]checker.CheckResult `json:"checks"`
	System    SystemInfo                     `json:"system"`
}

type SystemInfo struct {
	GoVersion  string `json:"go_version"`
	GoRoutines int    `json:"goroutines"`
	MemAllocMB uint64 `json:"mem_alloc_mb"`
}

type LiveResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

func NewHealthHandler(version, service string, checkers ...checker.Checker) *HealthHandler {
	return &HealthHandler{
		checkers:  checkers,
		startTime: time.Now(),
		version:   version,
		service:   service,
	}
}

// public — safe for K8s probes and load balancers
func (h *HealthHandler) RegisterRoutes(health *echo.Echo) {
	health.GET("/health/live", h.Live)
	health.GET("/health/ready", h.Ready)
	health.GET("/health/detailed", h.Detailed)
}

// Live — is the process alive? no dep checks
func (h *HealthHandler) Live(c echo.Context) error {
	return c.JSON(http.StatusOK, LiveResponse{
		Status:  "ok",
		Service: h.service,
	})
}

// Ready — minimal public response, used by K8s readiness probe
func (h *HealthHandler) Ready(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	results := h.runChecks(ctx)
	overallStatus, httpStatus := computeOverallStatus(results)

	return c.JSON(httpStatus, PublicReadyResponse{
		Status:    overallStatus,
		Service:   h.service,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// Detailed — full dep breakdown, internal/ops only
func (h *HealthHandler) Detailed(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	results := h.runChecks(ctx)
	overallStatus, httpStatus := computeOverallStatus(results)

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return c.JSON(httpStatus, DetailedHealthResponse{
		Status:    overallStatus,
		Service:   h.service,
		Version:   h.version,
		Uptime:    time.Since(h.startTime).Round(time.Second).String(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Checks:    results,
		System: SystemInfo{
			GoVersion:  runtime.Version(),
			GoRoutines: runtime.NumGoroutine(),
			MemAllocMB: memStats.Alloc / 1024 / 1024,
		},
	})
}

// computeOverallStatus derives status and HTTP code from all check results.
// rules:
//
//	any unhealthy  → unhealthy + 503
//	any degraded   → degraded  + 200  (stay in LB rotation)
//	all healthy    → healthy   + 200
func computeOverallStatus(results map[string]checker.CheckResult) (checker.Status, int) {
	status := checker.StatusHealthy

	for _, result := range results {
		switch result.Status {
		case checker.StatusUnhealthy:
			// unhealthy beats everything — no need to keep scanning
			return checker.StatusUnhealthy, http.StatusServiceUnavailable
		case checker.StatusDegraded:
			// keep scanning — a later result might be unhealthy
			status = checker.StatusDegraded
		}
	}

	return status, http.StatusOK
}

func (h *HealthHandler) runChecks(ctx context.Context) map[string]checker.CheckResult {
	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		results = make(map[string]checker.CheckResult, len(h.checkers))
	)

	for _, c := range h.checkers {
		wg.Add(1)
		go func(c checker.Checker) {
			defer wg.Done()
			result := c.Check(ctx)
			mu.Lock()
			results[c.Name()] = result
			mu.Unlock()
		}(c)
	}

	wg.Wait()
	return results
}
