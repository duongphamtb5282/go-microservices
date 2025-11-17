package config

import (
	"github.com/go-playground/validator/v10"
)

// Config holds audit configuration
type Config struct {
	Enabled     bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	LogLevel    string `mapstructure:"log_level" validate:"omitempty,oneof=debug info warn error" json:"log_level" yaml:"log_level"`
	IncludeData bool   `mapstructure:"include_data" json:"include_data" yaml:"include_data"`
}

// Validate checks if the audit configuration is valid
func (c *Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
