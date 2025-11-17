package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

// GinMiddleware creates a Gin middleware for OpenTelemetry
func (t *Telemetry) GinMiddleware() gin.HandlerFunc {
	if !t.Config.Enabled {
		return gin.Logger()
	}

	return otelgin.Middleware(t.Config.ServiceName, otelgin.WithTracerProvider(t.tracerProvider))
}

// GormPlugin creates a GORM plugin for OpenTelemetry
// Note: Disabled until otelgorm package is available
func (t *Telemetry) GormPlugin() gorm.Plugin {
	// if !t.Config.Enabled {
	// 	return nil
	// }

	// return otelgorm.NewPlugin(
	// 	otelgorm.WithTracerProvider(t.tracerProvider),
	// 	otelgorm.WithDBName("postgres"),
	// )
	return nil
}

// KafkaInstrumentation creates Kafka instrumentation
// Note: Disabled until otelkafka package is available
// func (t *Telemetry) KafkaInstrumentation() *otelkafka.Transport {
// 	if !t.Config.Enabled {
// 		return nil
// 	}

// 	return otelkafka.NewTransport(
// 		otelkafka.WithTracerProvider(t.tracerProvider),
// 	)
// }

// Custom HTTP Middleware with Business Metrics
func (t *Telemetry) HTTPMiddlewareWithMetrics(businessMetrics *BusinessMetrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !t.Config.Enabled {
			c.Next()
			return
		}

		start := time.Now()

		// Start span
		ctx, span := t.StartSpan(c.Request.Context(), "http.request",
			trace.WithAttributes(
				attribute.String("http.method", c.Request.Method),
				attribute.String("http.url", c.Request.URL.String()),
				attribute.String("http.user_agent", c.Request.UserAgent()),
			),
		)
		defer span.End()

		// Set context
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		status := fmt.Sprintf("%d", c.Writer.Status())

		businessMetrics.RecordHTTPRequest(ctx, c.Request.Method, c.Request.URL.Path, status, duration)

		if c.Writer.Status() >= 400 {
			businessMetrics.RecordHTTPError(ctx, c.Request.Method, c.Request.URL.Path, status)
		}

		// Add span attributes
		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.String("http.status_text", http.StatusText(c.Writer.Status())),
			attribute.Float64("http.duration", duration),
		)
	}
}

// Database Middleware with Business Metrics
func (t *Telemetry) DatabaseMiddlewareWithMetrics(businessMetrics *BusinessMetrics) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		if !t.Config.Enabled {
			return
		}

		ctx := db.Statement.Context
		start := time.Now()

		// Start span
		ctx, span := t.StartSpan(ctx, "db.operation",
			trace.WithAttributes(
				attribute.String("db.operation", db.Statement.SQL.String()),
				attribute.String("db.table", db.Statement.Table),
			),
		)
		defer span.End()

		// Set context
		db.Statement.Context = ctx

		// Execute operation
		_, err := db.Statement.ConnPool.QueryContext(ctx, db.Statement.SQL.String(), db.Statement.Vars...)
		if err != nil {
			t.SetSpanError(ctx, err)
		}

		// Record metrics
		duration := time.Since(start).Seconds()
		businessMetrics.RecordDBOperation(ctx, "query", db.Statement.Table, duration)
	}
}

// Cache Middleware with Business Metrics
func (t *Telemetry) CacheMiddlewareWithMetrics(businessMetrics *BusinessMetrics) func(ctx context.Context, operation, key string, duration time.Duration, err error) {
	return func(ctx context.Context, operation, key string, duration time.Duration, err error) {
		if !t.Config.Enabled {
			return
		}

		// Start span
		ctx, span := t.StartSpan(ctx, "cache.operation",
			trace.WithAttributes(
				attribute.String("cache.operation", operation),
				attribute.String("cache.key", key),
			),
		)
		defer span.End()

		// Record metrics
		businessMetrics.RecordCacheOperation(ctx, operation, "redis", duration.Seconds())

		if err != nil {
			t.SetSpanError(ctx, err)
		} else {
			if operation == "get" {
				businessMetrics.RecordCacheHit(ctx, "redis", key)
			}
		}
	}
}

// Kafka Middleware with Business Metrics
func (t *Telemetry) KafkaMiddlewareWithMetrics(businessMetrics *BusinessMetrics) func(ctx context.Context, topic string, duration time.Duration, err error) {
	return func(ctx context.Context, topic string, duration time.Duration, err error) {
		if !t.Config.Enabled {
			return
		}

		// Start span
		ctx, span := t.StartSpan(ctx, "kafka.publish",
			trace.WithAttributes(
				attribute.String("messaging.system", "kafka"),
				attribute.String("messaging.destination", topic),
			),
		)
		defer span.End()

		// Record metrics
		businessMetrics.RecordKafkaMessageProduced(ctx, topic, duration.Seconds())

		if err != nil {
			t.SetSpanError(ctx, err)
		}
	}
}
