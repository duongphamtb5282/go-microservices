package health

import (
	"fmt"
	"time"
)

// PerformanceMetrics represents database performance metrics
type PerformanceMetrics struct {
	activeConnections int
	idleConnections   int
	maxConnections    int
	responseTime      time.Duration
	lastQueryTime     time.Time
	queryCount        int64
	errorCount        int64
}

// NewPerformanceMetrics creates a new performance metrics instance
func NewPerformanceMetrics(activeConnections, idleConnections, maxConnections int, responseTime time.Duration, lastQueryTime time.Time, queryCount, errorCount int64) *PerformanceMetrics {
	return &PerformanceMetrics{
		activeConnections: activeConnections,
		idleConnections:   idleConnections,
		maxConnections:    maxConnections,
		responseTime:      responseTime,
		lastQueryTime:     lastQueryTime,
		queryCount:        queryCount,
		errorCount:        errorCount,
	}
}

// GetActiveConnections returns the number of active connections
func (p *PerformanceMetrics) GetActiveConnections() int {
	return p.activeConnections
}

// GetIdleConnections returns the number of idle connections
func (p *PerformanceMetrics) GetIdleConnections() int {
	return p.idleConnections
}

// GetMaxConnections returns the maximum number of connections
func (p *PerformanceMetrics) GetMaxConnections() int {
	return p.maxConnections
}

// GetResponseTime returns the response time
func (p *PerformanceMetrics) GetResponseTime() time.Duration {
	return p.responseTime
}

// GetLastQueryTime returns the last query time
func (p *PerformanceMetrics) GetLastQueryTime() time.Time {
	return p.lastQueryTime
}

// GetQueryCount returns the query count
func (p *PerformanceMetrics) GetQueryCount() int64 {
	return p.queryCount
}

// GetErrorCount returns the error count
func (p *PerformanceMetrics) GetErrorCount() int64 {
	return p.errorCount
}

// SetActiveConnections sets the number of active connections
func (p *PerformanceMetrics) SetActiveConnections(count int) {
	p.activeConnections = count
}

// SetIdleConnections sets the number of idle connections
func (p *PerformanceMetrics) SetIdleConnections(count int) {
	p.idleConnections = count
}

// SetMaxConnections sets the maximum number of connections
func (p *PerformanceMetrics) SetMaxConnections(count int) {
	p.maxConnections = count
}

// SetResponseTime sets the response time
func (p *PerformanceMetrics) SetResponseTime(duration time.Duration) {
	p.responseTime = duration
}

// SetLastQueryTime sets the last query time
func (p *PerformanceMetrics) SetLastQueryTime(t time.Time) {
	p.lastQueryTime = t
}

// SetQueryCount sets the query count
func (p *PerformanceMetrics) SetQueryCount(count int64) {
	p.queryCount = count
}

// SetErrorCount sets the error count
func (p *PerformanceMetrics) SetErrorCount(count int64) {
	p.errorCount = count
}

// GetConnectionUtilization returns the connection utilization percentage
func (p *PerformanceMetrics) GetConnectionUtilization() float64 {
	if p.maxConnections == 0 {
		return 0.0
	}
	return float64(p.activeConnections) / float64(p.maxConnections) * 100.0
}

// GetErrorRate returns the error rate percentage
func (p *PerformanceMetrics) GetErrorRate() float64 {
	if p.queryCount == 0 {
		return 0.0
	}
	return float64(p.errorCount) / float64(p.queryCount) * 100.0
}

// IsHealthy returns true if the metrics indicate a healthy state
func (p *PerformanceMetrics) IsHealthy() bool {
	utilization := p.GetConnectionUtilization()
	errorRate := p.GetErrorRate()

	// Consider healthy if utilization is below 80% and error rate is below 5%
	return utilization < 80.0 && errorRate < 5.0
}

// String returns a string representation of the performance metrics
func (p *PerformanceMetrics) String() string {
	return fmt.Sprintf("PerformanceMetrics{active=%d, idle=%d, max=%d, responseTime=%v, queries=%d, errors=%d}",
		p.activeConnections, p.idleConnections, p.maxConnections, p.responseTime, p.queryCount, p.errorCount)
}
