package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level" validate:"required,oneof=debug info warn error fatal"` // debug, info, warn, error
	Format     string `mapstructure:"format" validate:"required,oneof=json console"`               // json, console
	Output     string `mapstructure:"output" validate:"required,oneof=stdout stderr file"`         // stdout, stderr, file
	FilePath   string `mapstructure:"file_path" validate:"omitempty"`                              // Path to log file
	MaxSize    int    `mapstructure:"max_size" validate:"required,min=1"`                          // Max size in MB
	MaxBackups int    `mapstructure:"max_backups" validate:"required,min=0"`                       // Max number of backup files
	MaxAge     int    `mapstructure:"max_age" validate:"required,min=0"`                           // Max age in days
	Compress   bool   `mapstructure:"compress"`                                                    // Compress old log files
}

// Validate validates the logging configuration
func (c *LoggingConfig) Validate() error {
	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return fmt.Errorf("logging configuration validation failed: %w", err)
	}
	return nil
}
