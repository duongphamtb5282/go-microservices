package masking

import (
	"fmt"
	"regexp"
	"sync"
	"time"

	"go.uber.org/zap/zapcore"
)

// SensitiveDataMaskerImpl implements the SensitiveDataMasker interface
type SensitiveDataMaskerImpl struct {
	config           *MaskingConfig
	strategyRegistry *StrategyRegistry
	cache            MaskingCache
	auditor          MaskingAuditor
	metricsCollector MaskingMetricsCollector
	compiledRules    map[string]*regexp.Regexp
	mutex            sync.RWMutex
}

// NewSensitiveDataMasker creates a new sensitive data masker
func NewSensitiveDataMasker(config *MaskingConfig) *SensitiveDataMaskerImpl {
	masker := &SensitiveDataMaskerImpl{
		config:           config,
		strategyRegistry: NewStrategyRegistry(),
		compiledRules:    make(map[string]*regexp.Regexp),
	}

	// Initialize cache if enabled
	if config.Cache.Enabled {
		masker.cache = NewInMemoryCache(config.Cache.Size, config.Cache.TTL)
	}

	// Initialize auditor if enabled
	if config.Audit.Enabled {
		masker.auditor = NewFileAuditor(config.Audit)
	}

	// Initialize metrics collector
	masker.metricsCollector = NewInMemoryMetricsCollector()

	// Compile regex rules
	masker.compileRules()

	return masker
}

// MaskField masks a single field value
func (m *SensitiveDataMaskerImpl) MaskField(field string, value interface{}) (interface{}, error) {
	start := time.Now()
	defer func() {
		m.metricsCollector.RecordMasking(field, time.Since(start))
	}()

	// Check if field should be masked
	if !m.ShouldMaskField(field) {
		return value, nil
	}

	// Get masking rule
	rule, exists := m.GetMaskingRule(field)
	if !exists {
		return value, nil
	}

	// Convert value to string
	valueStr, ok := value.(string)
	if !ok {
		valueStr = fmt.Sprintf("%v", value)
	}

	// Get appropriate strategy
	strategy := m.strategyRegistry.GetStrategyForRule(rule)

	// Apply masking
	maskedValue, err := strategy.Mask(valueStr, rule)
	if err != nil {
		m.metricsCollector.RecordError(field, err)
		return value, err
	}

	// Log audit if enabled
	if m.auditor != nil {
		audit := MaskingAudit{
			Timestamp:   time.Now(),
			Field:       field,
			OriginalLen: len(valueStr),
			MaskedLen:   len(maskedValue),
			Method:      strategy.Name(),
			Environment: m.config.Environment,
			Rule:        rule.Field,
		}
		m.auditor.LogAudit(audit)
	}

	return maskedValue, nil
}

// MaskFields masks multiple fields
func (m *SensitiveDataMaskerImpl) MaskFields(fields map[string]interface{}) (map[string]interface{}, error) {
	maskedFields := make(map[string]interface{})

	for field, value := range fields {
		maskedValue, err := m.MaskField(field, value)
		if err != nil {
			return nil, fmt.Errorf("failed to mask field %s: %w", field, err)
		}
		maskedFields[field] = maskedValue
	}

	return maskedFields, nil
}

// MaskZapFields masks zap fields
func (m *SensitiveDataMaskerImpl) MaskZapFields(fields []zapcore.Field) []zapcore.Field {
	maskedFields := make([]zapcore.Field, 0, len(fields))

	for _, field := range fields {
		if m.ShouldMaskField(field.Key) {
			rule, exists := m.GetMaskingRule(field.Key)
			if exists && rule.Level == RemoveField {
				// Skip this field entirely
				continue
			}

			// Mask the field value
			maskedValue, err := m.MaskField(field.Key, field.Interface)
			if err != nil {
				// If masking fails, keep original value
				maskedFields = append(maskedFields, field)
				continue
			}

			// Create new field with masked value
			maskedField := zapcore.Field{
				Key:       field.Key,
				Type:      field.Type,
				String:    field.String,
				Interface: maskedValue,
			}
			maskedFields = append(maskedFields, maskedField)
		} else {
			maskedFields = append(maskedFields, field)
		}
	}

	return maskedFields
}

