package strategies

import (
	"fmt"
	"time"
)

// StrategyStats holds statistics for a cache strategy
type StrategyStats struct {
	Name             string        `json:"name"`
	Hits             int64         `json:"hits"`
	Misses           int64         `json:"misses"`
	Writes           int64         `json:"writes"`
	Deletes          int64         `json:"deletes"`
	Errors           int64         `json:"errors"`
	AverageReadTime  time.Duration `json:"average_read_time"`
	AverageWriteTime time.Duration `json:"average_write_time"`
	LastUsed         time.Time     `json:"last_used"`
}

// StrategyManager manages cache strategies
type StrategyManager struct {
	strategies      map[string]CacheStrategy
	defaultStrategy string
}

// NewStrategyManager creates a new strategy manager
func NewStrategyManager() *StrategyManager {
	return &StrategyManager{
		strategies: make(map[string]CacheStrategy),
	}
}

// RegisterStrategy registers a cache strategy
func (m *StrategyManager) RegisterStrategy(strategy CacheStrategy) {
	m.strategies[strategy.GetName()] = strategy
}

// GetStrategy returns a strategy by name
func (m *StrategyManager) GetStrategy(name string) (CacheStrategy, bool) {
	strategy, exists := m.strategies[name]
	return strategy, exists
}

// SetDefaultStrategy sets the default strategy
func (m *StrategyManager) SetDefaultStrategy(name string) error {
	if _, exists := m.strategies[name]; !exists {
		return fmt.Errorf("strategy %s not found", name)
	}
	m.defaultStrategy = name
	return nil
}

// GetDefaultStrategy returns the default strategy
func (m *StrategyManager) GetDefaultStrategy() (CacheStrategy, bool) {
	return m.GetStrategy(m.defaultStrategy)
}

// GetAllStrategies returns all registered strategies
func (m *StrategyManager) GetAllStrategies() map[string]CacheStrategy {
	return m.strategies
}
