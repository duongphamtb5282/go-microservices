package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type contextAwareMetrics interface {
	RecordHTTPRequest(ctx context.Context, method, path, status string, duration float64)
	RecordHTTPError(ctx context.Context, method, path, status string)
	RecordDBOperation(ctx context.Context, operation, table string, duration float64)
	SetDBConnectionsActive(ctx context.Context, count float64)
	RecordCacheHit(ctx context.Context, cacheType, key string)
	RecordCacheMiss(ctx context.Context, cacheType, key string)
	RecordCacheOperation(ctx context.Context, operation, cacheType string, duration float64)
	RecordKafkaMessageProduced(ctx context.Context, topic string, duration float64)
	RecordKafkaMessageConsumed(ctx context.Context, topic string)
	RecordUserCreation(ctx context.Context, username, email string)
	RecordUserRetrieval(ctx context.Context, userID string)
	RecordUserActivation(ctx context.Context, userID string)
	SetMemoryUsage(ctx context.Context, bytes float64)
	SetCPUUsage(ctx context.Context, percent float64)
	SetGoroutinesCount(ctx context.Context, count int64)
}

// ContextlessMetrics exposes a context-agnostic interface for external instrumentation packages.
type ContextlessMetrics interface {
	RecordHTTPRequest(method, path, status string, duration float64)
	RecordHTTPError(method, path, status string)
	RecordDBOperation(operation, table string, duration float64)
	SetDBConnectionsActive(count float64)
	RecordCacheHit(cacheType, key string)
	RecordCacheMiss(cacheType, key string)
	RecordCacheOperation(operation, cacheType string, duration float64)
	RecordKafkaMessageProduced(topic string, duration float64)
	RecordKafkaMessageConsumed(topic string)
	RecordUserCreation(username, email string)
	RecordUserRetrieval(userID string)
	RecordUserActivation(userID string)
	SetMemoryUsage(bytes float64)
	SetCPUUsage(percent float64)
	SetGoroutinesCount(count int64)
}

// BusinessMetrics provides business-specific metrics and bridges context-aware and context-less consumers.
type BusinessMetrics struct {
	recorder contextAwareMetrics

	HTTPRequestCounter    ContextlessMetrics
	HTTPErrorCounter      ContextlessMetrics
	DBOperationCounter    ContextlessMetrics
	CacheHitCounter       ContextlessMetrics
	CacheMissCounter      ContextlessMetrics
	CacheOperationCounter ContextlessMetrics
	KafkaMessageCounter   ContextlessMetrics
	UserCreationCounter   ContextlessMetrics
	UserRetrievalCounter  ContextlessMetrics
	UserActivationCounter ContextlessMetrics
	MemoryGauge           ContextlessMetrics
	CPUGauge              ContextlessMetrics
	GoroutinesGauge       ContextlessMetrics
}

// NewBusinessMetrics creates a new BusinessMetrics instance backed by OpenTelemetry metrics when enabled.
func NewBusinessMetrics(telemetry *Telemetry) (*BusinessMetrics, error) {
	if telemetry == nil || !telemetry.Config.Enabled {
		return &BusinessMetrics{}, nil
	}

	recorder, err := newOTELBusinessMetrics(telemetry)
	if err != nil {
		return nil, err
	}

	adapter := &contextlessAdapter{recorder: recorder}

	return &BusinessMetrics{
		recorder:              recorder,
		HTTPRequestCounter:    adapter,
		HTTPErrorCounter:      adapter,
		DBOperationCounter:    adapter,
		CacheHitCounter:       adapter,
		CacheMissCounter:      adapter,
		CacheOperationCounter: adapter,
		KafkaMessageCounter:   adapter,
		UserCreationCounter:   adapter,
		UserRetrievalCounter:  adapter,
		UserActivationCounter: adapter,
		MemoryGauge:           adapter,
		CPUGauge:              adapter,
		GoroutinesGauge:       adapter,
	}, nil
}

