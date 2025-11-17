package services

import (
	"context"
	"time"

	"admin-service/src/domain"
	"backend-core/logging"

	"github.com/google/uuid"
)

// UserEventService handles business logic for user events
type UserEventService struct {
	repo   domain.UserEventRepository
	logger *logging.Logger
}

// NewUserEventService creates a new user event service
func NewUserEventService(repo domain.UserEventRepository, logger *logging.Logger) *UserEventService {
	return &UserEventService{
		repo:   repo,
		logger: logger,
	}
}

// RecordUserCreated records a user creation event
func (s *UserEventService) RecordUserCreated(ctx context.Context, userID uuid.UUID, email, username, firstName, lastName, serviceName, performedBy string) (*domain.UserEvent, error) {
	event := &domain.UserEvent{
		UserID:      userID,
		EventType:   domain.EventTypeUserCreated,
		Email:       email,
		Username:    username,
		FirstName:   firstName,
		LastName:    lastName,
		ServiceName: serviceName,
		PerformedBy: performedBy,
		EventTime:   time.Now(),
		Metadata: map[string]interface{}{
			"action": "create_user",
		},
	}

	if err := s.repo.Create(ctx, event); err != nil {
		s.logger.Error("Failed to record user created event",
			logging.Error(err),
			logging.String("user_id", userID.String()))
		return nil, err
	}

	s.logger.Info("User created event recorded",
		logging.String("event_id", event.ID.String()),
		logging.String("user_id", userID.String()),
		logging.String("email", email))

	return event, nil
}

// RecordUserUpdated records a user update event
func (s *UserEventService) RecordUserUpdated(ctx context.Context, userID uuid.UUID, email, username, firstName, lastName, serviceName, performedBy string, changedFields []string) (*domain.UserEvent, error) {
	metadata := map[string]interface{}{
		"action": "update_user",
	}
	if len(changedFields) > 0 {
		metadata["changed_fields"] = changedFields
	}

	event := &domain.UserEvent{
		UserID:      userID,
		EventType:   domain.EventTypeUserUpdated,
		Email:       email,
		Username:    username,
		FirstName:   firstName,
		LastName:    lastName,
		ServiceName: serviceName,
		PerformedBy: performedBy,
		EventTime:   time.Now(),
		Metadata:    metadata,
	}

	if err := s.repo.Create(ctx, event); err != nil {
		s.logger.Error("Failed to record user updated event",
			logging.Error(err),
			logging.String("user_id", userID.String()))
		return nil, err
	}

	s.logger.Info("User updated event recorded",
		logging.String("event_id", event.ID.String()),
		logging.String("user_id", userID.String()))

	return event, nil
}

// RecordUserDeleted records a user deletion event
func (s *UserEventService) RecordUserDeleted(ctx context.Context, userID uuid.UUID, email, serviceName, performedBy, reason string) (*domain.UserEvent, error) {
	metadata := map[string]interface{}{
		"action": "delete_user",
	}
	if reason != "" {
		metadata["reason"] = reason
	}

	event := &domain.UserEvent{
		UserID:      userID,
		EventType:   domain.EventTypeUserDeleted,
		Email:       email,
		ServiceName: serviceName,
		PerformedBy: performedBy,
		EventTime:   time.Now(),
		Metadata:    metadata,
	}

	if err := s.repo.Create(ctx, event); err != nil {
		s.logger.Error("Failed to record user deleted event",
			logging.Error(err),
			logging.String("user_id", userID.String()))
		return nil, err
	}

	s.logger.Info("User deleted event recorded",
		logging.String("event_id", event.ID.String()),
		logging.String("user_id", userID.String()))

	return event, nil
}

// GetUserEvents retrieves user events with pagination and filtering
func (s *UserEventService) GetUserEvents(ctx context.Context, userID uuid.UUID, eventType domain.EventType, fromDate, toDate time.Time, page, pageSize int) ([]*domain.UserEvent, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	var events []*domain.UserEvent
	var total int64
	var err error

	// Filter by user ID and event type if provided
	if userID != uuid.Nil && eventType != "" {
		events, total, err = s.repo.GetByUserIDAndEventType(ctx, userID, eventType, pageSize, offset)
	} else if userID != uuid.Nil {
		events, total, err = s.repo.GetByUserID(ctx, userID, pageSize, offset)
	} else if eventType != "" {
		events, total, err = s.repo.GetByEventType(ctx, eventType, pageSize, offset)
	} else if !fromDate.IsZero() && !toDate.IsZero() {
		events, total, err = s.repo.GetByDateRange(ctx, fromDate, toDate, pageSize, offset)
	} else {
		// Get all events with pagination
		events, total, err = s.repo.GetByDateRange(ctx, time.Time{}, time.Now(), pageSize, offset)
	}

	if err != nil {
		s.logger.Error("Failed to get user events",
			logging.Error(err))
		return nil, 0, err
	}

	return events, total, nil
}

// GetEventByID retrieves a specific event by ID
func (s *UserEventService) GetEventByID(ctx context.Context, eventID uuid.UUID) (*domain.UserEvent, error) {
	event, err := s.repo.GetByID(ctx, eventID)
	if err != nil {
		s.logger.Error("Failed to get event by ID",
			logging.Error(err),
			logging.String("event_id", eventID.String()))
		return nil, err
	}

	return event, nil
}
