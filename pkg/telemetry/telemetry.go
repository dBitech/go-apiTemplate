package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/dBiTech/go-apiTemplate/pkg/logger"
)

// Telemetry holds the tracer provider and other telemetry components
type Telemetry struct {
	tracerProvider *sdktrace.TracerProvider
	log            logger.Logger
}

// Config holds the configuration for telemetry
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	Endpoint       string
	Enabled        bool
}

// New creates a new telemetry instance
func New(ctx context.Context, cfg Config, log logger.Logger) (*Telemetry, error) {
	if !cfg.Enabled {
		log.Info("telemetry is disabled")
		return &Telemetry{log: log}, nil
	}

	log.Info("initializing telemetry",
		logger.String("serviceName", cfg.ServiceName),
		logger.String("endpoint", cfg.Endpoint))

	// Create a resource describing the service
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create OTLP exporter
	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
		otlptracegrpc.WithInsecure(),
	)

	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, err
	}

	// Create trace provider
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set global trace provider
	otel.SetTracerProvider(tracerProvider)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Telemetry{
		tracerProvider: tracerProvider,
		log:            log,
	}, nil
}

// Shutdown shuts down the tracer provider
func (t *Telemetry) Shutdown(ctx context.Context) error {
	if t.tracerProvider == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	if err := t.tracerProvider.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}

// Tracer returns a tracer instance
func (t *Telemetry) Tracer(name string) trace.Tracer {
	if t.tracerProvider != nil {
		return t.tracerProvider.Tracer(name)
	}
	return otel.Tracer(name)
}
