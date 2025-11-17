package telemetry

import (
	"context"
	"fmt"

	"backend-core/telemetry"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsService integrates backend-core telemetry with Prometheus
type MetricsService struct {
	backendTelemetry  *telemetry.Telemetry
	businessMetrics   *telemetry.BusinessMetrics
	prometheusMetrics *prometheusBusinessMetrics
	registry          *prometheus.Registry
}

// NewMetricsService creates a new metrics service
func NewMetricsService(backendTelemetry *telemetry.Telemetry) (*MetricsService, error) {
	// Create a custom registry for our metrics
	registry := prometheus.NewRegistry()

	// Register standard collectors
	registry.MustRegister(
		prometheus.NewGoCollector(),
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
	)

	// Create Prometheus business metrics
	prometheusMetrics, err := createPrometheusBusinessMetrics(registry)
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus business metrics: %w", err)
	}

	// Try to create backend-core business metrics (may fail if telemetry disabled)
	var businessMetrics *telemetry.BusinessMetrics
	if backendTelemetry != nil && backendTelemetry.Config.Enabled {
		businessMetrics, err = telemetry.NewBusinessMetrics(backendTelemetry)
		if err != nil {
			// Log warning but continue
			fmt.Printf("Warning: Failed to create backend-core business metrics: %v\n", err)
		}
	}

	service := &MetricsService{
		backendTelemetry:  backendTelemetry,
		businessMetrics:   businessMetrics,
		registry:          registry,
		prometheusMetrics: prometheusMetrics,
	}

	return service, nil
}

// GetBusinessMetrics returns the business metrics instance
func (ms *MetricsService) GetBusinessMetrics() *telemetry.BusinessMetrics {
	return ms.businessMetrics
}

// GetBackendTelemetry returns the backend telemetry instance
func (ms *MetricsService) GetBackendTelemetry() *telemetry.Telemetry {
	return ms.backendTelemetry
}

// MetricsHandler returns the Prometheus metrics HTTP handler
func (ms *MetricsService) MetricsHandler() gin.HandlerFunc {
	h := promhttp.HandlerFor(ms.registry, promhttp.HandlerOpts{
		Registry: ms.registry,
	})

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// createPrometheusBusinessMetrics creates business metrics using Prometheus directly
func createPrometheusBusinessMetrics(registry *prometheus.Registry) (*telemetry.BusinessMetrics, error) {
	// Create Prometheus metrics directly

	// HTTP Metrics
	httpRequestsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	httpRequestDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	httpErrorsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_errors_total",
		Help: "Total number of HTTP errors",
	}, []string{"method", "path", "status"})

	// Database Metrics
	dbOperationsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "db_operations_total",
		Help: "Total number of database operations",
	}, []string{"operation", "table"})

	dbOperationDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "db_operation_duration_seconds",
		Help:    "Database operation duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation", "table"})

	dbConnectionsActive := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "db_connections_active",
		Help: "Number of active database connections",
	})

	// Cache Metrics
	cacheHitsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cache_hits_total",
		Help: "Total number of cache hits",
	}, []string{"cache_type", "key"})

	cacheMissesTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cache_misses_total",
		Help: "Total number of cache misses",
	}, []string{"cache_type", "key"})

	cacheOperationsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cache_operations_total",
		Help: "Total number of cache operations",
	}, []string{"operation", "cache_type"})

	cacheOperationDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "cache_operation_duration_seconds",
		Help:    "Cache operation duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation", "cache_type"})

	// Kafka Metrics
	kafkaMessagesProduced := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kafka_messages_produced_total",
		Help: "Total number of Kafka messages produced",
	}, []string{"topic"})

	kafkaMessagesConsumed := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kafka_messages_consumed_total",
		Help: "Total number of Kafka messages consumed",
	}, []string{"topic"})

	kafkaPublishDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kafka_publish_duration_seconds",
		Help:    "Kafka publish duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"topic"})

	// Business Metrics
	userCreationsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "user_creations_total",
		Help: "Total number of user creations",
	}, []string{"username", "email"})

	userRetrievalsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "user_retrievals_total",
		Help: "Total number of user retrievals",
	}, []string{"user_id"})

	userActivationsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "user_activations_total",
		Help: "Total number of user activations",
	}, []string{"user_id"})

	// System Metrics
	memoryUsage := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "memory_usage_bytes",
		Help: "Memory usage in bytes",
	})

	cpuUsage := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cpu_usage_percent",
		Help: "CPU usage percentage",
	})

	goroutinesCount := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "goroutines_count",
		Help: "Number of goroutines",
	})

	// Register all metrics with the registry
	registry.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		httpErrorsTotal,
		dbOperationsTotal,
		dbOperationDuration,
		dbConnectionsActive,
		cacheHitsTotal,
		cacheMissesTotal,
		cacheOperationsTotal,
		cacheOperationDuration,
		kafkaMessagesProduced,
		kafkaMessagesConsumed,
		kafkaPublishDuration,
		userCreationsTotal,
		userRetrievalsTotal,
		userActivationsTotal,
		memoryUsage,
		cpuUsage,
		goroutinesCount,
	)

	// Create a wrapper that implements the BusinessMetrics interface
	businessMetrics := &prometheusBusinessMetrics{
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
	}

	return &telemetry.BusinessMetrics{
		HTTPRequestCounter:    businessMetrics,
		HTTPErrorCounter:      businessMetrics,
		DBOperationCounter:    businessMetrics,
		CacheHitCounter:       businessMetrics,
		CacheMissCounter:      businessMetrics,
		CacheOperationCounter: businessMetrics,
		KafkaMessageCounter:   businessMetrics,
		UserCreationCounter:   businessMetrics,
		UserRetrievalCounter:  businessMetrics,
		UserActivationCounter: businessMetrics,
		MemoryGauge:           businessMetrics,
		CPUGauge:              businessMetrics,
		GoroutinesGauge:       businessMetrics,
	}, nil
}