// NewNoopBusinessMetrics returns an empty BusinessMetrics instance.
func NewNoopBusinessMetrics() *BusinessMetrics {
	return &BusinessMetrics{}
}

// RecordHTTPRequest records HTTP request metrics.
func (bm *BusinessMetrics) RecordHTTPRequest(ctx context.Context, method, path, status string, duration float64) {
	if bm == nil {
		return
	}
	if bm.recorder != nil {
		bm.recorder.RecordHTTPRequest(ctx, method, path, status, duration)
		return
	}
	if bm.HTTPRequestCounter != nil {
		bm.HTTPRequestCounter.RecordHTTPRequest(method, path, status, duration)
	}
}

// RecordHTTPError records HTTP error metrics.
func (bm *BusinessMetrics) RecordHTTPError(ctx context.Context, method, path, status string) {
	if bm == nil {
		return
	}
	if bm.recorder != nil {
		bm.recorder.RecordHTTPError(ctx, method, path, status)
		return
	}
	if bm.HTTPErrorCounter != nil {
		bm.HTTPErrorCounter.RecordHTTPError(method, path, status)
	}
}

// RecordDBOperation records database operation metrics.
func (bm *BusinessMetrics) RecordDBOperation(ctx context.Context, operation, table string, duration float64) {
	if bm == nil {
		return
	}
	if bm.recorder != nil {
		bm.recorder.RecordDBOperation(ctx, operation, table, duration)
		return
	}
	if bm.DBOperationCounter != nil {
		bm.DBOperationCounter.RecordDBOperation(operation, table, duration)
	}
}

// SetDBConnectionsActive records the number of active DB connections.
func (bm *BusinessMetrics) SetDBConnectionsActive(ctx context.Context, count float64) {
	if bm == nil {
		return
	}
	if bm.recorder != nil {
		bm.recorder.SetDBConnectionsActive(ctx, count)
		return
	}
	if bm.DBOperationCounter != nil {
		bm.DBOperationCounter.SetDBConnectionsActive(count)
	}
}

// RecordCacheHit records cache hit metrics.
func (bm *BusinessMetrics) RecordCacheHit(ctx context.Context, cacheType, key string) {
	if bm == nil {
		return
	}
	if bm.recorder != nil {
		bm.recorder.RecordCacheHit(ctx, cacheType, key)
		return
	}
	if bm.CacheHitCounter != nil {
		bm.CacheHitCounter.RecordCacheHit(cacheType, key)
	}
}

// RecordCacheMiss records cache miss metrics.
func (bm *BusinessMetrics) RecordCacheMiss(ctx context.Context, cacheType, key string) {
	if bm == nil {
		return
	}
	if bm.recorder != nil {
		bm.recorder.RecordCacheMiss(ctx, cacheType, key)
		return
	}
	if bm.CacheMissCounter != nil {
		bm.CacheMissCounter.RecordCacheMiss(cacheType, key)
	}
}

// RecordCacheOperation records cache operation metrics.
func (bm *BusinessMetrics) RecordCacheOperation(ctx context.Context, operation, cacheType string, duration float64) {
	if bm == nil {
		return
	}
	if bm.recorder != nil {
		bm.recorder.RecordCacheOperation(ctx, operation, cacheType, duration)
		return
	}
	if bm.CacheOperationCounter != nil {
		bm.CacheOperationCounter.RecordCacheOperation(operation, cacheType, duration)
	}
}

// RecordKafkaMessageProduced records Kafka publish metrics.
func (bm *BusinessMetrics) RecordKafkaMessageProduced(ctx context.Context, topic string, duration float64) {
	if bm == nil {
		return
	}
	if bm.recorder != nil {
		bm.recorder.RecordKafkaMessageProduced(ctx, topic, duration)
		return
	}
	if bm.KafkaMessageCounter != nil {
		bm.KafkaMessageCounter.RecordKafkaMessageProduced(topic, duration)
	}
}

