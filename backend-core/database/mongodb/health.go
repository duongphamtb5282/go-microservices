package mongodb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"backend-core/config"
	"backend-core/database"
	"backend-core/database/health"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoDBHealthChecker implements health checking for MongoDB
type MongoDBHealthChecker struct {
	client        *mongo.Client
	config        *config.DatabaseConfig
	status        database.HealthStatus
	lastCheck     time.Time
	healthHistory []health.HealthResult
	mu            sync.RWMutex
}

// NewMongoDBHealthChecker creates a new MongoDB health checker
func NewMongoDBHealthChecker(client *mongo.Client, config *config.DatabaseConfig) *MongoDBHealthChecker {
	return &MongoDBHealthChecker{
		client:        client,
		config:        config,
		status:        database.HealthStatusUnknown,
		healthHistory: make([]health.HealthResult, 0),
	}
}

// CheckHealth performs a comprehensive health check
func (h *MongoDBHealthChecker) CheckHealth(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	start := time.Now()
	var check *health.HealthCheck

	// Check connection
	if err := h.CheckConnection(ctx); err != nil {
		check = health.NewHealthCheck(health.HealthStatusUnhealthy, time.Since(start), err, make(map[string]interface{}))
		h.status = health.HealthStatusUnhealthy
	} else {
		// Check performance
		metrics, err := h.CheckPerformance(ctx)
		if err != nil {
			check = health.NewHealthCheck(health.HealthStatusDegraded, time.Since(start), err, make(map[string]interface{}))
			h.status = health.HealthStatusDegraded
		} else {
			details := make(map[string]interface{})
			details["metrics"] = metrics
			check = health.NewHealthCheck(health.HealthStatusHealthy, time.Since(start), nil, details)
			h.status = health.HealthStatusHealthy
		}
	}
	h.lastCheck = start
	h.healthHistory = append(h.healthHistory, check)

	// Keep only last 100 health checks
	if len(h.healthHistory) > 100 {
		h.healthHistory = h.healthHistory[len(h.healthHistory)-100:]
	}

	return check.GetError()
}

// CheckConnection checks the database connection
func (h *MongoDBHealthChecker) CheckConnection(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := h.client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	return nil
}

// CheckPerformance checks database performance metrics
func (h *MongoDBHealthChecker) CheckPerformance(ctx context.Context) (health.MetricsResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Test query performance
	start := time.Now()
	var result bson.M
	err := h.client.Database("admin").RunCommand(ctx, bson.M{"ping": 1}).Decode(&result)
	responseTime := time.Since(start)
	lastQueryTime := time.Now()

	if err != nil {
		return nil, fmt.Errorf("performance check query failed: %w", err)
	}

	// Get MongoDB-specific metrics
	activeConnections, idleConnections, maxConnections, queryCount, errorCount, err := h.getMongoDBMetrics(ctx)
	if err != nil {
		// Log error but don't fail the health check
		fmt.Printf("Warning: failed to get MongoDB metrics: %v\n", err)
		// Use default values
		activeConnections, idleConnections, maxConnections, queryCount, errorCount = 0, 0, 0, 0, 0
	}

	metrics := health.NewPerformanceMetrics(activeConnections, idleConnections, maxConnections, responseTime, lastQueryTime, queryCount, errorCount)
	return metrics, nil
}

// GetStatus returns the current health status
func (h *MongoDBHealthChecker) GetStatus() database.HealthStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.status
}

// GetLastHealthCheck returns the timestamp of the last health check
func (h *MongoDBHealthChecker) GetLastHealthCheck() time.Time {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.lastCheck
}

// GetHealthHistory returns the health check history
func (h *MongoDBHealthChecker) GetHealthHistory() []health.HealthResult {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.healthHistory
}

// getMongoDBMetrics gets MongoDB-specific metrics
func (h *MongoDBHealthChecker) getMongoDBMetrics(ctx context.Context) (activeConnections, idleConnections, maxConnections int, queryCount, errorCount int64, err error) {
	// Get server status
	var serverStatus bson.M
	err = h.client.Database("admin").RunCommand(ctx, bson.M{"serverStatus": 1}).Decode(&serverStatus)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	// Extract connection metrics
	if connections, ok := serverStatus["connections"].(bson.M); ok {
		if current, ok := connections["current"].(int32); ok {
			activeConnections = int(current)
		}
		if available, ok := connections["available"].(int32); ok {
			idleConnections = int(available)
		}
	}

	// Extract operation metrics
	if opcounters, ok := serverStatus["opcounters"].(bson.M); ok {
		if queryCountVal, ok := opcounters["query"].(int32); ok {
			queryCount = int64(queryCountVal)
		}
	}

	return activeConnections, idleConnections, maxConnections, queryCount, errorCount, nil
}

// MongoDBMonitor implements monitoring for MongoDB
type MongoDBMonitor struct {
	monitoring   bool
	queryStats   database.QueryStats
	connStats    database.ConnectionStats
	transStats   database.TransactionStats
	overallStats database.OverallStats
	startTime    time.Time
	lastActivity time.Time
	mu           sync.RWMutex
}

