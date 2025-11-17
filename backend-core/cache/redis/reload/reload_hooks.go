package reload

import (
	"context"
	"fmt"
	"time"
)

// ReloadHook defines a hook that can be executed during cache reloading
type ReloadHook interface {
	// BeforeReload is called before a reload operation starts
	BeforeReload(ctx context.Context, key string, data interface{}) error

	// AfterReload is called after a reload operation completes successfully
	AfterReload(ctx context.Context, key string, data interface{}) error

	// OnReloadError is called when a reload operation fails
	OnReloadError(ctx context.Context, key string, err error) error

	// GetPriority returns the priority of this hook (lower numbers = higher priority)
	GetPriority() int

	// GetName returns the name of this hook
	GetName() string
}

// ReloadHookManager manages reload hooks
type ReloadHookManager struct {
	hooks []ReloadHook
}

// NewReloadHookManager creates a new reload hook manager
func NewReloadHookManager() *ReloadHookManager {
	return &ReloadHookManager{
		hooks: make([]ReloadHook, 0),
	}
}

// AddHook adds a reload hook
func (m *ReloadHookManager) AddHook(hook ReloadHook) {
	m.hooks = append(m.hooks, hook)
	// Sort hooks by priority
	m.sortHooks()
}

// RemoveHook removes a reload hook by name
func (m *ReloadHookManager) RemoveHook(name string) {
	for i, hook := range m.hooks {
		if hook.GetName() == name {
			m.hooks = append(m.hooks[:i], m.hooks[i+1:]...)
			break
		}
	}
}

// ExecuteBeforeReload executes all before reload hooks
func (m *ReloadHookManager) ExecuteBeforeReload(ctx context.Context, key string, data interface{}) error {
	for _, hook := range m.hooks {
		if err := hook.BeforeReload(ctx, key, data); err != nil {
			return fmt.Errorf("hook %s failed in BeforeReload: %w", hook.GetName(), err)
		}
	}
	return nil
}

// ExecuteAfterReload executes all after reload hooks
func (m *ReloadHookManager) ExecuteAfterReload(ctx context.Context, key string, data interface{}) error {
	for _, hook := range m.hooks {
		if err := hook.AfterReload(ctx, key, data); err != nil {
			return fmt.Errorf("hook %s failed in AfterReload: %w", hook.GetName(), err)
		}
	}
	return nil
}

// ExecuteOnReloadError executes all error hooks
func (m *ReloadHookManager) ExecuteOnReloadError(ctx context.Context, key string, err error) error {
	for _, hook := range m.hooks {
		if hookErr := hook.OnReloadError(ctx, key, err); hookErr != nil {
			return fmt.Errorf("hook %s failed in OnReloadError: %w", hook.GetName(), hookErr)
		}
	}
	return nil
}

// sortHooks sorts hooks by priority (lower number = higher priority)
func (m *ReloadHookManager) sortHooks() {
	for i := 0; i < len(m.hooks)-1; i++ {
		for j := i + 1; j < len(m.hooks); j++ {
			if m.hooks[i].GetPriority() > m.hooks[j].GetPriority() {
				m.hooks[i], m.hooks[j] = m.hooks[j], m.hooks[i]
			}
		}
	}
}

// GetHooks returns all registered hooks
func (m *ReloadHookManager) GetHooks() []ReloadHook {
	return m.hooks
}

// ClearHooks removes all hooks
func (m *ReloadHookManager) ClearHooks() {
	m.hooks = make([]ReloadHook, 0)
}

// ============================================================================
// Built-in Reload Hooks
// ============================================================================

// LoggingHook logs reload operations
type LoggingHook struct {
	name     string
	priority int
}

// NewLoggingHook creates a new logging hook
func NewLoggingHook(name string, priority int) *LoggingHook {
	return &LoggingHook{
		name:     name,
		priority: priority,
	}
}

func (h *LoggingHook) BeforeReload(ctx context.Context, key string, data interface{}) error {
	fmt.Printf("[%s] Starting reload for key: %s\n", h.name, key)
	return nil
}

func (h *LoggingHook) AfterReload(ctx context.Context, key string, data interface{}) error {
	fmt.Printf("[%s] Completed reload for key: %s\n", h.name, key)
	return nil
}

func (h *LoggingHook) OnReloadError(ctx context.Context, key string, err error) error {
	fmt.Printf("[%s] Reload failed for key: %s, error: %v\n", h.name, key, err)
	return nil
}

func (h *LoggingHook) GetPriority() int {
	return h.priority
}

func (h *LoggingHook) GetName() string {
	return h.name
}

// MetricsHook tracks reload metrics
type MetricsHook struct {
	name     string
	priority int
	metrics  *ReloadMetrics
}

// NewMetricsHook creates a new metrics hook
func NewMetricsHook(name string, priority int, metrics *ReloadMetrics) *MetricsHook {
	return &MetricsHook{
		name:     name,
		priority: priority,
		metrics:  metrics,
	}
}

func (h *MetricsHook) BeforeReload(ctx context.Context, key string, data interface{}) error {
	// Track start time
	ctx = context.WithValue(ctx, "reload_start_time", time.Now())
	return nil
}

