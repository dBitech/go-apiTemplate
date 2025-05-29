// Package metrics provides Prometheus metrics collection and monitoring functionality.
// It includes metric registration, collection, and HTTP endpoint exposure for monitoring systems.
package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all metrics instances
type Metrics struct {
	registry             *prometheus.Registry
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight *prometheus.GaugeVec
	httpResponseSize     *prometheus.HistogramVec
	httpRequestSize      *prometheus.HistogramVec
}

// NewMetrics creates a new metrics instance
func NewMetrics(namespace string) *Metrics {
	registry := prometheus.NewRegistry()

	httpRequestsTotal := promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration := promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "Duration of HTTP requests in seconds.",
			Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1, 2, 5, 10},
		},
		[]string{"method", "path", "status"},
	)

	httpRequestsInFlight := promauto.With(registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "http_requests_in_flight",
			Help:      "Current number of HTTP requests being processed.",
		},
		[]string{"method", "path"},
	)

	httpResponseSize := promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_response_size_bytes",
			Help:      "Size of HTTP responses in bytes.",
			Buckets:   []float64{100, 1000, 10000, 100000, 1000000},
		},
		[]string{"method", "path", "status"},
	)

	httpRequestSize := promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_size_bytes",
			Help:      "Size of HTTP requests in bytes.",
			Buckets:   []float64{100, 1000, 10000, 100000, 1000000},
		},
		[]string{"method", "path"},
	)

	// Register default Go collectors
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	return &Metrics{
		registry:             registry,
		httpRequestsTotal:    httpRequestsTotal,
		httpRequestDuration:  httpRequestDuration,
		httpRequestsInFlight: httpRequestsInFlight,
		httpResponseSize:     httpResponseSize,
		httpRequestSize:      httpRequestSize,
	}
}

// Handler returns an HTTP handler for metrics endpoint
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// InstrumentHandler wraps an HTTP handler with metrics collection
func (m *Metrics) InstrumentHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method

		// Track request size
		requestSize := computeApproximateRequestSize(r)
		m.httpRequestSize.WithLabelValues(method, path).Observe(float64(requestSize))

		// Track in-flight requests
		m.httpRequestsInFlight.WithLabelValues(method, path).Inc()
		defer m.httpRequestsInFlight.WithLabelValues(method, path).Dec()

		// Track response size and status code
		rw := newResponseWriter(w)
		startTime := time.Now()

		next.ServeHTTP(rw, r)

		duration := time.Since(startTime).Seconds()
		statusCode := strconv.Itoa(rw.statusCode)

		m.httpRequestsTotal.WithLabelValues(method, path, statusCode).Inc()
		m.httpRequestDuration.WithLabelValues(method, path, statusCode).Observe(duration)
		m.httpResponseSize.WithLabelValues(method, path, statusCode).Observe(float64(rw.size))
	})
}

// responseWriter is a wrapper for http.ResponseWriter that stores status code and response size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// computeApproximateRequestSize returns the approximate request size in bytes
func computeApproximateRequestSize(r *http.Request) int {
	size := 0

	// Method and URL
	size += len(r.Method)
	if r.URL != nil {
		size += len(r.URL.String())
	}

	// Headers
	for name, values := range r.Header {
		size += len(name)
		for _, value := range values {
			size += len(value)
		}
	}

	// Host
	size += len(r.Host)

	// Content length
	if r.ContentLength > 0 {
		size += int(r.ContentLength)
	}

	return size
}
