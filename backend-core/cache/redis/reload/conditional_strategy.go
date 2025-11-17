package reload

import (
	"context"
	"fmt"
	"time"
)

// ConditionalReloadStrategy reloads based on custom conditions
type ConditionalReloadStrategy struct {
	name       string
	condition  func(context.Context, string, time.Time) bool
	batchSize  int
	retryCount int
	retryDelay time.Duration
}

// NewConditionalReloadStrategy creates a new conditional reload strategy
func NewConditionalReloadStrategy(name string, condition func(context.Context, string, time.Time) bool, batchSize int) *ConditionalReloadStrategy {
	return &ConditionalReloadStrategy{
		name:       name,
		condition:  condition,
		batchSize:  batchSize,
		retryCount: 3,
		retryDelay: time.Second,
	}
}

func (s *ConditionalReloadStrategy) GetName() string {
	return s.name
}

func (s *ConditionalReloadStrategy) GetDescription() string {
	return "Reloads data based on custom conditions"
}

func (s *ConditionalReloadStrategy) ShouldReload(ctx context.Context, key string, lastReload time.Time) bool {
	if s.condition != nil {
		return s.condition(ctx, key, lastReload)
	}
	return false
}

func (s *ConditionalReloadStrategy) GetReloadPriority(ctx context.Context, key string) int {
	// Default priority based on key pattern
	if len(key) > 10 {
		return 1 // High priority for long keys
	}
	return 2 // Medium priority
}

func (s *ConditionalReloadStrategy) PreReload(ctx context.Context, key string, data interface{}) error {
	fmt.Printf("[%s] Pre-reload: %s\n", s.name, key)
	return nil
}

func (s *ConditionalReloadStrategy) PostReload(ctx context.Context, key string, data interface{}) error {
	fmt.Printf("[%s] Post-reload: %s\n", s.name, key)
	return nil
}

func (s *ConditionalReloadStrategy) OnReloadError(ctx context.Context, key string, err error) error {
	fmt.Printf("[%s] Reload error for %s: %v\n", s.name, key, err)
	return nil
}

func (s *ConditionalReloadStrategy) GetBatchSize() int {
	return s.batchSize
}

func (s *ConditionalReloadStrategy) GetRetryCount() int {
	return s.retryCount
}

func (s *ConditionalReloadStrategy) GetRetryDelay() time.Duration {
	return s.retryDelay
}