// RecordKafkaMessageConsumed records Kafka consumption metrics.
func (bm *BusinessMetrics) RecordKafkaMessageConsumed(ctx context.Context, topic string) {
	if bm == nil {
		return
	}
	if bm.recorder != nil {
		bm.recorder.RecordKafkaMessageConsumed(ctx, topic)
		return
	}
	if bm.KafkaMessageCounter != nil {
		bm.KafkaMessageCounter.RecordKafkaMessageConsumed(topic)
	}
}

// RecordUserCreation records user creation metrics.
func (bm *BusinessMetrics) RecordUserCreation(ctx context.Context, username, email string) {
	if bm == nil {
		return
	}
	if bm.recorder != nil {
		bm.recorder.RecordUserCreation(ctx, username, email)
		return
	}
	if bm.UserCreationCounter != nil {
		bm.UserCreationCounter.RecordUserCreation(username, email)
	}
}

// RecordUserRetrieval records user retrieval metrics.
func (bm *BusinessMetrics) RecordUserRetrieval(ctx context.Context, userID string) {
	if bm == nil {
		return
	}
	if bm.recorder != nil {
		bm.recorder.RecordUserRetrieval(ctx, userID)
		return
	}
	if bm.UserRetrievalCounter != nil {
		bm.UserRetrievalCounter.RecordUserRetrieval(userID)
	}
}

// RecordUserActivation records user activation metrics.
func (bm *BusinessMetrics) RecordUserActivation(ctx context.Context, userID string) {
	if bm == nil {
		return
	}
	if bm.recorder != nil {
		bm.recorder.RecordUserActivation(ctx, userID)
		return
	}
	if bm.UserActivationCounter != nil {
		bm.UserActivationCounter.RecordUserActivation(userID)
	}
}

// SetMemoryUsage records memory usage metrics.
func (bm *BusinessMetrics) SetMemoryUsage(ctx context.Context, bytes float64) {
	if bm == nil {
		return
	}
	if bm.recorder != nil {
		bm.recorder.SetMemoryUsage(ctx, bytes)
		return
	}
	if bm.MemoryGauge != nil {
		bm.MemoryGauge.SetMemoryUsage(bytes)
	}
}

// SetCPUUsage records CPU usage metrics.
func (bm *BusinessMetrics) SetCPUUsage(ctx context.Context, percent float64) {
	if bm == nil {
		return
	}
	if bm.recorder != nil {
		bm.recorder.SetCPUUsage(ctx, percent)
		return
	}
	if bm.CPUGauge != nil {
		bm.CPUGauge.SetCPUUsage(percent)
	}
}

// SetGoroutinesCount records goroutine count metrics.
func (bm *BusinessMetrics) SetGoroutinesCount(ctx context.Context, count int64) {
	if bm == nil {
		return
	}
	if bm.recorder != nil {
		bm.recorder.SetGoroutinesCount(ctx, count)
		return
	}
	if bm.GoroutinesGauge != nil {
		bm.GoroutinesGauge.SetGoroutinesCount(count)
	}
}

type contextlessAdapter struct {
	recorder contextAwareMetrics
}

func (a *contextlessAdapter) RecordHTTPRequest(method, path, status string, duration float64) {
	if a == nil || a.recorder == nil {
		return
	}
	a.recorder.RecordHTTPRequest(context.Background(), method, path, status, duration)
}

func (a *contextlessAdapter) RecordHTTPError(method, path, status string) {
	if a == nil || a.recorder == nil {
		return
	}
	a.recorder.RecordHTTPError(context.Background(), method, path, status)
}

func (a *contextlessAdapter) RecordDBOperation(operation, table string, duration float64) {
	if a == nil || a.recorder == nil {
		return
	}
	a.recorder.RecordDBOperation(context.Background(), operation, table, duration)
}

func (a *contextlessAdapter) SetDBConnectionsActive(count float64) {
	if a == nil || a.recorder == nil {
		return
	}
	a.recorder.SetDBConnectionsActive(context.Background(), count)
}

func (a *contextlessAdapter) RecordCacheHit(cacheType, key string) {
	if a == nil || a.recorder == nil {
		return
	}
	a.recorder.RecordCacheHit(context.Background(), cacheType, key)
}

