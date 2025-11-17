package reloader

import (
	"context"
	"fmt"
	"time"
)

// CustomReloadStrategy defines a custom reload strategy
type CustomReloadStrategy interface {
	// GetName returns the name of the strategy
	GetName() string

	// GetDescription returns the description of the strategy
	GetDescription() string

	// ShouldReload determines if a key should be reloaded
	ShouldReload(ctx context.Context, key string, lastReload time.Time) bool

	// GetReloadPriority returns the priority for reloading (lower = higher priority)
	GetReloadPriority(ctx context.Context, key string) int

	// PreReload is called before reloading a key
	PreReload(ctx context.Context, key string, data interface{}) error

	// PostReload is called after reloading a key
	PostReload(ctx context.Context, key string, data interface{}) error

	// OnReloadError is called when reloading fails
	OnReloadError(ctx context.Context, key string, err error) error
}

// TimeBasedReloadStrategy reloads based on time intervals
type TimeBasedReloadStrategy struct {
	name        string
	description string
	interval    time.Duration
	priority    int
}

// NewTimeBasedReloadStrategy creates a new time-based reload strategy
func NewTimeBasedReloadStrategy(name, description string, interval time.Duration, priority int) *TimeBasedReloadStrategy {
	return &TimeBasedReloadStrategy{
		name:        name,
		description: description,
		interval:    interval,
		priority:    priority,
	}
}

// GetName returns the name of the strategy
func (s *TimeBasedReloadStrategy) GetName() string {
	return s.name
}

// GetDescription returns the description of the strategy
func (s *TimeBasedReloadStrategy) GetDescription() string {
	return s.description
}

// ShouldReload determines if a key should be reloaded based on time
func (s *TimeBasedReloadStrategy) ShouldReload(ctx context.Context, key string, lastReload time.Time) bool {
	return time.Since(lastReload) >= s.interval
}

// GetReloadPriority returns the priority for reloading
func (s *TimeBasedReloadStrategy) GetReloadPriority(ctx context.Context, key string) int {
	return s.priority
}

// PreReload is called before reloading a key
func (s *TimeBasedReloadStrategy) PreReload(ctx context.Context, key string, data interface{}) error {
	// Default implementation - no action
	return nil
}

// PostReload is called after reloading a key
func (s *TimeBasedReloadStrategy) PostReload(ctx context.Context, key string, data interface{}) error {
	// Default implementation - no action
	return nil
}

// OnReloadError is called when reloading fails
func (s *TimeBasedReloadStrategy) OnReloadError(ctx context.Context, key string, err error) error {
	// Default implementation - return the error
	return err
}

// FrequencyBasedReloadStrategy reloads based on access frequency
type FrequencyBasedReloadStrategy struct {
	name           string
	description    string
	accessCount    int
	reloadInterval time.Duration
	priority       int
	accessCounts   map[string]int
	lastAccess     map[string]time.Time
}

// NewFrequencyBasedReloadStrategy creates a new frequency-based reload strategy
func NewFrequencyBasedReloadStrategy(name, description string, accessCount int, reloadInterval time.Duration, priority int) *FrequencyBasedReloadStrategy {
	return &FrequencyBasedReloadStrategy{
		name:           name,
		description:    description,
		accessCount:    accessCount,
		reloadInterval: reloadInterval,
		priority:       priority,
		accessCounts:   make(map[string]int),
		lastAccess:     make(map[string]time.Time),
	}
}

// GetName returns the name of the strategy
func (s *FrequencyBasedReloadStrategy) GetName() string {
	return s.name
}

// GetDescription returns the description of the strategy
func (s *FrequencyBasedReloadStrategy) GetDescription() string {
	return s.description
}

// ShouldReload determines if a key should be reloaded based on access frequency
func (s *FrequencyBasedReloadStrategy) ShouldReload(ctx context.Context, key string, lastReload time.Time) bool {
	// Check if enough time has passed since last reload
	if time.Since(lastReload) < s.reloadInterval {
		return false
	}

	// Check if the key has been accessed enough times
	count, exists := s.accessCounts[key]
	if !exists {
		return false
	}

	return count >= s.accessCount
}

// GetReloadPriority returns the priority for reloading
func (s *FrequencyBasedReloadStrategy) GetReloadPriority(ctx context.Context, key string) int {
	return s.priority
}

// PreReload is called before reloading a key
func (s *FrequencyBasedReloadStrategy) PreReload(ctx context.Context, key string, data interface{}) error {
	// Reset access count for this key
	s.accessCounts[key] = 0
	return nil
}

// PostReload is called after reloading a key
func (s *FrequencyBasedReloadStrategy) PostReload(ctx context.Context, key string, data interface{}) error {
	// Update last access time
	s.lastAccess[key] = time.Now()
	return nil
}

// OnReloadError is called when reloading fails
func (s *FrequencyBasedReloadStrategy) OnReloadError(ctx context.Context, key string, err error) error {
	// Increment access count on error to try again sooner
	s.accessCounts[key]++
	return err
}

// RecordAccess records an access to a key
func (s *FrequencyBasedReloadStrategy) RecordAccess(key string) {
	s.accessCounts[key]++
	s.lastAccess[key] = time.Now()
}

// ConditionalReloadStrategy reloads based on custom conditions
type ConditionalReloadStrategy struct {
	name        string
	description string
	condition   func(ctx context.Context, key string, lastReload time.Time) bool
	priority    int
}