// prometheusBusinessMetrics implements the business metrics interface using Prometheus
type prometheusBusinessMetrics struct {
	httpRequestsTotal      *prometheus.CounterVec
	httpRequestDuration    *prometheus.HistogramVec
	httpErrorsTotal        *prometheus.CounterVec
	dbOperationsTotal      *prometheus.CounterVec
	dbOperationDuration    *prometheus.HistogramVec
	dbConnectionsActive    prometheus.Gauge
	cacheHitsTotal         *prometheus.CounterVec
	cacheMissesTotal       *prometheus.CounterVec
	cacheOperationsTotal   *prometheus.CounterVec
	cacheOperationDuration *prometheus.HistogramVec
	kafkaMessagesProduced  *prometheus.CounterVec
	kafkaMessagesConsumed  *prometheus.CounterVec
	kafkaPublishDuration   *prometheus.HistogramVec
	userCreationsTotal     *prometheus.CounterVec
	userRetrievalsTotal    *prometheus.CounterVec
	userActivationsTotal   *prometheus.CounterVec
	memoryUsage            prometheus.Gauge
	cpuUsage               prometheus.Gauge
	goroutinesCount        prometheus.Gauge
}

// Implement the BusinessMetrics interface methods
func (p *prometheusBusinessMetrics) RecordHTTPRequest(method, path, status string, duration float64) {
	p.httpRequestsTotal.WithLabelValues(method, path, status).Inc()
	p.httpRequestDuration.WithLabelValues(method, path).Observe(duration)
}

func (p *prometheusBusinessMetrics) RecordHTTPError(method, path, status string) {
	p.httpErrorsTotal.WithLabelValues(method, path, status).Inc()
}

func (p *prometheusBusinessMetrics) RecordDBOperation(operation, table string, duration float64) {
	p.dbOperationsTotal.WithLabelValues(operation, table).Inc()
	p.dbOperationDuration.WithLabelValues(operation, table).Observe(duration)
}

func (p *prometheusBusinessMetrics) SetDBConnectionsActive(count float64) {
	p.dbConnectionsActive.Set(count)
}

func (p *prometheusBusinessMetrics) RecordCacheHit(cacheType, key string) {
	p.cacheHitsTotal.WithLabelValues(cacheType, key).Inc()
}

func (p *prometheusBusinessMetrics) RecordCacheMiss(cacheType, key string) {
	p.cacheMissesTotal.WithLabelValues(cacheType, key).Inc()
}

func (p *prometheusBusinessMetrics) RecordCacheOperation(operation, cacheType string, duration float64) {
	p.cacheOperationsTotal.WithLabelValues(operation, cacheType).Inc()
	p.cacheOperationDuration.WithLabelValues(operation, cacheType).Observe(duration)
}

func (p *prometheusBusinessMetrics) RecordKafkaMessageProduced(topic string, duration float64) {
	p.kafkaMessagesProduced.WithLabelValues(topic).Inc()
	p.kafkaPublishDuration.WithLabelValues(topic).Observe(duration)
}

func (p *prometheusBusinessMetrics) RecordKafkaMessageConsumed(topic string) {
	p.kafkaMessagesConsumed.WithLabelValues(topic).Inc()
}

func (p *prometheusBusinessMetrics) RecordUserCreation(username, email string) {
	p.userCreationsTotal.WithLabelValues(username, email).Inc()
}

func (p *prometheusBusinessMetrics) RecordUserRetrieval(userID string) {
	p.userRetrievalsTotal.WithLabelValues(userID).Inc()
}

func (p *prometheusBusinessMetrics) RecordUserActivation(userID string) {
	p.userActivationsTotal.WithLabelValues(userID).Inc()
}

func (p *prometheusBusinessMetrics) SetMemoryUsage(bytes float64) {
	p.memoryUsage.Set(bytes)
}

func (p *prometheusBusinessMetrics) SetCPUUsage(percent float64) {
	p.cpuUsage.Set(percent)
}

func (p *prometheusBusinessMetrics) SetGoroutinesCount(count int64) {
	p.goroutinesCount.Set(float64(count))
}

// RecordHTTPMetrics records HTTP request metrics
func (ms *MetricsService) RecordHTTPMetrics(method, path, status string, duration float64, statusCode int) {
	if ms.prometheusMetrics != nil {
		ms.prometheusMetrics.RecordHTTPRequest(method, path, status, duration)
		if statusCode >= 400 {
			ms.prometheusMetrics.RecordHTTPError(method, path, status)
		}
	}
}

// Shutdown shuts down the metrics service
func (ms *MetricsService) Shutdown(ctx context.Context) error {
	if ms.backendTelemetry != nil {
		return ms.backendTelemetry.Shutdown(ctx)
	}
	return nil
}
