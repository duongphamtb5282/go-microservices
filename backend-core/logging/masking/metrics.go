package masking

import (
	"sync"
	"time"
)

// InMemoryMetricsCollector implements in-memory metrics collection
type InMemoryMetricsCollector struct {
	metrics MaskingMetrics
	mutex   sync.RWMutex
}

// NewInMemoryMetricsCollector creates a new in-memory metrics collector
func NewInMemoryMetricsCollector() *InMemoryMetricsCollector {
	return &InMemoryMetricsCollector{
		metrics: MaskingMetrics{},
	}
}

// RecordMasking records a masking operation
func (m *InMemoryMetricsCollector) RecordMasking(field string, duration time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.metrics.FieldsMasked++
	m.metrics.MaskingTime += duration
}

// RecordError records a masking error
func (m *InMemoryMetricsCollector) RecordError(field string, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.metrics.Errors++
}

// RecordCacheHit records a cache hit
func (m *InMemoryMetricsCollector) RecordCacheHit() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.metrics.CacheHits++
}

// RecordCacheMiss records a cache miss
func (m *InMemoryMetricsCollector) RecordCacheMiss() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.metrics.CacheMisses++
}

// GetMetrics returns current metrics
func (m *InMemoryMetricsCollector) GetMetrics() MaskingMetrics {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Return a copy to avoid race conditions
	return MaskingMetrics{
		FieldsMasked: m.metrics.FieldsMasked,
		MaskingTime:  m.metrics.MaskingTime,
		RulesMatched: m.metrics.RulesMatched,
		Errors:       m.metrics.Errors,
		CacheHits:    m.metrics.CacheHits,
		CacheMisses:  m.metrics.CacheMisses,
	}
}

// ResetMetrics resets metrics
func (m *InMemoryMetricsCollector) ResetMetrics() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.metrics = MaskingMetrics{}
}

// NoOpMetricsCollector implements a no-operation metrics collector
type NoOpMetricsCollector struct{}

// NewNoOpMetricsCollector creates a new no-operation metrics collector
func NewNoOpMetricsCollector() *NoOpMetricsCollector {
	return &NoOpMetricsCollector{}
}

// RecordMasking does nothing
func (m *NoOpMetricsCollector) RecordMasking(field string, duration time.Duration) {
	// No-op
}

// RecordError does nothing
func (m *NoOpMetricsCollector) RecordError(field string, err error) {
	// No-op
}

// RecordCacheHit does nothing
func (m *NoOpMetricsCollector) RecordCacheHit() {
	// No-op
}

// RecordCacheMiss does nothing
func (m *NoOpMetricsCollector) RecordCacheMiss() {
	// No-op
}

// GetMetrics returns zero metrics
func (m *NoOpMetricsCollector) GetMetrics() MaskingMetrics {
	return MaskingMetrics{}
}

// ResetMetrics does nothing
func (m *NoOpMetricsCollector) ResetMetrics() {
	// No-op
}
