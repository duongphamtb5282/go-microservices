package config

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

// SecurityConfig holds security configuration
type SecurityConfig struct {
	JWTSecret     string        `mapstructure:"jwt_secret" validate:"required,min=32"`
	JWTExpiration time.Duration `mapstructure:"jwt_expiration" validate:"required,min=1h"`
	Issuer        string        `mapstructure:"issuer" validate:"required"`
	Audience      string        `mapstructure:"audience" validate:"required"`
	BCryptCost    int           `mapstructure:"bcrypt_cost" validate:"required,min=4,max=31"`
}

// Validate validates the security configuration
func (c *SecurityConfig) Validate() error {
	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return fmt.Errorf("security configuration validation failed: %w", err)
	}
	return nil
}