func (a *contextlessAdapter) RecordCacheMiss(cacheType, key string) {
	if a == nil || a.recorder == nil {
		return
	}
	a.recorder.RecordCacheMiss(context.Background(), cacheType, key)
}

func (a *contextlessAdapter) RecordCacheOperation(operation, cacheType string, duration float64) {
	if a == nil || a.recorder == nil {
		return
	}
	a.recorder.RecordCacheOperation(context.Background(), operation, cacheType, duration)
}

func (a *contextlessAdapter) RecordKafkaMessageProduced(topic string, duration float64) {
	if a == nil || a.recorder == nil {
		return
	}
	a.recorder.RecordKafkaMessageProduced(context.Background(), topic, duration)
}

func (a *contextlessAdapter) RecordKafkaMessageConsumed(topic string) {
	if a == nil || a.recorder == nil {
		return
	}
	a.recorder.RecordKafkaMessageConsumed(context.Background(), topic)
}

func (a *contextlessAdapter) RecordUserCreation(username, email string) {
	if a == nil || a.recorder == nil {
		return
	}
	a.recorder.RecordUserCreation(context.Background(), username, email)
}

func (a *contextlessAdapter) RecordUserRetrieval(userID string) {
	if a == nil || a.recorder == nil {
		return
	}
	a.recorder.RecordUserRetrieval(context.Background(), userID)
}

func (a *contextlessAdapter) RecordUserActivation(userID string) {
	if a == nil || a.recorder == nil {
		return
	}
	a.recorder.RecordUserActivation(context.Background(), userID)
}

func (a *contextlessAdapter) SetMemoryUsage(bytes float64) {
	if a == nil || a.recorder == nil {
		return
	}
	a.recorder.SetMemoryUsage(context.Background(), bytes)
}

func (a *contextlessAdapter) SetCPUUsage(percent float64) {
	if a == nil || a.recorder == nil {
		return
	}
	a.recorder.SetCPUUsage(context.Background(), percent)
}

func (a *contextlessAdapter) SetGoroutinesCount(count int64) {
	if a == nil || a.recorder == nil {
		return
	}
	a.recorder.SetGoroutinesCount(context.Background(), count)
}

type otelBusinessMetrics struct {
	httpRequestsTotal   metric.Int64Counter
	httpRequestDuration metric.Float64Histogram
	httpErrorsTotal     metric.Int64Counter

	dbOperationsTotal   metric.Int64Counter
	dbOperationDuration metric.Float64Histogram
	dbConnectionsActive metric.Float64Gauge

	cacheHitsTotal         metric.Int64Counter
	cacheMissesTotal       metric.Int64Counter
	cacheOperationsTotal   metric.Int64Counter
	cacheOperationDuration metric.Float64Histogram

	kafkaMessagesProduced metric.Int64Counter
	kafkaMessagesConsumed metric.Int64Counter
	kafkaPublishDuration  metric.Float64Histogram

	userCreationsTotal   metric.Int64Counter
	userRetrievalsTotal  metric.Int64Counter
	userActivationsTotal metric.Int64Counter

	memoryUsage     metric.Float64Gauge
	cpuUsage        metric.Float64Gauge
	goroutinesCount metric.Int64Gauge
}