// NewMongoDBMonitor creates a new MongoDB monitor
func NewMongoDBMonitor() *MongoDBMonitor {
	return &MongoDBMonitor{
		queryStats:   health.NewQueryStats(0, 0, 0, 0, 0, 0),
		connStats:    health.NewConnectionStats(0, 0),
		transStats:   health.NewTransactionStats(0, 0, 0, 0, 0),
		overallStats: health.NewOverallStats(health.NewQueryStats(0, 0, 0, 0, 0, 0), health.NewConnectionStats(0, 0), health.NewTransactionStats(0, 0, 0, 0, 0), 0, time.Now()),
	}
}

// StartMonitoring starts monitoring database operations
func (m *MongoDBMonitor) StartMonitoring(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.monitoring {
		return fmt.Errorf("monitoring is already started")
	}

	m.monitoring = true
	m.startTime = time.Now()
	m.lastActivity = time.Now()

	return nil
}

// StopMonitoring stops monitoring database operations
func (m *MongoDBMonitor) StopMonitoring(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.monitoring {
		return fmt.Errorf("monitoring is not started")
	}

	m.monitoring = false
	return nil
}

// IsMonitoring returns whether monitoring is active
func (m *MongoDBMonitor) IsMonitoring() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.monitoring
}

// RecordQuery records a query execution
func (m *MongoDBMonitor) RecordQuery(query string, duration time.Duration, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get current stats
	totalQueries := m.queryStats.GetTotalQueries()
	successfulQueries := m.queryStats.GetSuccessfulQueries()
	failedQueries := m.queryStats.GetFailedQueries()
	averageDuration := m.queryStats.GetAverageDuration()
	slowestQuery := m.queryStats.GetSlowestQuery()
	fastestQuery := m.queryStats.GetFastestQuery()

	// Update counts
	totalQueries++
	if err != nil {
		failedQueries++
	} else {
		successfulQueries++
	}

	// Update average duration
	if totalQueries > 0 {
		totalDuration := averageDuration * time.Duration(totalQueries-1)
		averageDuration = (totalDuration + duration) / time.Duration(totalQueries)
	}

	// Update slowest/fastest query
	if slowestQuery == 0 || duration > slowestQuery {
		slowestQuery = duration
	}
	if fastestQuery == 0 || duration < fastestQuery {
		fastestQuery = duration
	}

	// Create new stats
	m.queryStats = health.NewQueryStats(totalQueries, successfulQueries, failedQueries, averageDuration, slowestQuery, fastestQuery)

	m.lastActivity = time.Now()
}

// RecordConnection records a connection event
func (m *MongoDBMonitor) RecordConnection(connected bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get current stats
	totalConnections := m.connStats.GetTotalConnections()
	connectionFailures := m.connStats.GetConnectionFailures()

	// Update counts
	totalConnections++
	if err != nil {
		connectionFailures++
	}

	// Create new stats
	m.connStats = health.NewConnectionStats(totalConnections, connectionFailures)

	m.lastActivity = time.Now()
}

// RecordTransaction records a transaction event
func (m *MongoDBMonitor) RecordTransaction(operation string, duration time.Duration, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get current stats
	totalTransactions := m.transStats.GetTotalTransactions()
	successfulTransactions := m.transStats.GetSuccessfulTransactions()
	failedTransactions := m.transStats.GetFailedTransactions()
	rolledBackTransactions := m.transStats.GetRolledBackTransactions()
	averageDuration := m.transStats.GetAverageDuration()

	// Update counts
	totalTransactions++
	if err != nil {
		failedTransactions++
		if operation == "rollback" {
			rolledBackTransactions++
		}
	} else {
		successfulTransactions++
	}

	// Update average duration
	if totalTransactions > 0 {
		totalDuration := averageDuration * time.Duration(totalTransactions-1)
		averageDuration = (totalDuration + duration) / time.Duration(totalTransactions)
	}

	// Create new stats
	m.transStats = health.NewTransactionStats(totalTransactions, successfulTransactions, failedTransactions, rolledBackTransactions, averageDuration)

	m.lastActivity = time.Now()
}

// GetQueryStats returns query statistics
func (m *MongoDBMonitor) GetQueryStats() database.QueryStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.queryStats
}

// GetConnectionStats returns connection statistics
func (m *MongoDBMonitor) GetConnectionStats() database.ConnectionStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connStats
}

// GetTransactionStats returns transaction statistics
func (m *MongoDBMonitor) GetTransactionStats() database.TransactionStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.transStats
}

// GetOverallStats returns overall statistics
func (m *MongoDBMonitor) GetOverallStats() database.OverallStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uptime := time.Since(m.startTime)
	overallStats := health.NewOverallStats(m.queryStats, m.connStats, m.transStats, uptime, m.lastActivity)

	return overallStats
}
