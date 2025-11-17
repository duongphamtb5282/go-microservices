package health

import (
	"time"
)

// QueryStatsImpl represents query statistics implementation
type QueryStatsImpl struct {
	totalQueries      int64
	successfulQueries int64
	failedQueries     int64
	averageDuration   time.Duration
	slowestQuery      time.Duration
	fastestQuery      time.Duration
}

// NewQueryStats creates a new query stats instance
func NewQueryStats(totalQueries, successfulQueries, failedQueries int64, averageDuration, slowestQuery, fastestQuery time.Duration) *QueryStatsImpl {
	return &QueryStatsImpl{
		totalQueries:      totalQueries,
		successfulQueries: successfulQueries,
		failedQueries:     failedQueries,
		averageDuration:   averageDuration,
		slowestQuery:      slowestQuery,
		fastestQuery:      fastestQuery,
	}
}

// GetTotalQueries returns the total number of queries
func (q *QueryStatsImpl) GetTotalQueries() int64 {
	return q.totalQueries
}

// GetSuccessfulQueries returns the number of successful queries
func (q *QueryStatsImpl) GetSuccessfulQueries() int64 {
	return q.successfulQueries
}

// GetFailedQueries returns the number of failed queries
func (q *QueryStatsImpl) GetFailedQueries() int64 {
	return q.failedQueries
}

// GetAverageDuration returns the average query duration
func (q *QueryStatsImpl) GetAverageDuration() time.Duration {
	return q.averageDuration
}

// GetSlowestQuery returns the slowest query duration
func (q *QueryStatsImpl) GetSlowestQuery() time.Duration {
	return q.slowestQuery
}

// GetFastestQuery returns the fastest query duration
func (q *QueryStatsImpl) GetFastestQuery() time.Duration {
	return q.fastestQuery
}

// GetSuccessRate returns the success rate percentage
func (q *QueryStatsImpl) GetSuccessRate() float64 {
	if q.totalQueries == 0 {
		return 0.0
	}
	return float64(q.successfulQueries) / float64(q.totalQueries) * 100.0
}

// GetFailureRate returns the failure rate percentage
func (q *QueryStatsImpl) GetFailureRate() float64 {
	if q.totalQueries == 0 {
		return 0.0
	}
	return float64(q.failedQueries) / float64(q.totalQueries) * 100.0
}

// ConnectionStatsImpl represents connection statistics implementation
type ConnectionStatsImpl struct {
	totalConnections   int64
	connectionFailures int64
}

// NewConnectionStats creates a new connection stats instance
func NewConnectionStats(totalConnections, connectionFailures int64) *ConnectionStatsImpl {
	return &ConnectionStatsImpl{
		totalConnections:   totalConnections,
		connectionFailures: connectionFailures,
	}
}

// GetTotalConnections returns the total number of connections
func (c *ConnectionStatsImpl) GetTotalConnections() int64 {
	return c.totalConnections
}

// GetConnectionFailures returns the number of connection failures
func (c *ConnectionStatsImpl) GetConnectionFailures() int64 {
	return c.connectionFailures
}

// GetConnectionFailureRate returns the connection failure rate percentage
func (c *ConnectionStatsImpl) GetConnectionFailureRate() float64 {
	if c.totalConnections == 0 {
		return 0.0
	}
	return float64(c.connectionFailures) / float64(c.totalConnections) * 100.0
}

// TransactionStatsImpl represents transaction statistics implementation
type TransactionStatsImpl struct {
	totalTransactions      int64
	successfulTransactions int64
	failedTransactions     int64
	rolledBackTransactions int64
	averageDuration        time.Duration
}

// NewTransactionStats creates a new transaction stats instance
func NewTransactionStats(totalTransactions, successfulTransactions, failedTransactions, rolledBackTransactions int64, averageDuration time.Duration) *TransactionStatsImpl {
	return &TransactionStatsImpl{
		totalTransactions:      totalTransactions,
		successfulTransactions: successfulTransactions,
		failedTransactions:     failedTransactions,
		rolledBackTransactions: rolledBackTransactions,
		averageDuration:        averageDuration,
	}
}

// GetTotalTransactions returns the total number of transactions
func (t *TransactionStatsImpl) GetTotalTransactions() int64 {
	return t.totalTransactions
}

