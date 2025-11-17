package client

import (
	"backend-core/config"

	"github.com/go-redis/redis/v8"
)

// RedisClientFactory is responsible for creating Redis clients
type RedisClientFactory struct{}

// NewRedisClientFactory creates a new RedisClientFactory
func NewRedisClientFactory() *RedisClientFactory {
	return &RedisClientFactory{}
}

// CreateClient creates a new Redis client (standalone or cluster) based on config
func (f *RedisClientFactory) CreateClient(cfg *config.RedisConfig) redis.UniversalClient {
	if cfg.UseCluster {
		// Cluster mode
		return redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    cfg.ClusterAddrs,
			Password: cfg.Password,
			PoolSize: cfg.PoolSize,
		})
	} else {
		// Standalone mode
		return redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
			PoolSize: cfg.PoolSize,
		})
	}
}
