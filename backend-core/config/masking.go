package config

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	auditconfig "backend-core/audit/config"
	cacheconfig "backend-core/cache/config"
	"backend-core/logging/masking"

	"github.com/go-playground/validator/v10"
)

// MaskingConfig holds masking configuration
type MaskingConfig struct {
	Enabled       bool                   `mapstructure:"enabled" validate:"required" json:"enabled" yaml:"enabled"`
	Environment   string                 `mapstructure:"environment" validate:"required,oneof=development staging production compliance" json:"environment" yaml:"environment"`
	SecurityLevel string                 `mapstructure:"security_level" validate:"required,oneof=development staging production compliance" json:"security_level" yaml:"security_level"`
	Rules         []MaskingRule          `mapstructure:"rules" json:"rules" yaml:"rules"`
	GlobalRules   []MaskingRule          `mapstructure:"global_rules" json:"global_rules" yaml:"global_rules"`
	FieldRules    map[string]MaskingRule `mapstructure:"field_rules" json:"field_rules" yaml:"field_rules"`
	Cache         cacheconfig.Config     `mapstructure:"cache" json:"cache" yaml:"cache"`
	Audit         auditconfig.Config     `mapstructure:"audit" json:"audit" yaml:"audit"`
}

// MaskingRule defines a masking rule
type MaskingRule struct {
	Field       string `mapstructure:"field" validate:"required" json:"field" yaml:"field"`
	Level       string `mapstructure:"level" validate:"required,oneof=no_mask partial_mask full_mask hash_mask remove_field" json:"level" yaml:"level"`
	Pattern     string `mapstructure:"pattern" json:"pattern" yaml:"pattern"`
	Visible     int    `mapstructure:"visible" validate:"min=0" json:"visible" yaml:"visible"`
	Algorithm   string `mapstructure:"algorithm" validate:"omitempty,oneof=sha256 sha1 md5" json:"algorithm" yaml:"algorithm"`
	IsRegex     bool   `mapstructure:"is_regex" json:"is_regex" yaml:"is_regex"`
	Environment string `mapstructure:"environment" validate:"omitempty,oneof=development staging production compliance" json:"environment" yaml:"environment"`
	UserRole    string `mapstructure:"user_role" json:"user_role" yaml:"user_role"`
}

// ToMaskingConfig converts to the masking package config
func (c *MaskingConfig) ToMaskingConfig() *masking.MaskingConfig {
	securityLevel := masking.Development
	switch c.SecurityLevel {
	case "staging":
		securityLevel = masking.Staging
	case "production":
		securityLevel = masking.Production
	case "compliance":
		securityLevel = masking.Compliance
	}

	rules := make([]masking.MaskingRule, len(c.Rules))
	for i, rule := range c.Rules {
		rules[i] = masking.MaskingRule{
			Field:       rule.Field,
			Level:       c.parseMaskLevel(rule.Level),
			Pattern:     rule.Pattern,
			Visible:     rule.Visible,
			Algorithm:   rule.Algorithm,
			IsRegex:     rule.IsRegex,
			Environment: rule.Environment,
			UserRole:    rule.UserRole,
		}
	}

	globalRules := make([]masking.MaskingRule, len(c.GlobalRules))
	for i, rule := range c.GlobalRules {
		globalRules[i] = masking.MaskingRule{
			Field:       rule.Field,
			Level:       c.parseMaskLevel(rule.Level),
			Pattern:     rule.Pattern,
			Visible:     rule.Visible,
			Algorithm:   rule.Algorithm,
			IsRegex:     rule.IsRegex,
			Environment: rule.Environment,
			UserRole:    rule.UserRole,
		}
	}

	fieldRules := make(map[string]masking.MaskingRule)
	for field, rule := range c.FieldRules {
		fieldRules[field] = masking.MaskingRule{
			Field:       rule.Field,
			Level:       c.parseMaskLevel(rule.Level),
			Pattern:     rule.Pattern,
			Visible:     rule.Visible,
			Algorithm:   rule.Algorithm,
			IsRegex:     rule.IsRegex,
			Environment: rule.Environment,
			UserRole:    rule.UserRole,
		}
	}

	return &masking.MaskingConfig{
		Enabled:       c.Enabled,
		Environment:   c.Environment,
		SecurityLevel: securityLevel,
		Rules:         rules,
		GlobalRules:   globalRules,
		FieldRules:    fieldRules,
		Cache: cacheconfig.Config{
			Enabled: c.Cache.Enabled,
			TTL:     c.Cache.TTL,
			Size:    c.Cache.Size,
		},
		Audit: auditconfig.Config{
			Enabled:     c.Audit.Enabled,
			LogLevel:    c.Audit.LogLevel,
			IncludeData: c.Audit.IncludeData,
		},
	}
}

