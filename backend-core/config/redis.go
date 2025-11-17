package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// RedisConfig holds enhanced Redis configuration for production
type RedisConfig struct {
	// Basic configuration
	Name     string `mapstructure:"name" json:"name" yaml:"name" validate:"required"`
	Addr     string `mapstructure:"addr" json:"addr" yaml:"addr"`
	Password string `mapstructure:"password" json:"password" yaml:"password"`
	DB       int    `mapstructure:"db" json:"db" yaml:"db" validate:"min=0,max=15"`

	// Enhanced connection pooling for production
	PoolSize     int `mapstructure:"pool_size" validate:"required,min=1"`
	MinIdleConns int `mapstructure:"min_idle_conns" validate:"min=0"`
	MaxRetries   int `mapstructure:"max_retries" validate:"min=0"`

	// Connection timeouts
	DialTimeout        string `mapstructure:"dial_timeout"`
	ReadTimeout        string `mapstructure:"read_timeout"`
	WriteTimeout       string `mapstructure:"write_timeout"`
	PoolTimeout        string `mapstructure:"pool_timeout"`
	IdleTimeout        string `mapstructure:"idle_timeout"`
	IdleCheckFrequency string `mapstructure:"idle_check_frequency"`

	// Cluster configuration
	UseCluster          bool     `mapstructure:"use_cluster" json:"use_cluster" yaml:"use_cluster"`
	ClusterAddrs        []string `mapstructure:"cluster_addrs" json:"cluster_addrs" yaml:"cluster_addrs"`
	ClusterMaxRedirects int      `mapstructure:"cluster_max_redirects" validate:"min=1"`

	// Performance optimization
	PipelineWindow      string `mapstructure:"pipeline_window"`
	PipelineLimit       int    `mapstructure:"pipeline_limit" validate:"min=1"`
	PipelineConcurrency int    `mapstructure:"pipeline_concurrency" validate:"min=1"`

	// Connection settings
	MaxConnAge      string `mapstructure:"max_conn_age"`
	PoolFifo        bool   `mapstructure:"pool_fifo"`
	MinRetryBackoff string `mapstructure:"min_retry_backoff"`
	MaxRetryBackoff string `mapstructure:"max_retry_backoff"`
}

// GetRedisAddr returns Redis address
func (c *RedisConfig) GetRedisAddr() string {
	return c.Addr
}

// GetClusterAddrs returns cluster addresses
func (c *RedisConfig) GetClusterAddrs() []string {
	return c.ClusterAddrs
}

// IsCluster returns true if cluster mode is enabled
func (c *RedisConfig) IsCluster() bool {
	return c.UseCluster
}

// Validate validates the Redis configuration
func (c *RedisConfig) Validate() error {
	validate := validator.New()

	// Basic validation
	if err := validate.Struct(c); err != nil {
		return fmt.Errorf("redis configuration validation failed: %w", err)
	}

	// Validate based on mode
	if c.UseCluster {
		// Cluster mode validation
		if len(c.ClusterAddrs) == 0 {
			return fmt.Errorf("cluster mode enabled but no cluster addresses provided")
		}
		// DB should be 0 in cluster mode
		if c.DB != 0 {
			return fmt.Errorf("DB must be 0 in cluster mode")
		}
	} else {
		// Standalone mode validation
		if c.Addr == "" {
			return fmt.Errorf("addr is required for standalone mode")
		}
	}

	return nil
}