func newOTELBusinessMetrics(telemetry *Telemetry) (*otelBusinessMetrics, error) {
	meter := telemetry.GetMeter()

	httpRequestsTotal, err := meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	if err != nil {
		return nil, err
	}

	httpRequestDuration, err := meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
	)
	if err != nil {
		return nil, err
	}

	httpErrorsTotal, err := meter.Int64Counter(
		"http_errors_total",
		metric.WithDescription("Total number of HTTP errors"),
	)
	if err != nil {
		return nil, err
	}

	dbOperationsTotal, err := meter.Int64Counter(
		"db_operations_total",
		metric.WithDescription("Total number of database operations"),
	)
	if err != nil {
		return nil, err
	}

	dbOperationDuration, err := meter.Float64Histogram(
		"db_operation_duration_seconds",
		metric.WithDescription("Database operation duration in seconds"),
	)
	if err != nil {
		return nil, err
	}

	dbConnectionsActive, err := meter.Float64Gauge(
		"db_connections_active",
		metric.WithDescription("Number of active database connections"),
	)
	if err != nil {
		return nil, err
	}

	cacheHitsTotal, err := meter.Int64Counter(
		"cache_hits_total",
		metric.WithDescription("Total number of cache hits"),
	)
	if err != nil {
		return nil, err
	}

	cacheMissesTotal, err := meter.Int64Counter(
		"cache_misses_total",
		metric.WithDescription("Total number of cache misses"),
	)
	if err != nil {
		return nil, err
	}

	cacheOperationsTotal, err := meter.Int64Counter(
		"cache_operations_total",
		metric.WithDescription("Total number of cache operations"),
	)
	if err != nil {
		return nil, err
	}

	cacheOperationDuration, err := meter.Float64Histogram(
		"cache_operation_duration_seconds",
		metric.WithDescription("Cache operation duration in seconds"),
	)
	if err != nil {
		return nil, err
	}

	kafkaMessagesProduced, err := meter.Int64Counter(
		"kafka_messages_produced_total",
		metric.WithDescription("Total number of Kafka messages produced"),
	)
	if err != nil {
		return nil, err
	}

	kafkaMessagesConsumed, err := meter.Int64Counter(
		"kafka_messages_consumed_total",
		metric.WithDescription("Total number of Kafka messages consumed"),
	)
	if err != nil {
		return nil, err
	}

	kafkaPublishDuration, err := meter.Float64Histogram(
		"kafka_publish_duration_seconds",
		metric.WithDescription("Kafka publish duration in seconds"),
	)
	if err != nil {
		return nil, err
	}

	userCreationsTotal, err := meter.Int64Counter(
		"user_creations_total",
		metric.WithDescription("Total number of user creations"),
	)
	if err != nil {
		return nil, err
	}

	userRetrievalsTotal, err := meter.Int64Counter(
		"user_retrievals_total",
		metric.WithDescription("Total number of user retrievals"),
	)
	if err != nil {
		return nil, err
	}

	userActivationsTotal, err := meter.Int64Counter(
		"user_activations_total",
		metric.WithDescription("Total number of user activations"),
	)
	if err != nil {
		return nil, err
	}

	memoryUsage, err := meter.Float64Gauge(
		"memory_usage_bytes",
		metric.WithDescription("Memory usage in bytes"),
	)
	if err != nil {
		return nil, err
	}

	cpuUsage, err := meter.Float64Gauge(
		"cpu_usage_percent",
		metric.WithDescription("CPU usage percentage"),
	)
	if err != nil {
		return nil, err
	}

	goroutinesCount, err := meter.Int64Gauge(
		"goroutines_count",
		metric.WithDescription("Number of goroutines"),
	)
	if err != nil {
		return nil, err
	}

	return &otelBusinessMetrics{
		httpRequestsTotal:      httpRequestsTotal,
		httpRequestDuration:    httpRequestDuration,
		httpErrorsTotal:        httpErrorsTotal,
		dbOperationsTotal:      dbOperationsTotal,
		dbOperationDuration:    dbOperationDuration,
		dbConnectionsActive:    dbConnectionsActive,
		cacheHitsTotal:         cacheHitsTotal,
		cacheMissesTotal:       cacheMissesTotal,
		cacheOperationsTotal:   cacheOperationsTotal,
		cacheOperationDuration: cacheOperationDuration,
		kafkaMessagesProduced:  kafkaMessagesProduced,
		kafkaMessagesConsumed:  kafkaMessagesConsumed,
		kafkaPublishDuration:   kafkaPublishDuration,
		userCreationsTotal:     userCreationsTotal,
		userRetrievalsTotal:    userRetrievalsTotal,
		userActivationsTotal:   userActivationsTotal,
		memoryUsage:            memoryUsage,
		cpuUsage:               cpuUsage,
		goroutinesCount:        goroutinesCount,
	}, nil
}

