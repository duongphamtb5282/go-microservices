package telemetry

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	// Note: Jaeger exporter is deprecated, use OTLP instead
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"

	// Note: Prometheus exporter moved to separate module
	// "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// TelemetryConfig holds configuration for OpenTelemetry
type TelemetryConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	JaegerEndpoint string
	OTLPEndpoint   string
	PrometheusPort string
	Enabled        bool
}

// Telemetry provides OpenTelemetry functionality
type Telemetry struct {
	Config         TelemetryConfig
	tracerProvider *sdktrace.TracerProvider
	meterProvider  metric.MeterProvider
	tracer         trace.Tracer
	meter          metric.Meter
}

// NewTelemetry creates a new Telemetry instance
func NewTelemetry(config TelemetryConfig) (*Telemetry, error) {
	if !config.Enabled {
		return &Telemetry{Config: config}, nil
	}

	// Create resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.DeploymentEnvironment(config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Initialize tracer provider
	tp, err := initTracerProvider(context.Background(), res, config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tracer provider: %w", err)
	}

	// Initialize meter provider
	mp, err := initMeterProvider(context.Background(), res, config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize meter provider: %w", err)
	}

	// Set global providers
	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)

	return &Telemetry{
		Config:         config,
		tracerProvider: tp,
		meterProvider:  mp,
		tracer:         tp.Tracer(config.ServiceName),
		meter:          mp.Meter(config.ServiceName),
	}, nil
}

// initTracerProvider initializes the tracer provider
func initTracerProvider(ctx context.Context, res *resource.Resource, config TelemetryConfig) (*sdktrace.TracerProvider, error) {
	var exporters []sdktrace.SpanExporter

	// Jaeger exporter - Disabled (deprecated)
	// if config.JaegerEndpoint != "" {
	// 	jaegerExporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.JaegerEndpoint)))
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to create Jaeger exporter: %w", err)
	// 	}
	// 	exporters = append(exporters, jaegerExporter)
	// }

	// OTLP HTTP exporter (primary)
	if config.OTLPEndpoint != "" {
		// Create HTTP exporter
		otlpExporter, err := otlptracehttp.New(ctx,
			otlptracehttp.WithEndpointURL(config.OTLPEndpoint),
			otlptracehttp.WithInsecure(), // Use insecure connection for local development
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
		}
		exporters = append(exporters, otlpExporter)
	}

	// Jaeger exporter (fallback)
	if config.JaegerEndpoint != "" {
		jaegerExporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.JaegerEndpoint)))
		if err != nil {
			return nil, fmt.Errorf("failed to create Jaeger exporter: %w", err)
		}
		exporters = append(exporters, jaegerExporter)
	}

	// Add stdout exporter for debugging
	stdoutExporter, err := stdouttrace.New(stdouttrace.WithWriter(os.Stdout))
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout exporter: %w", err)
	}
	exporters = append(exporters, stdoutExporter)

	if len(exporters) == 0 {
		return nil, fmt.Errorf("no exporters configured")
	}

	// Create tracer provider with multiple span processors
	tpOpts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	}

	// Add a batch span processor for each exporter
	for _, exporter := range exporters {
		bsp := sdktrace.NewBatchSpanProcessor(exporter)
		tpOpts = append(tpOpts, sdktrace.WithSpanProcessor(bsp))
	}

	tp := sdktrace.NewTracerProvider(tpOpts...)

	return tp, nil
}

// initMeterProvider initializes the meter provider
func initMeterProvider(ctx context.Context, res *resource.Resource, config TelemetryConfig) (metric.MeterProvider, error) {
	// Prometheus exporter - Disabled (moved to separate module)
	// exporter, err := prometheus.New()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create Prometheus exporter: %w", err)
	// }

	// Create meter provider
	// mp := metric.NewMeterProvider(
	// 	metric.WithResource(res),
	// 	metric.WithReader(exporter),
	// )

	// return mp, nil

	// Return global meter provider for now
	return otel.GetMeterProvider(), nil
}

// GetTracer returns the tracer
func (t *Telemetry) GetTracer() trace.Tracer {
	if t.tracer == nil {
		return otel.Tracer(t.Config.ServiceName)
	}
	return t.tracer
}

// GetMeter returns the meter
func (t *Telemetry) GetMeter() metric.Meter {
	if t.meter == nil {
		return otel.Meter(t.Config.ServiceName)
	}
	return t.meter
}

// StartSpan starts a new span
func (t *Telemetry) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if !t.Config.Enabled {
		return ctx, trace.SpanFromContext(ctx)
	}
	return t.GetTracer().Start(ctx, name, opts...)
}

// AddSpanEvent adds an event to the current span
func (t *Telemetry) AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	if !t.Config.Enabled {
		return
	}
	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SetSpanError marks span as error
func (t *Telemetry) SetSpanError(ctx context.Context, err error) {
	if !t.Config.Enabled {
		return
	}
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
	// span.SetStatus(otel.Codes().Error, err.Error())
}

// SetSpanAttributes sets attributes on span
func (t *Telemetry) SetSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	if !t.Config.Enabled {
		return
	}
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// CreateCounter creates a counter metric
func (t *Telemetry) CreateCounter(name, description string) (metric.Int64Counter, error) {
	if !t.Config.Enabled {
		return nil, fmt.Errorf("telemetry not enabled")
	}
	return t.GetMeter().Int64Counter(
		name,
		metric.WithDescription(description),
	)
}

// CreateHistogram creates a histogram metric
func (t *Telemetry) CreateHistogram(name, description string) (metric.Float64Histogram, error) {
	if !t.Config.Enabled {
		return nil, fmt.Errorf("telemetry not enabled")
	}
	return t.GetMeter().Float64Histogram(
		name,
		metric.WithDescription(description),
	)
}

// CreateGauge creates a gauge metric
func (t *Telemetry) CreateGauge(name, description string) (metric.Float64Gauge, error) {
	if !t.Config.Enabled {
		return nil, fmt.Errorf("telemetry not enabled")
	}
	return t.GetMeter().Float64Gauge(
		name,
		metric.WithDescription(description),
	)
}

// Shutdown gracefully shuts down the telemetry
func (t *Telemetry) Shutdown(ctx context.Context) error {
	if !t.Config.Enabled {
		return nil
	}

	var errs []error

	if t.tracerProvider != nil {
		if err := t.tracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown tracer provider: %w", err))
		}
	}

	// Note: Global meter provider doesn't have Shutdown method
	// if t.meterProvider != nil {
	// 	if err := t.meterProvider.Shutdown(ctx); err != nil {
	// 		errs = append(errs, fmt.Errorf("failed to shutdown meter provider: %w", err))
	// 	}
	// }

	if len(errs) > 0 {
		return fmt.Errorf("telemetry shutdown errors: %v", errs)
	}

	return nil
}
