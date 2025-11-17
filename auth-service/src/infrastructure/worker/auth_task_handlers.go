package worker

import (
	"context"
	"fmt"

	"auth-service/src/domain/entities"
	"auth-service/src/domain/repositories"
	"backend-core/logging"
)

// Task types for auth-service
const (
	TaskTypeCacheUser        = "cache_user"
	TaskTypeCacheUserByEmail = "cache_user_by_email"
	TaskTypePublishEvent     = "publish_event"
	TaskTypeRecordAdminUser  = "record_admin_user"
)

// CacheUserPayload represents the payload for caching a user
type CacheUserPayload struct {
	UserID   string         `json:"user_id"`
	UserData *entities.User `json:"user_data"`
}

// CacheUserByEmailPayload represents the payload for caching a user by email
type CacheUserByEmailPayload struct {
	Email    string         `json:"email"`
	UserData *entities.User `json:"user_data"`
}

// PublishEventPayload represents the payload for publishing an event
type PublishEventPayload struct {
	EventType string      `json:"event_type"`
	EventData interface{} `json:"event_data"`
}

// RecordAdminUserPayload represents the payload for recording user in admin service
type RecordAdminUserPayload struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	CreatedBy string `json:"created_by"`
}

// AuthTaskHandler handles auth-service specific tasks
type AuthTaskHandler struct {
	userCache   repositories.UserCache
	eventBus    EventBus
	adminClient AdminClient
	logger      *logging.Logger
}

// EventBus defines the interface for publishing events
type EventBus interface {
	Publish(event interface{}) error
}

// AdminClient defines the interface for admin service communication
type AdminClient interface {
	RecordUserCreated(ctx context.Context, userID, email, username, firstName, lastName, createdBy string) error
}

// NewAuthTaskHandler creates a new auth task handler
func NewAuthTaskHandler(userCache repositories.UserCache, eventBus EventBus, adminClient AdminClient, logger *logging.Logger) *AuthTaskHandler {
	return &AuthTaskHandler{
		userCache:   userCache,
		eventBus:    eventBus,
		adminClient: adminClient,
		logger:      logger,
	}
}

// GetTaskType returns the task type this handler processes
func (h *AuthTaskHandler) GetTaskType() string {
	return "auth_handler" // This handler handles multiple task types
}

// HandleTask processes a task
func (h *AuthTaskHandler) HandleTask(ctx context.Context, task *Task) error {
	switch task.Type {
	case TaskTypeCacheUser:
		return h.handleCacheUser(ctx, task)
	case TaskTypeCacheUserByEmail:
		return h.handleCacheUserByEmail(ctx, task)
	case TaskTypePublishEvent:
		return h.handlePublishEvent(ctx, task)
	case TaskTypeRecordAdminUser:
		return h.handleRecordAdminUser(ctx, task)
	default:
		return fmt.Errorf("unknown task type: %s", task.Type)
	}
}

// handleCacheUser handles caching a user
func (h *AuthTaskHandler) handleCacheUser(ctx context.Context, task *Task) error {
	payload, ok := task.Payload.(*CacheUserPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for cache_user task")
	}

	h.logger.Info("Processing cache user task", "task_id", task.ID, "user_id", payload.UserID)

	if err := h.userCache.CacheUser(ctx, payload.UserData); err != nil {
		h.logger.Error("Failed to cache user", "error", err, "task_id", task.ID, "user_id", payload.UserID)
		return fmt.Errorf("failed to cache user: %w", err)
	}

	h.logger.Info("Successfully cached user", "task_id", task.ID, "user_id", payload.UserID)

	return nil
}

// handleCacheUserByEmail handles caching a user by email
func (h *AuthTaskHandler) handleCacheUserByEmail(ctx context.Context, task *Task) error {
	payload, ok := task.Payload.(*CacheUserByEmailPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for cache_user_by_email task")
	}

	h.logger.Info("Processing cache user by email task", "task_id", task.ID, "email", payload.Email)

	if err := h.userCache.CacheUserByEmail(ctx, payload.UserData.Email(), payload.UserData); err != nil {
		h.logger.Error("Failed to cache user by email", "error", err, "task_id", task.ID, "email", payload.Email)
		return fmt.Errorf("failed to cache user by email: %w", err)
	}

	h.logger.Info("Successfully cached user by email", "task_id", task.ID, "email", payload.Email)

	return nil
}

// handlePublishEvent handles publishing an event
func (h *AuthTaskHandler) handlePublishEvent(ctx context.Context, task *Task) error {
	payload, ok := task.Payload.(*PublishEventPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for publish_event task")
	}

	h.logger.Info("Processing publish event task", "task_id", task.ID, "event_type", payload.EventType)

	if err := h.eventBus.Publish(payload.EventData); err != nil {
		h.logger.Error("Failed to publish event", "error", err, "task_id", task.ID, "event_type", payload.EventType)
		return fmt.Errorf("failed to publish event: %w", err)
	}

	h.logger.Info("Successfully published event", "task_id", task.ID, "event_type", payload.EventType)

	return nil
}

// handleRecordAdminUser handles recording user in admin service
func (h *AuthTaskHandler) handleRecordAdminUser(ctx context.Context, task *Task) error {
	if h.adminClient == nil {
		h.logger.Warn("Admin client not available, skipping admin user recording",
			logging.String("task_id", task.ID))
		return nil // Not an error, just not configured
	}

	payload, ok := task.Payload.(*RecordAdminUserPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for record_admin_user task")
	}

	h.logger.Info("Processing record admin user task", "task_id", task.ID, "user_id", payload.UserID, "email", payload.Email)

	if err := h.adminClient.RecordUserCreated(
		ctx,
		payload.UserID,
		payload.Email,
		payload.Username,
		payload.FirstName,
		payload.LastName,
		payload.CreatedBy,
	); err != nil {
		h.logger.Error("Failed to record user in admin service", "error", err, "task_id", task.ID, "user_id", payload.UserID)
		return fmt.Errorf("failed to record user in admin service: %w", err)
	}

	h.logger.Info("Successfully recorded user in admin service", "task_id", task.ID, "user_id", payload.UserID)

	return nil
}