func (h *MetricsHook) AfterReload(ctx context.Context, key string, data interface{}) error {
	// Track successful reload
	h.metrics.SuccessfulReloads++
	h.metrics.TotalReloads++

	// Calculate duration
	if startTime, ok := ctx.Value("reload_start_time").(time.Time); ok {
		duration := time.Since(startTime)
		h.metrics.LastReloadTime = time.Now()

		// Update average reload time
		if h.metrics.TotalReloads > 0 {
			totalTime := h.metrics.AverageReloadTime * time.Duration(h.metrics.TotalReloads-1)
			h.metrics.AverageReloadTime = (totalTime + duration) / time.Duration(h.metrics.TotalReloads)
		}
	}

	return nil
}

func (h *MetricsHook) OnReloadError(ctx context.Context, key string, err error) error {
	// Track failed reload
	h.metrics.FailedReloads++
	h.metrics.TotalReloads++
	h.metrics.LastReloadTime = time.Now()
	return nil
}

func (h *MetricsHook) GetPriority() int {
	return h.priority
}

func (h *MetricsHook) GetName() string {
	return h.name
}

// ValidationHook validates data before and after reload
type ValidationHook struct {
	name      string
	priority  int
	validator func(interface{}) error
}

// NewValidationHook creates a new validation hook
func NewValidationHook(name string, priority int, validator func(interface{}) error) *ValidationHook {
	return &ValidationHook{
		name:      name,
		priority:  priority,
		validator: validator,
	}
}

func (h *ValidationHook) BeforeReload(ctx context.Context, key string, data interface{}) error {
	if h.validator != nil {
		return h.validator(data)
	}
	return nil
}

func (h *ValidationHook) AfterReload(ctx context.Context, key string, data interface{}) error {
	// Additional validation after reload if needed
	return nil
}

func (h *ValidationHook) OnReloadError(ctx context.Context, key string, err error) error {
	// Handle validation errors
	return nil
}

func (h *ValidationHook) GetPriority() int {
	return h.priority
}

func (h *ValidationHook) GetName() string {
	return h.name
}

// TransformHook transforms data during reload
type TransformHook struct {
	name        string
	priority    int
	transformer func(interface{}) (interface{}, error)
}

// NewTransformHook creates a new transform hook
func NewTransformHook(name string, priority int, transformer func(interface{}) (interface{}, error)) *TransformHook {
	return &TransformHook{
		name:        name,
		priority:    priority,
		transformer: transformer,
	}
}

func (h *TransformHook) BeforeReload(ctx context.Context, key string, data interface{}) error {
	if h.transformer != nil {
		transformed, err := h.transformer(data)
		if err != nil {
			return err
		}
		// Store transformed data in context for use in AfterReload
		ctx = context.WithValue(ctx, "transformed_data", transformed)
	}
	return nil
}

func (h *TransformHook) AfterReload(ctx context.Context, key string, data interface{}) error {
	// Clean up context if needed
	return nil
}

func (h *TransformHook) OnReloadError(ctx context.Context, key string, err error) error {
	// Handle transformation errors
	return nil
}

func (h *TransformHook) GetPriority() int {
	return h.priority
}

func (h *TransformHook) GetName() string {
	return h.name
}

// NotificationHook sends notifications about reload operations
type NotificationHook struct {
	name     string
	priority int
	notifier func(string, string, interface{}) error
}

// NewNotificationHook creates a new notification hook
func NewNotificationHook(name string, priority int, notifier func(string, string, interface{}) error) *NotificationHook {
	return &NotificationHook{
		name:     name,
		priority: priority,
		notifier: notifier,
	}
}

func (h *NotificationHook) BeforeReload(ctx context.Context, key string, data interface{}) error {
	if h.notifier != nil {
		return h.notifier("reload_started", key, data)
	}
	return nil
}

func (h *NotificationHook) AfterReload(ctx context.Context, key string, data interface{}) error {
	if h.notifier != nil {
		return h.notifier("reload_completed", key, data)
	}
	return nil
}

func (h *NotificationHook) OnReloadError(ctx context.Context, key string, err error) error {
	if h.notifier != nil {
		return h.notifier("reload_failed", key, err)
	}
	return nil
}

func (h *NotificationHook) GetPriority() int {
	return h.priority
}

func (h *NotificationHook) GetName() string {
	return h.name
}

// ConditionalHook executes custom logic based on conditions
type ConditionalHook struct {
	name      string
	priority  int
	condition func(string, interface{}) bool
	action    func(context.Context, string, interface{}) error
}

// NewConditionalHook creates a new conditional hook
func NewConditionalHook(name string, priority int, condition func(string, interface{}) bool, action func(context.Context, string, interface{}) error) *ConditionalHook {
	return &ConditionalHook{
		name:      name,
		priority:  priority,
		condition: condition,
		action:    action,
	}
}

func (h *ConditionalHook) BeforeReload(ctx context.Context, key string, data interface{}) error {
	if h.condition != nil && h.condition(key, data) {
		if h.action != nil {
			return h.action(ctx, key, data)
		}
	}
	return nil
}

func (h *ConditionalHook) AfterReload(ctx context.Context, key string, data interface{}) error {
	// Execute action after reload if condition is met
	if h.condition != nil && h.condition(key, data) {
		if h.action != nil {
			return h.action(ctx, key, data)
		}
	}
	return nil
}

func (h *ConditionalHook) OnReloadError(ctx context.Context, key string, err error) error {
	// Execute action on error if condition is met
	if h.condition != nil && h.condition(key, err) {
		if h.action != nil {
			return h.action(ctx, key, err)
		}
	}
	return nil
}

func (h *ConditionalHook) GetPriority() int {
	return h.priority
}

func (h *ConditionalHook) GetName() string {
	return h.name
}
