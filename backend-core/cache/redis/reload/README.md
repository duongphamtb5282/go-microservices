# Custom Reload Strategies

This package provides custom reload strategies for cache reloading, following the Single Responsibility Principle (SRP).

## Package Structure

```
backend-core/cache/redis/reload/
├── strategy_interface.go        # CustomReloadStrategy interface
├── strategy_manager.go          # CustomReloadStrategyManager
├── time_based_strategy.go      # TimeBasedReloadStrategy
├── priority_based_strategy.go  # PriorityBasedReloadStrategy
├── conditional_strategy.go     # ConditionalReloadStrategy
├── smart_strategy.go           # SmartReloadStrategy
└── README.md                   # This file
```

## Strategy Interface

All reload strategies implement the `CustomReloadStrategy` interface:

```go
type CustomReloadStrategy interface {
    GetName() string
    GetDescription() string
    ShouldReload(ctx context.Context, key string, lastReload time.Time) bool
    GetReloadPriority(ctx context.Context, key string) int
    PreReload(ctx context.Context, key string, data interface{}) error
    PostReload(ctx context.Context, key string, data interface{}) error
    OnReloadError(ctx context.Context, key string, err error) error
    GetBatchSize() int
    GetRetryCount() int
    GetRetryDelay() time.Duration
}
```

## Available Strategies

### 1. TimeBasedReloadStrategy

Reloads data based on time intervals.

```go
strategy := NewTimeBasedReloadStrategy("hourly", time.Hour, 100)
```

**Features:**

- Configurable time intervals
- Batch size control
- Retry configuration

### 2. PriorityBasedReloadStrategy

Reloads data based on key priority.

```go
priorities := map[string]int{
    "user:1": 1,    // High priority
    "user:2": 2,    // Medium priority
    "user:3": 3,    // Low priority
}
strategy := NewPriorityBasedReloadStrategy("priority", priorities, 50)
```

**Features:**

- Key-specific priority mapping
- Priority-based reload decisions
- Configurable batch sizes

### 3. ConditionalReloadStrategy

Reloads data based on custom conditions.

```go
condition := func(ctx context.Context, key string, lastReload time.Time) bool {
    return time.Since(lastReload) > time.Hour && strings.Contains(key, "critical")
}
strategy := NewConditionalReloadStrategy("conditional", condition, 25)
```

**Features:**

- Custom condition functions
- Flexible reload logic
- Context-aware decisions

### 4. SmartReloadStrategy

Uses access patterns to determine reload timing.

```go
strategy := NewSmartReloadStrategy("smart", 75)
```

**Features:**

- Access pattern tracking
- Frequency-based reload decisions
- Machine learning ready

## Strategy Manager

The `CustomReloadStrategyManager` manages multiple strategies:

```go
manager := NewCustomReloadStrategyManager()

// Register strategies
manager.RegisterStrategy(timeBasedStrategy)
manager.RegisterStrategy(priorityBasedStrategy)
manager.RegisterStrategy(conditionalStrategy)
manager.RegisterStrategy(smartStrategy)

// Set default strategy
manager.SetDefaultStrategy(timeBasedStrategy)

// Get strategy by name
strategy, exists := manager.GetStrategy("time-based")

// List all strategies
strategies := manager.ListStrategies()
```

## Usage Examples

### Basic Usage

```go
package main

import (
    "context"
    "time"

    "backend-core/cache/redis/reload"
)

func main() {
    // Create strategy manager
    manager := reload.NewCustomReloadStrategyManager()

    // Create and register time-based strategy
    timeStrategy := reload.NewTimeBasedReloadStrategy("hourly", time.Hour, 100)
    manager.RegisterStrategy(timeStrategy)

    // Create and register priority-based strategy
    priorities := map[string]int{
        "critical:*": 1,
        "important:*": 2,
        "normal:*": 3,
    }
    priorityStrategy := reload.NewPriorityBasedReloadStrategy("priority", priorities, 50)
    manager.RegisterStrategy(priorityStrategy)

    // Set default strategy
    manager.SetDefaultStrategy(timeStrategy)

    // Use strategies
    strategy, exists := manager.GetStrategy("hourly")
    if exists {
        shouldReload := strategy.ShouldReload(context.Background(), "user:1", time.Now().Add(-2*time.Hour))
        fmt.Printf("Should reload: %v\n", shouldReload)
    }
}
```

### Advanced Usage with Custom Conditions