// ShouldMaskField determines if a field should be masked
func (m *SensitiveDataMaskerImpl) ShouldMaskField(field string) bool {
	if !m.config.Enabled {
		return false
	}

	// Check cache first
	if m.cache != nil {
		if _, exists := m.cache.Get(field); exists {
			m.metricsCollector.RecordCacheHit()
			return true
		}
		m.metricsCollector.RecordCacheMiss()
	}

	// Check field rules
	if _, exists := m.config.FieldRules[field]; exists {
		if m.cache != nil {
			m.cache.Set(field, m.config.FieldRules[field])
		}
		return true
	}

	// Check specific rules
	for _, rule := range m.config.Rules {
		if m.matchesRule(field, rule) {
			if m.cache != nil {
				m.cache.Set(field, rule)
			}
			return true
		}
	}

	// Check global rules
	for _, rule := range m.config.GlobalRules {
		if m.matchesRule(field, rule) {
			if m.cache != nil {
				m.cache.Set(field, rule)
			}
			return true
		}
	}

	return false
}

// GetMaskingRule returns the masking rule for a field
func (m *SensitiveDataMaskerImpl) GetMaskingRule(field string) (MaskingRule, bool) {
	// Check cache first
	if m.cache != nil {
		if rule, exists := m.cache.Get(field); exists {
			return rule, true
		}
	}

	// Check field rules
	if rule, exists := m.config.FieldRules[field]; exists {
		if m.cache != nil {
			m.cache.Set(field, rule)
		}
		return rule, true
	}

	// Check specific rules
	for _, rule := range m.config.Rules {
		if m.matchesRule(field, rule) {
			if m.cache != nil {
				m.cache.Set(field, rule)
			}
			return rule, true
		}
	}

	// Check global rules
	for _, rule := range m.config.GlobalRules {
		if m.matchesRule(field, rule) {
			if m.cache != nil {
				m.cache.Set(field, rule)
			}
			return rule, true
		}
	}

	return MaskingRule{}, false
}

// matchesRule checks if a field matches a rule
func (m *SensitiveDataMaskerImpl) matchesRule(field string, rule MaskingRule) bool {
	// Check environment if specified
	if rule.Environment != "" && rule.Environment != m.config.Environment {
		return false
	}

	// Check if it's a regex rule
	if rule.IsRegex {
		m.mutex.RLock()
		compiled, exists := m.compiledRules[rule.Field]
		m.mutex.RUnlock()

		if !exists {
			// Compile regex
			compiled, err := regexp.Compile(rule.Field)
			if err != nil {
				return false
			}

			m.mutex.Lock()
			m.compiledRules[rule.Field] = compiled
			m.mutex.Unlock()
		}

		return compiled.MatchString(field)
	}

	// Exact match
	return field == rule.Field
}

// compileRules compiles all regex rules
func (m *SensitiveDataMaskerImpl) compileRules() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Compile specific rules
	for _, rule := range m.config.Rules {
		if rule.IsRegex {
			compiled, err := regexp.Compile(rule.Field)
			if err == nil {
				m.compiledRules[rule.Field] = compiled
			}
		}
	}

	// Compile global rules
	for _, rule := range m.config.GlobalRules {
		if rule.IsRegex {
			compiled, err := regexp.Compile(rule.Field)
			if err == nil {
				m.compiledRules[rule.Field] = compiled
			}
		}
	}
}

// ReloadConfig reloads the masking configuration
func (m *SensitiveDataMaskerImpl) ReloadConfig(config *MaskingConfig) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.config = config
	m.compiledRules = make(map[string]*regexp.Regexp)
	m.compileRules()

	// Clear cache
	if m.cache != nil {
		m.cache.Clear()
	}
}

// GetMetrics returns current masking metrics
func (m *SensitiveDataMaskerImpl) GetMetrics() MaskingMetrics {
	return m.metricsCollector.GetMetrics()
}

// ResetMetrics resets masking metrics
func (m *SensitiveDataMaskerImpl) ResetMetrics() {
	m.metricsCollector.ResetMetrics()
}

// GetAuditLogs returns audit logs
func (m *SensitiveDataMaskerImpl) GetAuditLogs(filter AuditFilter) ([]MaskingAudit, error) {
	if m.auditor == nil {
		return nil, fmt.Errorf("auditor not configured")
	}
	return m.auditor.GetAuditLogs(filter)
}

// ExportAuditLogs exports audit logs
func (m *SensitiveDataMaskerImpl) ExportAuditLogs(format string) ([]byte, error) {
	if m.auditor == nil {
		return nil, fmt.Errorf("auditor not configured")
	}
	return m.auditor.ExportAuditLogs(format)
}
