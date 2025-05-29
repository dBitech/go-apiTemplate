// Package middleware provides HTTP middleware components for request processing.
// It includes authentication, logging, CORS, and other cross-cutting concerns.
package middleware

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/dBiTech/go-apiTemplate/pkg/logger"
	"github.com/dBiTech/go-apiTemplate/pkg/metrics"
	"github.com/dBiTech/go-apiTemplate/pkg/telemetry"
)

// RequestIDKey is the context key for the request ID
const RequestIDKey = "request_id"

// RequestLogger adds request logging
func RequestLogger(log logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Generate request ID if not present
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
				r.Header.Set("X-Request-ID", requestID)
			}

			// Set request ID in response
			w.Header().Set("X-Request-ID", requestID)

			// Create a logger with request context
			reqLogger := log.With(
				logger.String("request_id", requestID),
				logger.String("method", r.Method),
				logger.String("path", r.URL.Path),
				logger.String("remote_addr", r.RemoteAddr),
				logger.String("user_agent", r.UserAgent()),
			)

			// Add logger to context
			ctx := logger.ToContext(r.Context(), reqLogger)
			r = r.WithContext(ctx)

			// Create response wrapper to capture status
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Log request start
			reqLogger.Info("request started")

			// Process request
			next.ServeHTTP(rw, r)

			// Calculate duration
			duration := time.Since(start)

			// Log request completion
			reqLogger.Info("request completed",
				logger.Int("status", rw.statusCode),
				logger.Duration("duration", duration),
				logger.Int("response_size", rw.size),
			)
		})
	}
}

// Metrics adds prometheus metrics
func Metrics(m *metrics.Metrics) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return m.InstrumentHandler(next)
	}
}

// Tracing adds OpenTelemetry tracing
func Tracing(tel *telemetry.Telemetry) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Start a span
			tracer := tel.Tracer("http")
			ctx, span := tracer.Start(r.Context(), r.Method+" "+r.URL.Path)
			defer span.End()

			// Add HTTP details to span
			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.url", r.URL.String()),
				attribute.String("http.host", r.Host),
				attribute.String("http.user_agent", r.UserAgent()),
			)

			// Add request ID to span
			requestID := r.Header.Get("X-Request-ID")
			if requestID != "" {
				span.SetAttributes(attribute.String("request_id", requestID))
			}

			// Create response wrapper to capture status
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Process request with tracing context
			next.ServeHTTP(rw, r.WithContext(ctx))

			// Add response details to span
			span.SetAttributes(attribute.Int("http.status_code", rw.statusCode))
			if rw.statusCode >= 400 {
				span.SetStatus(codes.Error, http.StatusText(rw.statusCode))
			}
		})
	}
}

// Recover middleware handles panics
func Recover(log logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the error
					log.Error("panic recovered",
						logger.Any("error", err),
					)

					// Extract span from context
					span := trace.SpanFromContext(r.Context())
					span.SetStatus(codes.Error, "panic")
					span.RecordError(err.(error))

					// Return 500 Internal Server Error
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// CORS middleware handles Cross-Origin Resource Sharing
func CORS(allowedOrigins []string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			if len(allowedOrigins) == 0 || allowedOrigins[0] == "*" {
				allowed = true
			} else {
				for _, allowedOrigin := range allowedOrigins {
					if origin == allowedOrigin {
						allowed = true
						break
					}
				}
			}

			// Set CORS headers if origin is allowed
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
			}

			// Handle preflight request
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter is a wrapper for http.ResponseWriter that tracks status code and size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the response size
func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// Flush implements http.Flusher if the underlying ResponseWriter supports it
func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
