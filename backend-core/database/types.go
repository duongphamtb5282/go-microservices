package database

import (
	"backend-core/database/core"
	"backend-core/database/events"
	"backend-core/database/health"
	"backend-core/database/query"
	"backend-core/database/repository"
)

// Re-export all types from their respective packages for backward compatibility

// Core database types
type DatabaseType = core.DatabaseType
type Database = core.Database
type ExtendedDatabase = core.ExtendedDatabase
type QueryBuilder = core.QueryBuilder
type MigrationManager = core.MigrationManager
type Migration = core.Migration
type TransactionManager = core.TransactionManager
type Transaction = core.Transaction
type DatabaseHealthChecker = core.DatabaseHealthChecker
type DatabaseMonitor = core.DatabaseMonitor

// Repository types
type Repository[T any] = repository.Repository[T]
type Filter = repository.Filter
type Pagination = repository.Pagination
type Query = repository.Query
type CacheableRepository[T any] = repository.CacheableRepository[T]
type SearchableRepository[T any] = repository.SearchableRepository[T]

// Health types
type HealthStatus = health.HealthStatus
type HealthCheck = health.HealthCheck
type PerformanceMetrics = health.PerformanceMetrics
type ConnectionStats = health.ConnectionStats
type QueryStats = health.QueryStats
type TransactionStats = health.TransactionStats
type OverallStats = health.StatsResult

// Event sourcing types
type EventStoreRepository = events.EventStoreRepository
type Event = events.Event
type Snapshot = events.Snapshot

// Query types
type Index = query.Index
type IndexManager = query.IndexManager

// Config types
type CoreConfig = core.Config
type RepositoryConfig = repository.Config
type HealthConfig = health.Config
type EventsConfig = events.Config
type QueryConfig = query.Config

// Constants
const (
	PostgreSQL = core.PostgreSQL
	MongoDB    = core.MongoDB
)

const (
	HealthStatusHealthy   = health.HealthStatusHealthy
	HealthStatusDegraded  = health.HealthStatusDegraded
	HealthStatusUnhealthy = health.HealthStatusUnhealthy
	HealthStatusUnknown   = health.HealthStatusUnknown
)
