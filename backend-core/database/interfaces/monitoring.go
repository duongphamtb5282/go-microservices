package interfaces

import (
	"context"
	"time"

	"backend-core/database/health"
)

// DatabaseMonitor defines the monitoring interface
type DatabaseMonitor interface {
	// Monitoring control
	StartMonitoring(ctx context.Context) error
	StopMonitoring(ctx context.Context) error
	IsMonitoring() bool

	// Statistics
	GetQueryStats() health.QueryStats
	GetConnectionStats() health.ConnectionStats
	GetTransactionStats() health.TransactionStats
	GetOverallStats() health.StatsResult

	// Recording
	RecordQuery(query string, duration time.Duration, err error)
	RecordConnection(connected bool, err error)
	RecordTransaction(operation string, duration time.Duration, err error)
}