func (bm *otelBusinessMetrics) RecordHTTPRequest(ctx context.Context, method, path, status string, duration float64) {
	bm.httpRequestsTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("method", method),
		attribute.String("path", path),
		attribute.String("status", status),
	))
	bm.httpRequestDuration.Record(ctx, duration, metric.WithAttributes(
		attribute.String("method", method),
		attribute.String("path", path),
	))
}

func (bm *otelBusinessMetrics) RecordHTTPError(ctx context.Context, method, path, status string) {
	bm.httpErrorsTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("method", method),
		attribute.String("path", path),
		attribute.String("status", status),
	))
}

func (bm *otelBusinessMetrics) RecordDBOperation(ctx context.Context, operation, table string, duration float64) {
	bm.dbOperationsTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("operation", operation),
		attribute.String("table", table),
	))
	bm.dbOperationDuration.Record(ctx, duration, metric.WithAttributes(
		attribute.String("operation", operation),
		attribute.String("table", table),
	))
}

func (bm *otelBusinessMetrics) SetDBConnectionsActive(ctx context.Context, count float64) {
	bm.dbConnectionsActive.Record(ctx, count)
}

func (bm *otelBusinessMetrics) RecordCacheHit(ctx context.Context, cacheType, key string) {
	bm.cacheHitsTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("cache_type", cacheType),
		attribute.String("key", key),
	))
	bm.cacheOperationsTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("operation", "hit"),
		attribute.String("cache_type", cacheType),
	))
}

func (bm *otelBusinessMetrics) RecordCacheMiss(ctx context.Context, cacheType, key string) {
	bm.cacheMissesTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("cache_type", cacheType),
		attribute.String("key", key),
	))
	bm.cacheOperationsTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("operation", "miss"),
		attribute.String("cache_type", cacheType),
	))
}

func (bm *otelBusinessMetrics) RecordCacheOperation(ctx context.Context, operation, cacheType string, duration float64) {
	bm.cacheOperationsTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("operation", operation),
		attribute.String("cache_type", cacheType),
	))
	bm.cacheOperationDuration.Record(ctx, duration, metric.WithAttributes(
		attribute.String("operation", operation),
		attribute.String("cache_type", cacheType),
	))
}

func (bm *otelBusinessMetrics) RecordKafkaMessageProduced(ctx context.Context, topic string, duration float64) {
	bm.kafkaMessagesProduced.Add(ctx, 1, metric.WithAttributes(
		attribute.String("topic", topic),
	))
	bm.kafkaPublishDuration.Record(ctx, duration, metric.WithAttributes(
		attribute.String("topic", topic),
	))
}

func (bm *otelBusinessMetrics) RecordKafkaMessageConsumed(ctx context.Context, topic string) {
	bm.kafkaMessagesConsumed.Add(ctx, 1, metric.WithAttributes(
		attribute.String("topic", topic),
	))
}

func (bm *otelBusinessMetrics) RecordUserCreation(ctx context.Context, username, email string) {
	bm.userCreationsTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("username", username),
		attribute.String("email", email),
	))
}

func (bm *otelBusinessMetrics) RecordUserRetrieval(ctx context.Context, userID string) {
	bm.userRetrievalsTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("user_id", userID),
	))
}

func (bm *otelBusinessMetrics) RecordUserActivation(ctx context.Context, userID string) {
	bm.userActivationsTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("user_id", userID),
	))
}

func (bm *otelBusinessMetrics) SetMemoryUsage(ctx context.Context, bytes float64) {
	bm.memoryUsage.Record(ctx, bytes)
}

func (bm *otelBusinessMetrics) SetCPUUsage(ctx context.Context, percent float64) {
	bm.cpuUsage.Record(ctx, percent)
}

func (bm *otelBusinessMetrics) SetGoroutinesCount(ctx context.Context, count int64) {
	bm.goroutinesCount.Record(ctx, count)
}
