package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// TelemetryInterface defines the common interface for telemetry implementations
type TelemetryInterface interface {
	// Span operations
	StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
	AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue)
	SetSpanError(ctx context.Context, err error)
	SetSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue)

	// Tracer and Meter getters
	GetTracer() trace.Tracer
	GetMeter() metric.Meter

	// Metrics creation
	CreateCounter(name, description string) (metric.Int64Counter, error)
	CreateHistogram(name, description string) (metric.Float64Histogram, error)
	CreateGauge(name, description string) (metric.Float64Gauge, error)

	// Lifecycle
	Shutdown(ctx context.Context) error
}

// Ensure both implementations satisfy the interface
var _ TelemetryInterface = (*Telemetry)(nil)
var _ TelemetryInterface = (*SimpleTelemetry)(nil)


