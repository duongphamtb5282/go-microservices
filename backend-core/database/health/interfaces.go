package health

import (
	"context"
	"time"
)

// HealthStatus represents the health status of a database
type HealthStatus int

const (
	HealthStatusUnknown HealthStatus = iota
	HealthStatusHealthy
	HealthStatusDegraded
	HealthStatusUnhealthy
)

// String returns the string representation of HealthStatus
func (s HealthStatus) String() string {
	switch s {
	case HealthStatusUnknown:
		return "unknown"
	case HealthStatusHealthy:
		return "healthy"
	case HealthStatusDegraded:
		return "degraded"
	case HealthStatusUnhealthy:
		return "unhealthy"
	default:
		return "unknown"
	}
}

// HealthChecker defines the interface for health checking
type HealthChecker interface {
	CheckHealth(ctx context.Context) HealthResult
	GetName() string
}

// MetricsCollector defines the interface for collecting metrics
type MetricsCollector interface {
	CollectMetrics(ctx context.Context) MetricsResult
	GetName() string
}

// StatsCollector defines the interface for collecting statistics
type StatsCollector interface {
	CollectStats(ctx context.Context) StatsResult
	GetName() string
}

// MigrationChecker defines the interface for checking migrations
type MigrationChecker interface {
	CheckMigrations(ctx context.Context) MigrationResult
	GetName() string
}

// HealthResult represents the result of a health check
type HealthResult interface {
	GetStatus() HealthStatus
	GetResponseTime() time.Duration
	GetError() error
	GetDetails() map[string]interface{}
	GetTimestamp() time.Time
}

// MetricsResult represents the result of metrics collection
type MetricsResult interface {
	GetActiveConnections() int
	GetIdleConnections() int
	GetMaxConnections() int
	GetResponseTime() time.Duration
	GetLastQueryTime() time.Time
	GetQueryCount() int64
	GetErrorCount() int64
}

// StatsResult represents the result of statistics collection
type StatsResult interface {
	GetQueryStats() QueryStats
	GetConnectionStats() ConnectionStats
	GetTransactionStats() TransactionStats
	GetUptime() time.Duration
	GetLastActivity() time.Time
}

// QueryStats represents query statistics
type QueryStats interface {
	GetTotalQueries() int64
	GetSuccessfulQueries() int64
	GetFailedQueries() int64
	GetAverageDuration() time.Duration
	GetSlowestQuery() time.Duration
	GetFastestQuery() time.Duration
}

// ConnectionStats represents connection statistics
type ConnectionStats interface {
	GetTotalConnections() int64
	GetConnectionFailures() int64
}

// TransactionStats represents transaction statistics
type TransactionStats interface {
	GetTotalTransactions() int64
	GetSuccessfulTransactions() int64
	GetFailedTransactions() int64
	GetRolledBackTransactions() int64
	GetAverageDuration() time.Duration
}

// MigrationResult represents the result of a migration check
type MigrationResult interface {
	GetSuccess() bool
	GetError() error
	GetMigrations() []Migration
	GetDuration() time.Duration
}

// Migration represents a database migration
type Migration interface {
	GetID() string
	GetVersion() string
	GetDescription() string
	GetAppliedAt() time.Time
	GetChecksum() string
}

// HealthReporter defines the interface for reporting health status
type HealthReporter interface {
	ReportHealth(ctx context.Context, result HealthResult) error
	GetName() string
}

// MetricsReporter defines the interface for reporting metrics
type MetricsReporter interface {
	ReportMetrics(ctx context.Context, result MetricsResult) error
	GetName() string
}

// StatsReporter defines the interface for reporting statistics
type StatsReporter interface {
	ReportStats(ctx context.Context, result StatsResult) error
	GetName() string
}