// parseMaskLevel converts string to MaskLevel
func (c *MaskingConfig) parseMaskLevel(level string) masking.MaskLevel {
	switch level {
	case "no_mask":
		return masking.NoMask
	case "partial_mask":
		return masking.PartialMask
	case "full_mask":
		return masking.FullMask
	case "hash_mask":
		return masking.HashMask
	case "remove_field":
		return masking.RemoveField
	default:
		return masking.NoMask
	}
}

// Validate validates the masking configuration
func (c *MaskingConfig) Validate() error {
	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return fmt.Errorf("masking configuration validation failed: %w", err)
	}

	// Validate that at least one rule is defined if masking is enabled
	if c.Enabled && len(c.Rules) == 0 && len(c.GlobalRules) == 0 && len(c.FieldRules) == 0 {
		return fmt.Errorf("masking is enabled but no rules are defined")
	}

	// Validate rule patterns
	for i, rule := range c.Rules {
		if err := c.validateRule(rule, fmt.Sprintf("rules[%d]", i)); err != nil {
			return err
		}
	}

	for i, rule := range c.GlobalRules {
		if err := c.validateRule(rule, fmt.Sprintf("global_rules[%d]", i)); err != nil {
			return err
		}
	}

	for field, rule := range c.FieldRules {
		if err := c.validateRule(rule, fmt.Sprintf("field_rules[%s]", field)); err != nil {
			return err
		}
	}

	return nil
}

// validateRule validates a single masking rule
func (c *MaskingConfig) validateRule(rule MaskingRule, path string) error {
	// Validate regex patterns
	if rule.IsRegex {
		if _, err := regexp.Compile(rule.Field); err != nil {
			return fmt.Errorf("%s: invalid regex pattern '%s': %w", path, rule.Field, err)
		}
	}

	// Validate pattern-based masking
	if rule.Level == "full_mask" && rule.Pattern != "" {
		if err := c.validatePattern(rule.Pattern); err != nil {
			return fmt.Errorf("%s: invalid pattern '%s': %w", path, rule.Pattern, err)
		}
	}

	// Validate partial masking
	if rule.Level == "partial_mask" && rule.Visible < 0 {
		return fmt.Errorf("%s: visible characters must be non-negative for partial masking", path)
	}

	// Validate hash masking
	if rule.Level == "hash_mask" && rule.Algorithm != "" {
		validAlgorithms := []string{"sha256", "sha1", "md5"}
		valid := false
		for _, alg := range validAlgorithms {
			if rule.Algorithm == alg {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("%s: invalid hash algorithm '%s', must be one of: %v", path, rule.Algorithm, validAlgorithms)
		}
	}

	return nil
}

// validatePattern validates a masking pattern
func (c *MaskingConfig) validatePattern(pattern string) error {
	// Check for valid pattern characters
	validChars := ".*#X-"
	for _, char := range pattern {
		if !strings.ContainsRune(validChars, char) {
			return fmt.Errorf("invalid pattern character '%c', must be one of: %s", char, validChars)
		}
	}
	return nil
}

// GetDefaultMaskingConfig returns a default masking configuration
func GetDefaultMaskingConfig() *MaskingConfig {
	return &MaskingConfig{
		Enabled:       true,
		Environment:   "production",
		SecurityLevel: "production",
		Rules: []MaskingRule{
			{
				Field:       "password",
				Level:       "full_mask",
				Pattern:     "********",
				IsRegex:     false,
				Environment: "production",
			},
			{
				Field:       "secret",
				Level:       "full_mask",
				Pattern:     "********",
				IsRegex:     false,
				Environment: "production",
			},
			{
				Field:       "token",
				Level:       "partial_mask",
				Visible:     4,
				IsRegex:     false,
				Environment: "production",
			},
			{
				Field:       "email",
				Level:       "partial_mask",
				Visible:     3,
				Pattern:     "***@***.***",
				IsRegex:     false,
				Environment: "production",
			},
			{
				Field:       "credit_card",
				Level:       "full_mask",
				Pattern:     "****-****-****-####",
				IsRegex:     false,
				Environment: "production",
			},
		},
		GlobalRules: []MaskingRule{
			{
				Field:   ".*password.*",
				Level:   "full_mask",
				IsRegex: true,
			},
			{
				Field:   ".*secret.*",
				Level:   "full_mask",
				IsRegex: true,
			},
			{
				Field:   ".*key.*",
				Level:   "full_mask",
				IsRegex: true,
			},
		},
		FieldRules: make(map[string]MaskingRule),
		Cache: cacheconfig.Config{
			Enabled: true,
			TTL:     5 * time.Minute,
			Size:    1000,
		},
		Audit: auditconfig.Config{
			Enabled:     true,
			LogLevel:    "info",
			IncludeData: false,
		},
	}
}