// GetSuccessfulTransactions returns the number of successful transactions
func (t *TransactionStatsImpl) GetSuccessfulTransactions() int64 {
	return t.successfulTransactions
}

// GetFailedTransactions returns the number of failed transactions
func (t *TransactionStatsImpl) GetFailedTransactions() int64 {
	return t.failedTransactions
}

// GetRolledBackTransactions returns the number of rolled back transactions
func (t *TransactionStatsImpl) GetRolledBackTransactions() int64 {
	return t.rolledBackTransactions
}

// GetAverageDuration returns the average transaction duration
func (t *TransactionStatsImpl) GetAverageDuration() time.Duration {
	return t.averageDuration
}

// GetTransactionSuccessRate returns the transaction success rate percentage
func (t *TransactionStatsImpl) GetTransactionSuccessRate() float64 {
	if t.totalTransactions == 0 {
		return 0.0
	}
	return float64(t.successfulTransactions) / float64(t.totalTransactions) * 100.0
}

// GetTransactionFailureRate returns the transaction failure rate percentage
func (t *TransactionStatsImpl) GetTransactionFailureRate() float64 {
	if t.totalTransactions == 0 {
		return 0.0
	}
	return float64(t.failedTransactions) / float64(t.totalTransactions) * 100.0
}

// GetRollbackRate returns the rollback rate percentage
func (t *TransactionStatsImpl) GetRollbackRate() float64 {
	if t.totalTransactions == 0 {
		return 0.0
	}
	return float64(t.rolledBackTransactions) / float64(t.totalTransactions) * 100.0
}

// OverallStatsImpl represents overall database statistics implementation
type OverallStatsImpl struct {
	queryStats       QueryStats
	connectionStats  ConnectionStats
	transactionStats TransactionStats
	uptime           time.Duration
	lastActivity     time.Time
}

// NewOverallStats creates a new overall stats instance
func NewOverallStats(queryStats QueryStats, connectionStats ConnectionStats, transactionStats TransactionStats, uptime time.Duration, lastActivity time.Time) *OverallStatsImpl {
	return &OverallStatsImpl{
		queryStats:       queryStats,
		connectionStats:  connectionStats,
		transactionStats: transactionStats,
		uptime:           uptime,
		lastActivity:     lastActivity,
	}
}

// GetQueryStats returns the query statistics
func (o *OverallStatsImpl) GetQueryStats() QueryStats {
	return o.queryStats
}

// GetConnectionStats returns the connection statistics
func (o *OverallStatsImpl) GetConnectionStats() ConnectionStats {
	return o.connectionStats
}

// GetTransactionStats returns the transaction statistics
func (o *OverallStatsImpl) GetTransactionStats() TransactionStats {
	return o.transactionStats
}

// GetUptime returns the uptime
func (o *OverallStatsImpl) GetUptime() time.Duration {
	return o.uptime
}

// GetLastActivity returns the last activity time
func (o *OverallStatsImpl) GetLastActivity() time.Time {
	return o.lastActivity
}

// GetOverallHealthScore returns an overall health score (0-100)
func (o *OverallStatsImpl) GetOverallHealthScore() float64 {
	// Calculate success rates from basic stats
	querySuccessRate := float64(0)
	if o.queryStats.GetTotalQueries() > 0 {
		querySuccessRate = float64(o.queryStats.GetSuccessfulQueries()) / float64(o.queryStats.GetTotalQueries()) * 100.0
	}

	connectionFailureRate := float64(0)
	if o.connectionStats.GetTotalConnections() > 0 {
		connectionFailureRate = float64(o.connectionStats.GetConnectionFailures()) / float64(o.connectionStats.GetTotalConnections()) * 100.0
	}

	transactionSuccessRate := float64(0)
	if o.transactionStats.GetTotalTransactions() > 0 {
		transactionSuccessRate = float64(o.transactionStats.GetSuccessfulTransactions()) / float64(o.transactionStats.GetTotalTransactions()) * 100.0
	}

	// Calculate weighted average
	score := (querySuccessRate * 0.4) + ((100 - connectionFailureRate) * 0.3) + (transactionSuccessRate * 0.3)

	// Ensure score is between 0 and 100
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}

	return score
}

// IsHealthy returns true if the overall stats indicate a healthy state
func (o *OverallStatsImpl) IsHealthy() bool {
	score := o.GetOverallHealthScore()
	return score >= 80.0 // Consider healthy if score is 80% or above
}