```go
package main

import (
    "context"
    "strings"
    "time"

    "backend-core/cache/redis/reload"
)

func main() {
    // Create custom condition
    condition := func(ctx context.Context, key string, lastReload time.Time) bool {
        // Reload if key contains "critical" and hasn't been reloaded in 30 minutes
        return strings.Contains(key, "critical") && time.Since(lastReload) > 30*time.Minute
    }

    // Create conditional strategy
    strategy := reload.NewConditionalReloadStrategy("critical-reload", condition, 25)

    // Test the strategy
    ctx := context.Background()
    shouldReload := strategy.ShouldReload(ctx, "critical:user:1", time.Now().Add(-45*time.Minute))
    fmt.Printf("Should reload critical key: %v\n", shouldReload)
}
```

### Smart Strategy with Access Tracking

```go
package main

import (
    "context"
    "time"

    "backend-core/cache/redis/reload"
)

func main() {
    // Create smart strategy
    strategy := reload.NewSmartReloadStrategy("smart", 75)

    // Simulate access patterns
    ctx := context.Background()

    // Track access (this would normally be called by the cache system)
    strategy.PreReload(ctx, "user:1", nil) // Access count: 1
    strategy.PreReload(ctx, "user:1", nil) // Access count: 2
    // ... more accesses

    // Check if should reload based on access patterns
    shouldReload := strategy.ShouldReload(ctx, "user:1", time.Now().Add(-time.Hour))
    fmt.Printf("Should reload based on access patterns: %v\n", shouldReload)
}
```

## Strategy Selection

### By Name

```go
strategy, exists := manager.GetStrategy("time-based")
if exists {
    // Use the strategy
}
```

### By Priority

```go
strategies := manager.ListStrategies()
for _, strategy := range strategies {
    priority := strategy.GetReloadPriority(ctx, key)
    if priority <= 2 { // High priority
        // Use this strategy
    }
}
```

### Default Strategy

```go
defaultStrategy := manager.GetDefaultStrategy()
if defaultStrategy != nil {
    // Use default strategy
}
```

## Configuration

### Batch Sizes

Different strategies support different batch sizes:

- **TimeBasedReloadStrategy**: 100 (default)
- **PriorityBasedReloadStrategy**: 50 (default)
- **ConditionalReloadStrategy**: 25 (default)
- **SmartReloadStrategy**: 75 (default)

### Retry Configuration

All strategies support retry configuration:

- **Retry Count**: 3 (default)
- **Retry Delay**: 1 second (default)

### Custom Configuration

```go
// Create strategy with custom configuration
strategy := &reload.TimeBasedReloadStrategy{
    name:       "custom",
    interval:   30 * time.Minute,
    batchSize:  200,
    retryCount: 5,
    retryDelay: 2 * time.Second,
}
```

## Best Practices

### 1. Strategy Selection

- Use **TimeBasedReloadStrategy** for regular, predictable reloads
- Use **PriorityBasedReloadStrategy** for critical data that needs frequent updates
- Use **ConditionalReloadStrategy** for complex business logic
- Use **SmartReloadStrategy** for adaptive, access-pattern-based reloads

### 2. Performance Considerations

- Set appropriate batch sizes based on your data volume
- Use retry configuration to handle transient failures
- Monitor strategy performance and adjust parameters

### 3. Error Handling

- Always check for errors in strategy methods
- Implement proper error handling in your reload logic
- Use strategy error callbacks for logging and monitoring

### 4. Testing

- Test each strategy independently
- Use mock data for testing
- Verify strategy behavior under different conditions

## Migration from Monolithic File

The strategies have been separated from the monolithic `custom_reload_strategies.go` file:

**Before:**

```go
// All strategies in one file
import "backend-core/cache/redis/reload"
```

**After:**

```go
// Individual strategy files
import "backend-core/cache/redis/reload"
// All strategies are now in separate files but same package
```

**Benefits:**

- ✅ **Single Responsibility**: Each file has one strategy
- ✅ **Better Maintainability**: Easier to modify individual strategies
- ✅ **Improved Testing**: Test strategies independently
- ✅ **Cleaner Code**: Smaller, focused files
- ✅ **Better Organization**: Logical file structure

## Future Enhancements

### 1. Additional Strategies

- **LoadBasedReloadStrategy**: Reload based on system load
- **MemoryBasedReloadStrategy**: Reload based on memory usage
- **NetworkBasedReloadStrategy**: Reload based on network conditions

### 2. Strategy Composition

- **CompositeReloadStrategy**: Combine multiple strategies
- **FallbackReloadStrategy**: Use fallback strategies
- **WeightedReloadStrategy**: Weighted strategy selection

### 3. Advanced Features

- **Strategy Metrics**: Track strategy performance
- **Dynamic Configuration**: Runtime strategy configuration
- **Strategy Validation**: Validate strategy configurations
