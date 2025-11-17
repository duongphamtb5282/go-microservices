package providers

import (
	"backend-core/config"
	"backend-core/logging"
)

// LoggerProvider creates a logger provider
func LoggerProvider(cfg *config.Config) *logging.Logger {
	logger, _ := logging.NewLogger(&cfg.Logging)
	return logger
}