// NewConditionalReloadStrategy creates a new conditional reload strategy
func NewConditionalReloadStrategy(name, description string, condition func(ctx context.Context, key string, lastReload time.Time) bool, priority int) *ConditionalReloadStrategy {
	return &ConditionalReloadStrategy{
		name:        name,
		description: description,
		condition:   condition,
		priority:    priority,
	}
}

// GetName returns the name of the strategy
func (s *ConditionalReloadStrategy) GetName() string {
	return s.name
}

// GetDescription returns the description of the strategy
func (s *ConditionalReloadStrategy) GetDescription() string {
	return s.description
}

// ShouldReload determines if a key should be reloaded based on custom condition
func (s *ConditionalReloadStrategy) ShouldReload(ctx context.Context, key string, lastReload time.Time) bool {
	if s.condition == nil {
		return false
	}
	return s.condition(ctx, key, lastReload)
}

// GetReloadPriority returns the priority for reloading
func (s *ConditionalReloadStrategy) GetReloadPriority(ctx context.Context, key string) int {
	return s.priority
}

// PreReload is called before reloading a key
func (s *ConditionalReloadStrategy) PreReload(ctx context.Context, key string, data interface{}) error {
	// Default implementation - no action
	return nil
}

// PostReload is called after reloading a key
func (s *ConditionalReloadStrategy) PostReload(ctx context.Context, key string, data interface{}) error {
	// Default implementation - no action
	return nil
}

// OnReloadError is called when reloading fails
func (s *ConditionalReloadStrategy) OnReloadError(ctx context.Context, key string, err error) error {
	// Default implementation - return the error
	return err
}

// ReloadStrategyManager manages multiple reload strategies
type ReloadStrategyManager struct {
	strategies []CustomReloadStrategy
}

// NewReloadStrategyManager creates a new reload strategy manager
func NewReloadStrategyManager() *ReloadStrategyManager {
	return &ReloadStrategyManager{
		strategies: make([]CustomReloadStrategy, 0),
	}
}

// AddStrategy adds a reload strategy
func (m *ReloadStrategyManager) AddStrategy(strategy CustomReloadStrategy) {
	m.strategies = append(m.strategies, strategy)
}

// GetStrategies returns all strategies
func (m *ReloadStrategyManager) GetStrategies() []CustomReloadStrategy {
	return m.strategies
}

// GetStrategyByName returns a strategy by name
func (m *ReloadStrategyManager) GetStrategyByName(name string) CustomReloadStrategy {
	for _, strategy := range m.strategies {
		if strategy.GetName() == name {
			return strategy
		}
	}
	return nil
}

// ShouldReload checks if any strategy indicates a key should be reloaded
func (m *ReloadStrategyManager) ShouldReload(ctx context.Context, key string, lastReload time.Time) bool {
	for _, strategy := range m.strategies {
		if strategy.ShouldReload(ctx, key, lastReload) {
			return true
		}
	}
	return false
}

// GetReloadPriority returns the highest priority (lowest number) for reloading
func (m *ReloadStrategyManager) GetReloadPriority(ctx context.Context, key string) int {
	if len(m.strategies) == 0 {
		return 0
	}

	priority := m.strategies[0].GetReloadPriority(ctx, key)
	for _, strategy := range m.strategies[1:] {
		if strategy.ShouldReload(ctx, key, time.Time{}) {
			p := strategy.GetReloadPriority(ctx, key)
			if p < priority {
				priority = p
			}
		}
	}
	return priority
}

// PreReload calls PreReload on all strategies
func (m *ReloadStrategyManager) PreReload(ctx context.Context, key string, data interface{}) error {
	for _, strategy := range m.strategies {
		if err := strategy.PreReload(ctx, key, data); err != nil {
			return fmt.Errorf("strategy %s pre-reload failed: %w", strategy.GetName(), err)
		}
	}
	return nil
}

// PostReload calls PostReload on all strategies
func (m *ReloadStrategyManager) PostReload(ctx context.Context, key string, data interface{}) error {
	for _, strategy := range m.strategies {
		if err := strategy.PostReload(ctx, key, data); err != nil {
			return fmt.Errorf("strategy %s post-reload failed: %w", strategy.GetName(), err)
		}
	}
	return nil
}

// OnReloadError calls OnReloadError on all strategies
func (m *ReloadStrategyManager) OnReloadError(ctx context.Context, key string, err error) error {
	var lastErr error
	for _, strategy := range m.strategies {
		if e := strategy.OnReloadError(ctx, key, err); e != nil {
			lastErr = e
		}
	}
	return lastErr
}

// GetStrategyInfo returns information about all strategies
func (m *ReloadStrategyManager) GetStrategyInfo() []map[string]interface{} {
	info := make([]map[string]interface{}, len(m.strategies))
	for i, strategy := range m.strategies {
		info[i] = map[string]interface{}{
			"name":        strategy.GetName(),
			"description": strategy.GetDescription(),
		}
	}
	return info
}

// ClearStrategies removes all strategies
func (m *ReloadStrategyManager) ClearStrategies() {
	m.strategies = make([]CustomReloadStrategy, 0)
}

// GetStrategyCount returns the number of strategies
func (m *ReloadStrategyManager) GetStrategyCount() int {
	return len(m.strategies)
}

// HasStrategy checks if a strategy with the given name exists
func (m *ReloadStrategyManager) HasStrategy(name string) bool {
	return m.GetStrategyByName(name) != nil
}

// RemoveStrategy removes a strategy by name
func (m *ReloadStrategyManager) RemoveStrategy(name string) bool {
	for i, strategy := range m.strategies {
		if strategy.GetName() == name {
			m.strategies = append(m.strategies[:i], m.strategies[i+1:]...)
			return true
		}
	}
	return false
}
