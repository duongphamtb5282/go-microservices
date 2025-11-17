package config

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// Config holds cache configuration
type Config struct {
	Enabled bool          `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	TTL     time.Duration `mapstructure:"ttl" validate:"min=1s" json:"ttl" yaml:"ttl"`
	Size    int           `mapstructure:"size" validate:"min=1" json:"size" yaml:"size"`
}

// Validate checks if the cache configuration is valid
func (c *Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
