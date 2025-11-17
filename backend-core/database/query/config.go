package query

import (
	"time"
)

// Config holds query building configuration
type Config struct {
	// Index settings
	IndexEnabled         bool          `mapstructure:"index_enabled" json:"index_enabled" yaml:"index_enabled"`
	AutoCreateIndexes    bool          `mapstructure:"auto_create_indexes" json:"auto_create_indexes" yaml:"auto_create_indexes"`
	IndexCreationTimeout time.Duration `mapstructure:"index_creation_timeout" json:"index_creation_timeout" yaml:"index_creation_timeout"`

	// Query optimization
	QueryOptimization  bool          `mapstructure:"query_optimization" json:"query_optimization" yaml:"query_optimization"`
	MaxQueryComplexity int           `mapstructure:"max_query_complexity" json:"max_query_complexity" yaml:"max_query_complexity"`
	QueryCacheSize     int           `mapstructure:"query_cache_size" json:"query_cache_size" yaml:"query_cache_size"`
	QueryCacheTTL      time.Duration `mapstructure:"query_cache_ttl" json:"query_cache_ttl" yaml:"query_cache_ttl"`

	// Performance settings
	MaxConcurrentQueries int           `mapstructure:"max_concurrent_queries" json:"max_concurrent_queries" yaml:"max_concurrent_queries"`
	QueryTimeout         time.Duration `mapstructure:"query_timeout" json:"query_timeout" yaml:"query_timeout"`
	SlowQueryThreshold   time.Duration `mapstructure:"slow_query_threshold" json:"slow_query_threshold" yaml:"slow_query_threshold"`
}

// DefaultConfig returns a default query configuration
func DefaultConfig() *Config {
	return &Config{
		IndexEnabled:         true,
		AutoCreateIndexes:    true,
		IndexCreationTimeout: 30 * time.Second,
		QueryOptimization:    true,
		MaxQueryComplexity:   10,
		QueryCacheSize:       1000,
		QueryCacheTTL:        15 * time.Minute,
		MaxConcurrentQueries: 100,
		QueryTimeout:         30 * time.Second,
		SlowQueryThreshold:   1 * time.Second,
	}
}
