package reload

import (
	"context"
	"fmt"
	"time"
)

// TimeBasedReloadStrategy reloads based on time intervals
type TimeBasedReloadStrategy struct {
	name       string
	interval   time.Duration
	batchSize  int
	retryCount int
	retryDelay time.Duration
}

// NewTimeBasedReloadStrategy creates a new time-based reload strategy
func NewTimeBasedReloadStrategy(name string, interval time.Duration, batchSize int) *TimeBasedReloadStrategy {
	return &TimeBasedReloadStrategy{
		name:       name,
		interval:   interval,
		batchSize:  batchSize,
		retryCount: 3,
		retryDelay: time.Second,
	}
}

func (s *TimeBasedReloadStrategy) GetName() string {
	return s.name
}

func (s *TimeBasedReloadStrategy) GetDescription() string {
	return fmt.Sprintf("Reloads data every %v", s.interval)
}

func (s *TimeBasedReloadStrategy) ShouldReload(ctx context.Context, key string, lastReload time.Time) bool {
	return time.Since(lastReload) >= s.interval
}

func (s *TimeBasedReloadStrategy) GetReloadPriority(ctx context.Context, key string) int {
	// This is a simplified implementation
	// In practice, you'd need to track last reload time per key
	return 2 // Medium priority
}

func (s *TimeBasedReloadStrategy) PreReload(ctx context.Context, key string, data interface{}) error {
	fmt.Printf("[%s] Pre-reload: %s\n", s.name, key)
	return nil
}

func (s *TimeBasedReloadStrategy) PostReload(ctx context.Context, key string, data interface{}) error {
	fmt.Printf("[%s] Post-reload: %s\n", s.name, key)
	return nil
}

func (s *TimeBasedReloadStrategy) OnReloadError(ctx context.Context, key string, err error) error {
	fmt.Printf("[%s] Reload error for %s: %v\n", s.name, key, err)
	return nil
}

func (s *TimeBasedReloadStrategy) GetBatchSize() int {
	return s.batchSize
}

func (s *TimeBasedReloadStrategy) GetRetryCount() int {
	return s.retryCount
}

func (s *TimeBasedReloadStrategy) GetRetryDelay() time.Duration {
	return s.retryDelay
}
