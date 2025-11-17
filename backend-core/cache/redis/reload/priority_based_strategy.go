package reload

import (
	"context"
	"fmt"
	"time"
)

// PriorityBasedReloadStrategy reloads based on key priority
type PriorityBasedReloadStrategy struct {
	name       string
	priorities map[string]int
	batchSize  int
	retryCount int
	retryDelay time.Duration
}

// NewPriorityBasedReloadStrategy creates a new priority-based reload strategy
func NewPriorityBasedReloadStrategy(name string, priorities map[string]int, batchSize int) *PriorityBasedReloadStrategy {
	return &PriorityBasedReloadStrategy{
		name:       name,
		priorities: priorities,
		batchSize:  batchSize,
		retryCount: 3,
		retryDelay: time.Second,
	}
}

func (s *PriorityBasedReloadStrategy) GetName() string {
	return s.name
}

func (s *PriorityBasedReloadStrategy) GetDescription() string {
	return "Reloads data based on key priority"
}

func (s *PriorityBasedReloadStrategy) ShouldReload(ctx context.Context, key string, lastReload time.Time) bool {
	// Always reload if key has high priority
	if priority, exists := s.priorities[key]; exists {
		return priority <= 2 // High priority
	}
	return false
}

func (s *PriorityBasedReloadStrategy) GetReloadPriority(ctx context.Context, key string) int {
	if priority, exists := s.priorities[key]; exists {
		return priority
	}
	return 999 // Low priority for unknown keys
}

func (s *PriorityBasedReloadStrategy) PreReload(ctx context.Context, key string, data interface{}) error {
	priority := s.GetReloadPriority(ctx, key)
	fmt.Printf("[%s] Pre-reload: %s (priority: %d)\n", s.name, key, priority)
	return nil
}

func (s *PriorityBasedReloadStrategy) PostReload(ctx context.Context, key string, data interface{}) error {
	fmt.Printf("[%s] Post-reload: %s\n", s.name, key)
	return nil
}

func (s *PriorityBasedReloadStrategy) OnReloadError(ctx context.Context, key string, err error) error {
	fmt.Printf("[%s] Reload error for %s: %v\n", s.name, key, err)
	return nil
}

func (s *PriorityBasedReloadStrategy) GetBatchSize() int {
	return s.batchSize
}

func (s *PriorityBasedReloadStrategy) GetRetryCount() int {
	return s.retryCount
}

func (s *PriorityBasedReloadStrategy) GetRetryDelay() time.Duration {
	return s.retryDelay
}
