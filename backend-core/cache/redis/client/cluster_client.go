package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"backend-core/config"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// ClusterClientManager manages Redis cluster connections with enhanced features
type ClusterClientManager struct {
	config *config.RedisConfig
	client redis.UniversalClient
	logger *zap.Logger
}

// NewClusterClientManager creates a new cluster client manager
func NewClusterClientManager(config *config.RedisConfig, logger *zap.Logger) *ClusterClientManager {
	return &ClusterClientManager{
		config: config,
		logger: logger,
	}
}

// Connect establishes connection to Redis cluster
func (cm *ClusterClientManager) Connect(ctx context.Context) error {
	if cm.client != nil {
		return nil // Already connected
	}

	var client redis.UniversalClient
	var err error

	if cm.config.UseCluster {
		client, err = cm.createClusterClient()
	} else {
		client, err = cm.createStandaloneClient()
	}

	if err != nil {
		return fmt.Errorf("failed to create Redis client: %w", err)
	}

	// Test connection
	_, err = client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to ping Redis: %w", err)
	}

	cm.client = client
	cm.logger.Info("Redis connected successfully",
		zap.Bool("cluster_mode", cm.config.UseCluster),
		zap.Int("pool_size", cm.config.PoolSize),
		zap.Int("min_idle_conns", cm.config.MinIdleConns),
	)

	return nil
}

// GetClient returns the Redis client
func (cm *ClusterClientManager) GetClient() redis.UniversalClient {
	return cm.client
}

// Close closes the Redis connection
func (cm *ClusterClientManager) Close() error {
	if cm.client == nil {
		return nil
	}
	return cm.client.Close()
}

// createClusterClient creates a Redis cluster client
func (cm *ClusterClientManager) createClusterClient() (redis.UniversalClient, error) {
	if len(cm.config.ClusterAddrs) == 0 {
		return nil, fmt.Errorf("cluster addresses are required for cluster mode")
	}

	opts := &redis.ClusterOptions{
		Addrs:    cm.config.ClusterAddrs,
		Password: cm.config.Password,

		// Enhanced connection pooling
		PoolSize:     cm.config.PoolSize,
		MinIdleConns: cm.config.MinIdleConns,
		MaxRetries:   cm.config.MaxRetries,

		// Connection timeouts
		DialTimeout:  cm.parseDuration(cm.config.DialTimeout, 5*time.Second),
		ReadTimeout:  cm.parseDuration(cm.config.ReadTimeout, 3*time.Second),
		WriteTimeout: cm.parseDuration(cm.config.WriteTimeout, 3*time.Second),
		PoolTimeout:  cm.parseDuration(cm.config.PoolTimeout, 4*time.Second),

		// Idle connection management
		IdleTimeout:        cm.parseDuration(cm.config.IdleTimeout, 5*time.Minute),
		IdleCheckFrequency: cm.parseDuration(cm.config.IdleCheckFrequency, 1*time.Minute),

		// Cluster settings
		MaxRedirects: cm.config.ClusterMaxRedirects,

		// Connection settings
		MaxConnAge: cm.parseDuration(cm.config.MaxConnAge, 30*time.Minute),

		// Retry settings
		MinRetryBackoff: cm.parseDuration(cm.config.MinRetryBackoff, 8*time.Millisecond),
		MaxRetryBackoff: cm.parseDuration(cm.config.MaxRetryBackoff, 512*time.Millisecond),
	}

	// Add TLS if needed
	if strings.Contains(cm.config.ClusterAddrs[0], "rediss://") {
		opts.TLSConfig = &tls.Config{
			InsecureSkipVerify: false,
		}
	}

	return redis.NewClusterClient(opts), nil
}

// createStandaloneClient creates a standalone Redis client
func (cm *ClusterClientManager) createStandaloneClient() (redis.UniversalClient, error) {
	if cm.config.Addr == "" {
		return nil, fmt.Errorf("address is required for standalone mode")
	}

	opts := &redis.Options{
		Addr:     cm.config.Addr,
		Password: cm.config.Password,
		DB:       cm.config.DB,

		// Enhanced connection pooling
		PoolSize:     cm.config.PoolSize,
		MinIdleConns: cm.config.MinIdleConns,
		MaxRetries:   cm.config.MaxRetries,

		// Connection timeouts
		DialTimeout:  cm.parseDuration(cm.config.DialTimeout, 5*time.Second),
		ReadTimeout:  cm.parseDuration(cm.config.ReadTimeout, 3*time.Second),
		WriteTimeout: cm.parseDuration(cm.config.WriteTimeout, 3*time.Second),
		PoolTimeout:  cm.parseDuration(cm.config.PoolTimeout, 4*time.Second),

		// Idle connection management
		IdleTimeout:        cm.parseDuration(cm.config.IdleTimeout, 5*time.Minute),
		IdleCheckFrequency: cm.parseDuration(cm.config.IdleCheckFrequency, 1*time.Minute),

		// Connection settings
		MaxConnAge: cm.parseDuration(cm.config.MaxConnAge, 30*time.Minute),

		// Retry settings
		MinRetryBackoff: cm.parseDuration(cm.config.MinRetryBackoff, 8*time.Millisecond),
		MaxRetryBackoff: cm.parseDuration(cm.config.MaxRetryBackoff, 512*time.Millisecond),
	}

	// Add TLS if needed
	if strings.Contains(cm.config.Addr, "rediss://") {
		opts.TLSConfig = &tls.Config{
			InsecureSkipVerify: false,
		}
	}

	return redis.NewClient(opts), nil
}

// parseDuration parses duration string with default fallback
func (cm *ClusterClientManager) parseDuration(durationStr string, defaultValue time.Duration) time.Duration {
	if durationStr == "" {
		return defaultValue
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		cm.logger.Warn("Failed to parse duration, using default",
			zap.String("duration_str", durationStr),
			zap.Duration("default", defaultValue),
			zap.Error(err),
		)
		return defaultValue
	}

	return duration
}

// GetPoolStats returns connection pool statistics
func (cm *ClusterClientManager) GetPoolStats() *redis.PoolStats {
	if cm.client == nil {
		return &redis.PoolStats{}
	}

	switch client := cm.client.(type) {
	case *redis.Client:
		return client.PoolStats()
	case *redis.ClusterClient:
		// For cluster client, return aggregate stats
		return &redis.PoolStats{
			TotalConns: 0, // Would need to aggregate from all nodes
			IdleConns:  0,
			StaleConns: 0,
			Hits:       0,
			Misses:     0,
			Timeouts:   0,
		}
	default:
		return &redis.PoolStats{}
	}
}
