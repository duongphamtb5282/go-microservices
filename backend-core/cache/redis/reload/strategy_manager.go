package reload

// CustomReloadStrategyManager manages custom reload strategies
type CustomReloadStrategyManager struct {
	strategies      map[string]CustomReloadStrategy
	defaultStrategy CustomReloadStrategy
}

// NewCustomReloadStrategyManager creates a new custom reload strategy manager
func NewCustomReloadStrategyManager() *CustomReloadStrategyManager {
	return &CustomReloadStrategyManager{
		strategies: make(map[string]CustomReloadStrategy),
	}
}

// RegisterStrategy registers a custom reload strategy
func (m *CustomReloadStrategyManager) RegisterStrategy(strategy CustomReloadStrategy) {
	m.strategies[strategy.GetName()] = strategy
}

// GetStrategy returns a strategy by name
func (m *CustomReloadStrategyManager) GetStrategy(name string) (CustomReloadStrategy, bool) {
	strategy, exists := m.strategies[name]
	return strategy, exists
}

// SetDefaultStrategy sets the default strategy
func (m *CustomReloadStrategyManager) SetDefaultStrategy(strategy CustomReloadStrategy) {
	m.defaultStrategy = strategy
}

// GetDefaultStrategy returns the default strategy
func (m *CustomReloadStrategyManager) GetDefaultStrategy() CustomReloadStrategy {
	return m.defaultStrategy
}

// ListStrategies returns all registered strategies
func (m *CustomReloadStrategyManager) ListStrategies() []CustomReloadStrategy {
	strategies := make([]CustomReloadStrategy, 0, len(m.strategies))
	for _, strategy := range m.strategies {
		strategies = append(strategies, strategy)
	}
	return strategies
}
