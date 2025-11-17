package telemetry

import (
	"context"
	"fmt"
	"time"

	"backend-core/logging"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// SimpleTelemetry provides a simple telemetry implementation without OpenTelemetry dependencies
type SimpleTelemetry struct {
	Config          TelemetryConfig
	logger          *logging.Logger
	businessMetrics *BusinessMetrics
}

// NewSimpleTelemetry creates a new simple telemetry instance
func NewSimpleTelemetry(config TelemetryConfig, logger *logging.Logger) *SimpleTelemetry {
	businessMetrics := NewNoopBusinessMetrics()

	return &SimpleTelemetry{
		Config:          config,
		logger:          logger,
		businessMetrics: businessMetrics,
	}
}

// StartSpan starts a new span (no-op for simple implementation)
func (t *SimpleTelemetry) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if !t.Config.Enabled {
		return ctx, trace.SpanFromContext(ctx)
	}

	t.logger.Debug("Starting span", logging.String("name", name))
	return ctx, trace.SpanFromContext(ctx)
}

// AddSpanEvent adds an event to the current span (no-op for simple implementation)
func (t *SimpleTelemetry) AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	if !t.Config.Enabled {
		return
	}

	t.logger.Debug("Adding span event", logging.String("name", name))
}

// SetSpanError marks span as error (no-op for simple implementation)
func (t *SimpleTelemetry) SetSpanError(ctx context.Context, err error) {
	if !t.Config.Enabled {
		return
	}

	t.logger.Error("Span error", logging.Error(err))
}

// SetSpanAttributes sets attributes on span (no-op for simple implementation)
func (t *SimpleTelemetry) SetSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	if !t.Config.Enabled {
		return
	}

	t.logger.Debug("Setting span attributes")
}

// GetTracer returns a no-op tracer
func (t *SimpleTelemetry) GetTracer() trace.Tracer {
	return otel.Tracer(t.Config.ServiceName)
}

// GetMeter returns a no-op meter
func (t *SimpleTelemetry) GetMeter() metric.Meter {
	return otel.Meter(t.Config.ServiceName)
}

// CreateCounter creates a counter metric (no-op for simple implementation)
func (t *SimpleTelemetry) CreateCounter(name, description string) (metric.Int64Counter, error) {
	if !t.Config.Enabled {
		return nil, fmt.Errorf("telemetry not enabled")
	}
	return t.GetMeter().Int64Counter(name, metric.WithDescription(description))
}

// CreateHistogram creates a histogram metric (no-op for simple implementation)
func (t *SimpleTelemetry) CreateHistogram(name, description string) (metric.Float64Histogram, error) {
	if !t.Config.Enabled {
		return nil, fmt.Errorf("telemetry not enabled")
	}
	return t.GetMeter().Float64Histogram(name, metric.WithDescription(description))
}

// CreateGauge creates a gauge metric (no-op for simple implementation)
func (t *SimpleTelemetry) CreateGauge(name, description string) (metric.Float64Gauge, error) {
	if !t.Config.Enabled {
		return nil, fmt.Errorf("telemetry not enabled")
	}
	return t.GetMeter().Float64Gauge(name, metric.WithDescription(description))
}

// GinMiddleware creates a simple Gin middleware
func (t *SimpleTelemetry) GinMiddleware() interface{} {
	if !t.Config.Enabled {
		return func(c interface{}) {
			// Simple middleware that just logs
			t.logger.Debug("Processing HTTP request")
		}
	}

	return func(c interface{}) {
		t.logger.Debug("Processing HTTP request with telemetry")
	}
}

// HTTPMiddlewareWithMetrics creates HTTP middleware with business metrics
func (t *SimpleTelemetry) HTTPMiddlewareWithMetrics(businessMetrics *BusinessMetrics) interface{} {
	return func(c interface{}) {
		if !t.Config.Enabled {
			return
		}

		start := time.Now()
		t.logger.Debug("Processing HTTP request with metrics")
		duration := time.Since(start).Seconds()
		t.logger.Debug("HTTP request processed", "duration_seconds", duration)
	}
}

// Shutdown gracefully shuts down the telemetry
func (t *SimpleTelemetry) Shutdown(ctx context.Context) error {
	if !t.Config.Enabled {
		return nil
	}

	t.logger.Info("Shutting down simple telemetry")
	return nil
}
