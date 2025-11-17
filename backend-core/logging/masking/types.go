package masking

import (
	"time"

	auditconfig "backend-core/audit/config"
	cacheconfig "backend-core/cache/config"

	"go.uber.org/zap/zapcore"
)

// MaskLevel represents the level of masking to apply
type MaskLevel int

const (
	NoMask      MaskLevel = iota // No masking (safe fields)
	PartialMask                  // Show first/last few characters
	FullMask                     // Completely mask with asterisks
	HashMask                     // Replace with hash
	RemoveField                  // Remove field entirely
)

// String returns the string representation of MaskLevel
func (m MaskLevel) String() string {
	switch m {
	case NoMask:
		return "no_mask"
	case PartialMask:
		return "partial_mask"
	case FullMask:
		return "full_mask"
	case HashMask:
		return "hash_mask"
	case RemoveField:
		return "remove_field"
	default:
		return "unknown"
	}
}

// SecurityLevel represents the security level for different environments
type SecurityLevel int

const (
	Development SecurityLevel = iota // Minimal masking
	Staging                          // Moderate masking
	Production                       // Full masking
	Compliance                       // Maximum masking (PCI, HIPAA)
)

// String returns the string representation of SecurityLevel
func (s SecurityLevel) String() string {
	switch s {
	case Development:
		return "development"
	case Staging:
		return "staging"
	case Production:
		return "production"
	case Compliance:
		return "compliance"
	default:
		return "unknown"
	}
}

// MaskingRule defines a rule for masking sensitive data
type MaskingRule struct {
	Field       string    `json:"field" yaml:"field"`                                 // Field name or pattern
	Level       MaskLevel `json:"level" yaml:"level"`                                 // Masking level
	Pattern     string    `json:"pattern,omitempty" yaml:"pattern,omitempty"`         // Custom pattern for partial masking
	Visible     int       `json:"visible,omitempty" yaml:"visible,omitempty"`         // Number of visible characters
	Algorithm   string    `json:"algorithm,omitempty" yaml:"algorithm,omitempty"`     // Hash algorithm for hash masking
	IsRegex     bool      `json:"is_regex" yaml:"is_regex"`                           // Whether field is a regex pattern
	Environment string    `json:"environment,omitempty" yaml:"environment,omitempty"` // Target environment
	UserRole    string    `json:"user_role,omitempty" yaml:"user_role,omitempty"`     // Target user role
}

// MaskingConfig holds the configuration for data masking
type MaskingConfig struct {
	Enabled       bool                   `json:"enabled" yaml:"enabled"`
	Environment   string                 `json:"environment" yaml:"environment"`
	SecurityLevel SecurityLevel          `json:"security_level" yaml:"security_level"`
	Rules         []MaskingRule          `json:"rules" yaml:"rules"`
	GlobalRules   []MaskingRule          `json:"global_rules" yaml:"global_rules"`
	FieldRules    map[string]MaskingRule `json:"field_rules" yaml:"field_rules"`
	Cache         cacheconfig.Config     `json:"cache" yaml:"cache"`
	Audit         auditconfig.Config     `json:"audit" yaml:"audit"`
}

// MaskingAudit represents an audit entry for masking operations
type MaskingAudit struct {
	Timestamp   time.Time `json:"timestamp"`
	Field       string    `json:"field"`
	OriginalLen int       `json:"original_length"`
	MaskedLen   int       `json:"masked_length"`
	Method      string    `json:"method"`
	User        string    `json:"user,omitempty"`
	Environment string    `json:"environment"`
	Rule        string    `json:"rule"`
}

// MaskingMetrics holds metrics for masking operations
type MaskingMetrics struct {
	FieldsMasked int64         `json:"fields_masked"`
	MaskingTime  time.Duration `json:"masking_time"`
	RulesMatched int64         `json:"rules_matched"`
	Errors       int64         `json:"errors"`
	CacheHits    int64         `json:"cache_hits"`
	CacheMisses  int64         `json:"cache_misses"`
}

// SensitiveDataMasker defines the interface for masking sensitive data
type SensitiveDataMasker interface {
	// MaskField masks a single field value
	MaskField(field string, value interface{}) (interface{}, error)

	// MaskFields masks multiple fields
	MaskFields(fields map[string]interface{}) (map[string]interface{}, error)

	// MaskZapFields masks zap fields
	MaskZapFields(fields []zapcore.Field) []zapcore.Field

	// ShouldMaskField determines if a field should be masked
	ShouldMaskField(field string) bool

	// GetMaskingRule returns the masking rule for a field
	GetMaskingRule(field string) (MaskingRule, bool)
}

// MaskingRuleLoader defines the interface for loading masking rules
type MaskingRuleLoader interface {
	// LoadRules loads masking rules from configuration
	LoadRules() ([]MaskingRule, error)

	// ReloadRules reloads masking rules
	ReloadRules() error

	// WatchForChanges watches for configuration changes
	WatchForChanges() <-chan []MaskingRule
}

// MaskingStrategy defines the interface for different masking strategies
type MaskingStrategy interface {
	// Mask applies masking to a value
	Mask(value string, rule MaskingRule) (string, error)

	// CanHandle determines if this strategy can handle the given rule
	CanHandle(rule MaskingRule) bool

	// Name returns the name of the strategy
	Name() string
}

// MaskingCache defines the interface for caching masking rules
type MaskingCache interface {
	// Get retrieves a masking rule from cache
	Get(field string) (MaskingRule, bool)

	// Set stores a masking rule in cache
	Set(field string, rule MaskingRule)

	// Clear clears the cache
	Clear()

	// Size returns the current cache size
	Size() int
}

// MaskingAuditor defines the interface for auditing masking operations
type MaskingAuditor interface {
	// LogAudit logs a masking operation
	LogAudit(audit MaskingAudit)

	// GetAuditLogs retrieves audit logs
	GetAuditLogs(filter AuditFilter) ([]MaskingAudit, error)

	// ExportAuditLogs exports audit logs
	ExportAuditLogs(format string) ([]byte, error)
}

// AuditFilter defines filters for audit log queries
type AuditFilter struct {
	Field       string    `json:"field,omitempty"`
	User        string    `json:"user,omitempty"`
	Environment string    `json:"environment,omitempty"`
	StartTime   time.Time `json:"start_time,omitempty"`
	EndTime     time.Time `json:"end_time,omitempty"`
	Limit       int       `json:"limit,omitempty"`
	Offset      int       `json:"offset,omitempty"`
}

// MaskingMetricsCollector defines the interface for collecting masking metrics
type MaskingMetricsCollector interface {
	// RecordMasking records a masking operation
	RecordMasking(field string, duration time.Duration)

	// RecordError records a masking error
	RecordError(field string, err error)

	// RecordCacheHit records a cache hit
	RecordCacheHit()

	// RecordCacheMiss records a cache miss
	RecordCacheMiss()

	// GetMetrics returns current metrics
	GetMetrics() MaskingMetrics

	// ResetMetrics resets metrics
	ResetMetrics()
}
