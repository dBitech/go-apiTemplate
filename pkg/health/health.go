// Package health provides health check functionality for the application.
// It includes health status monitoring, dependency checks, and HTTP health endpoints.
package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/dBiTech/go-apiTemplate/pkg/logger"
)

// Status represents the current status of a component
type Status string

const (
	// StatusUp indicates component is functioning normally
	StatusUp Status = "UP"

	// StatusDown indicates component is not functioning
	StatusDown Status = "DOWN"

	// StatusDegraded indicates component is functioning but with issues
	StatusDegraded Status = "DEGRADED"
)

// Component represents a health check for a system component
type Component struct {
	Name        string                 `json:"name"`
	Status      Status                 `json:"status"`
	Description string                 `json:"description,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
	LastChecked time.Time              `json:"lastChecked"`
}

// Check is a function that performs a health check on a component
type Check func(ctx context.Context) Component

// Checker provides health/readiness/liveness endpoints
type Checker struct {
	appName     string
	version     string
	description string
	checks      []Check
	mu          sync.RWMutex
	cache       *StatusResponse
	cacheTTL    time.Duration
	lastUpdate  time.Time
	log         logger.Logger // Add logger for error handling
}

// StatusResponse represents the overall health status of the service
type StatusResponse struct {
	Name        string      `json:"name"`
	Version     string      `json:"version"`
	Description string      `json:"description,omitempty"`
	Status      Status      `json:"status"`
	Components  []Component `json:"components,omitempty"`
	Timestamp   time.Time   `json:"timestamp"`
}

// NewHealthCheck creates a new health check handler
func NewHealthCheck(appName, version, description string, log logger.Logger) *Checker {
	return &Checker{
		appName:     appName,
		version:     version,
		description: description,
		checks:      []Check{},
		cacheTTL:    time.Second * 10,
		log:         log,
	}
}

// AddCheck adds a health check component
func (h *Checker) AddCheck(check Check) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checks = append(h.checks, check)
	h.cache = nil // Invalidate cache
}

// HealthHandler handles the /health endpoint
func (h *Checker) HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		status, httpStatus := h.getHealth(ctx)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpStatus)
		if err := json.NewEncoder(w).Encode(status); err != nil {
			h.log.Error("Failed to encode health status", logger.Error(err))
		}
	}
}

// LivenessHandler handles the /health/liveness endpoint
func (h *Checker) LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		// For liveness, we just return 200 OK if the server is running
		status := &StatusResponse{
			Name:      h.appName,
			Version:   h.version,
			Status:    StatusUp,
			Timestamp: time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(status); err != nil {
			h.log.Error("Failed to encode liveness status", logger.Error(err))
		}
	}
}

// ReadinessHandler handles the /health/readiness endpoint
func (h *Checker) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		status, httpStatus := h.getHealth(ctx)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpStatus)
		if err := json.NewEncoder(w).Encode(status); err != nil {
			h.log.Error("Failed to encode readiness status", logger.Error(err))
		}
	}
}

// getHealth performs health checks and returns the overall status
func (h *Checker) getHealth(ctx context.Context) (*StatusResponse, int) {
	h.mu.RLock()
	cache := h.cache
	lastUpdate := h.lastUpdate
	h.mu.RUnlock()

	// Return cached result if valid
	if cache != nil && time.Since(lastUpdate) < h.cacheTTL {
		return cache, statusToHTTP(cache.Status)
	}

	// Perform health checks
	h.mu.Lock()
	defer h.mu.Unlock()

	// Double check if cache was updated while waiting for the lock
	if h.cache != nil && time.Since(h.lastUpdate) < h.cacheTTL {
		return h.cache, statusToHTTP(h.cache.Status)
	}

	components := make([]Component, 0, len(h.checks))
	status := StatusUp

	// Execute all health checks concurrently
	var wg sync.WaitGroup
	resultCh := make(chan Component, len(h.checks))

	for _, check := range h.checks {
		wg.Add(1)
		go func(c Check) {
			defer wg.Done()
			// Use a timeout to prevent hanging health checks
			ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			resultCh <- c(ctxTimeout)
		}(check)
	}

	// Wait for all checks to complete
	wg.Wait()
	close(resultCh)

	// Collect results
	for component := range resultCh {
		components = append(components, component)
		if component.Status == StatusDown {
			status = StatusDown
		} else if component.Status == StatusDegraded && status == StatusUp {
			status = StatusDegraded
		}
	}

	result := &StatusResponse{
		Name:        h.appName,
		Version:     h.version,
		Description: h.description,
		Status:      status,
		Components:  components,
		Timestamp:   time.Now(),
	}

	// Cache the result
	h.cache = result
	h.lastUpdate = time.Now()

	return result, statusToHTTP(status)
}

// statusToHTTP converts a health status to an HTTP status code
func statusToHTTP(status Status) int {
	switch status {
	case StatusUp:
		return http.StatusOK
	case StatusDegraded:
		return http.StatusOK // Still consider service available even if degraded
	default:
		return http.StatusServiceUnavailable
	}
}

// DBCheck creates a database connection health check
func DBCheck(name string, pingFn func(context.Context) error) Check {
	return func(ctx context.Context) Component {
		start := time.Now()
		err := pingFn(ctx)
		duration := time.Since(start)

		component := Component{
			Name:        name,
			Status:      StatusUp,
			Description: "Database connection is healthy",
			Details: map[string]interface{}{
				"responseTime": duration.String(),
			},
			LastChecked: time.Now(),
		}

		if err != nil {
			component.Status = StatusDown
			component.Description = "Database connection failed"
			component.Details["error"] = err.Error()
		}

		return component
	}
}
