package providers

import (
	"backend-core/config"
)

// ConfigProvider creates a configuration provider
func ConfigProvider() *config.Config {
	return &config.Config{}
}
