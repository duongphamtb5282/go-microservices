package masking

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// NoMaskingStrategy implements no masking
type NoMaskingStrategy struct{}

func (s *NoMaskingStrategy) Mask(value string, rule MaskingRule) (string, error) {
	return value, nil
}

func (s *NoMaskingStrategy) CanHandle(rule MaskingRule) bool {
	return rule.Level == NoMask
}

func (s *NoMaskingStrategy) Name() string {
	return "no_masking"
}

// PartialMaskingStrategy implements partial masking
type PartialMaskingStrategy struct{}

func (s *PartialMaskingStrategy) Mask(value string, rule MaskingRule) (string, error) {
	if len(value) == 0 {
		return value, nil
	}

	visible := rule.Visible
	if visible <= 0 {
		visible = 2 // Default to showing 2 characters
	}

	if len(value) <= visible*2 {
		// If value is too short, use full mask
		return strings.Repeat("*", len(value)), nil
	}

	// Show first and last visible characters
	start := value[:visible]
	end := value[len(value)-visible:]
	middle := strings.Repeat("*", len(value)-visible*2)

	return start + middle + end, nil
}

func (s *PartialMaskingStrategy) CanHandle(rule MaskingRule) bool {
	return rule.Level == PartialMask
}

func (s *PartialMaskingStrategy) Name() string {
	return "partial_masking"
}

// FullMaskingStrategy implements full masking
type FullMaskingStrategy struct{}

func (s *FullMaskingStrategy) Mask(value string, rule MaskingRule) (string, error) {
	if len(value) == 0 {
		return value, nil
	}

	if rule.Pattern != "" {
		// Use custom pattern
		pattern := strings.ReplaceAll(rule.Pattern, "*", "\\*")
		pattern = strings.ReplaceAll(pattern, "#", "\\d")
		pattern = strings.ReplaceAll(pattern, "X", "\\w")
		return pattern, nil
	}

	// Default to asterisks
	return strings.Repeat("*", len(value)), nil
}

func (s *FullMaskingStrategy) CanHandle(rule MaskingRule) bool {
	return rule.Level == FullMask
}

func (s *FullMaskingStrategy) Name() string {
	return "full_masking"
}

// HashMaskingStrategy implements hash masking
type HashMaskingStrategy struct{}

func (s *HashMaskingStrategy) Mask(value string, rule MaskingRule) (string, error) {
	if len(value) == 0 {
		return value, nil
	}

	algorithm := rule.Algorithm
	if algorithm == "" {
		algorithm = "sha256"
	}

	var hash string
	switch strings.ToLower(algorithm) {
	case "sha256":
		hashBytes := sha256.Sum256([]byte(value))
		hash = fmt.Sprintf("sha256:%x", hashBytes)
	case "sha1":
		// Note: In production, you'd use crypto/sha1
		hashBytes := sha256.Sum256([]byte(value))
		hash = fmt.Sprintf("sha1:%x", hashBytes[:20])
	case "md5":
		// Note: In production, you'd use crypto/md5
		hashBytes := sha256.Sum256([]byte(value))
		hash = fmt.Sprintf("md5:%x", hashBytes[:16])
	default:
		hashBytes := sha256.Sum256([]byte(value))
		hash = fmt.Sprintf("hash:%x", hashBytes)
	}

	return hash, nil
}

func (s *HashMaskingStrategy) CanHandle(rule MaskingRule) bool {
	return rule.Level == HashMask
}

func (s *HashMaskingStrategy) Name() string {
	return "hash_masking"
}

// PatternMaskingStrategy implements pattern-based masking
type PatternMaskingStrategy struct{}

func (s *PatternMaskingStrategy) Mask(value string, rule MaskingRule) (string, error) {
	if len(value) == 0 {
		return value, nil
	}

	pattern := rule.Pattern
	if pattern == "" {
		// Default pattern based on value length
		pattern = strings.Repeat("*", len(value))
	}

	// Apply pattern masking
	result := s.applyPattern(value, pattern)
	return result, nil
}

func (s *PatternMaskingStrategy) applyPattern(value, pattern string) string {
	if len(pattern) != len(value) {
		// If pattern length doesn't match, use full mask
		return strings.Repeat("*", len(value))
	}

	result := make([]byte, len(value))
	for i, char := range pattern {
		switch char {
		case '*':
			result[i] = '*'
		case '#':
			// Keep digits
			if value[i] >= '0' && value[i] <= '9' {
				result[i] = value[i]
			} else {
				result[i] = '*'
			}
		case 'X':
			// Keep alphanumeric
			if (value[i] >= 'a' && value[i] <= 'z') ||
				(value[i] >= 'A' && value[i] <= 'Z') ||
				(value[i] >= '0' && value[i] <= '9') {
				result[i] = value[i]
			} else {
				result[i] = '*'
			}
		default:
			// Keep the character from pattern
			result[i] = byte(char)
		}
	}

	return string(result)
}

func (s *PatternMaskingStrategy) CanHandle(rule MaskingRule) bool {
	return rule.Level == FullMask && rule.Pattern != ""
}

func (s *PatternMaskingStrategy) Name() string {
	return "pattern_masking"
}

// RemoveFieldStrategy implements field removal
type RemoveFieldStrategy struct{}

func (s *RemoveFieldStrategy) Mask(value string, rule MaskingRule) (string, error) {
	return "", nil
}

func (s *RemoveFieldStrategy) CanHandle(rule MaskingRule) bool {
	return rule.Level == RemoveField
}

func (s *RemoveFieldStrategy) Name() string {
	return "remove_field"
}

// StrategyRegistry manages masking strategies
type StrategyRegistry struct {
	strategies map[string]MaskingStrategy
}

// NewStrategyRegistry creates a new strategy registry
func NewStrategyRegistry() *StrategyRegistry {
	registry := &StrategyRegistry{
		strategies: make(map[string]MaskingStrategy),
	}

	// Register default strategies
	registry.Register(&NoMaskingStrategy{})
	registry.Register(&PartialMaskingStrategy{})
	registry.Register(&FullMaskingStrategy{})
	registry.Register(&HashMaskingStrategy{})
	registry.Register(&PatternMaskingStrategy{})
	registry.Register(&RemoveFieldStrategy{})

	return registry
}

// Register registers a masking strategy
func (r *StrategyRegistry) Register(strategy MaskingStrategy) {
	r.strategies[strategy.Name()] = strategy
}

// GetStrategy returns a strategy by name
func (r *StrategyRegistry) GetStrategy(name string) (MaskingStrategy, bool) {
	strategy, exists := r.strategies[name]
	return strategy, exists
}

// GetStrategyForRule returns the appropriate strategy for a rule
func (r *StrategyRegistry) GetStrategyForRule(rule MaskingRule) MaskingStrategy {
	for _, strategy := range r.strategies {
		if strategy.CanHandle(rule) {
			return strategy
		}
	}
	return &NoMaskingStrategy{} // Fallback to no masking
}

// ListStrategies returns all registered strategies
func (r *StrategyRegistry) ListStrategies() []string {
	strategies := make([]string, 0, len(r.strategies))
	for name := range r.strategies {
		strategies = append(strategies, name)
	}
	return strategies
}
