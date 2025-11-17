package logging

import (
	"backend-core/config"
	"context"
	"fmt"
	"time"
)

// ExampleGenericLogging demonstrates the new generic logging approach
func ExampleGenericLogging() {
	// Create a logger (this would normally be injected)
	logger, _ := NewLogger(&config.LoggingConfig{
		Level:  "debug",
		Format: "json",
		Output: "stdout",
	})

	// Example 1: Simple logging with generic parameters
	logger.Debug("user login attempt", "user_id", "123", "ip", "192.168.1.1", "timestamp", time.Now())

	// Example 2: Error logging with generic parameters
	err := fmt.Errorf("database connection failed")
	logger.Error("database error", "error", err, "database", "postgres", "retry_count", 3)

	// Example 3: Info logging with mixed types
	logger.Info("cache operation", "operation", "set", "key", "user:123", "ttl", 300*time.Second, "success", true)

	// Example 4: Complex data structures
	user := struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}{
		ID:    "123",
		Name:  "John Doe",
		Email: "john@example.com",
	}

	logger.Info("user created", "user", user, "source", "api", "version", "v1.2.3")

	// Example 5: With context
	ctx := context.WithValue(context.Background(), "request_id", "req-456")
	_ = ctx // Use context in real application
	logger.Debug("request processing", "step", "validation", "duration", 150*time.Millisecond)
}

// Benefits of Generic Logging:
// 1. No need for specific helper functions like logging.String(), logging.Int()
// 2. Type-safe - Go compiler handles type checking
// 3. Flexible - can pass any type of parameter
// 4. Simple - just pass key-value pairs as interface{}
// 5. Consistent - same pattern for all log levels
