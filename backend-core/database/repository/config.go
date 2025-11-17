package repository

import (
	"time"
)

// Config holds repository configuration
type Config struct {
	// Caching settings
	CacheEnabled bool          `mapstructure:"cache_enabled" json:"cache_enabled" yaml:"cache_enabled"`
	CacheTTL     time.Duration `mapstructure:"cache_ttl" json:"cache_ttl" yaml:"cache_ttl"`
	CacheSize    int           `mapstructure:"cache_size" json:"cache_size" yaml:"cache_size"`

	// Search settings
	SearchEnabled     bool     `mapstructure:"search_enabled" json:"search_enabled" yaml:"search_enabled"`
	SearchIndexFields []string `mapstructure:"search_index_fields" json:"search_index_fields" yaml:"search_index_fields"`

	// Pagination settings
	DefaultPageSize int `mapstructure:"default_page_size" json:"default_page_size" yaml:"default_page_size"`
	MaxPageSize     int `mapstructure:"max_page_size" json:"max_page_size" yaml:"max_page_size"`

	// Query settings
	QueryTimeout time.Duration `mapstructure:"query_timeout" json:"query_timeout" yaml:"query_timeout"`
	MaxQuerySize int           `mapstructure:"max_query_size" json:"max_query_size" yaml:"max_query_size"`
}

// DefaultConfig returns a default repository configuration
func DefaultConfig() *Config {
	return &Config{
		CacheEnabled:      true,
		CacheTTL:          15 * time.Minute,
		CacheSize:         1000,
		SearchEnabled:     true,
		SearchIndexFields: []string{},
		DefaultPageSize:   20,
		MaxPageSize:       100,
		QueryTimeout:      30 * time.Second,
		MaxQuerySize:      1000,
	}
}
