package persistence

import (
	"context"
	"fmt"
	"time"

	"admin-service/src/domain"
	"backend-core/logging"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// userEventRepository implements domain.UserEventRepository
type userEventRepository struct {
	db     *gorm.DB
	logger *logging.Logger
}

// NewUserEventRepository creates a new user event repository
func NewUserEventRepository(db *gorm.DB, logger *logging.Logger) domain.UserEventRepository {
	return &userEventRepository{
		db:     db,
		logger: logger,
	}
}

func (r *userEventRepository) Create(ctx context.Context, event *domain.UserEvent) error {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.EventTime.IsZero() {
		event.EventTime = time.Now()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	if err := r.db.WithContext(ctx).Create(event).Error; err != nil {
		r.logger.Error("Failed to create user event",
			logging.Error(err),
			logging.String("user_id", event.UserID.String()),
			logging.String("event_type", string(event.EventType)))
		return fmt.Errorf("failed to create user event: %w", err)
	}

	r.logger.Info("User event created successfully",
		logging.String("event_id", event.ID.String()),
		logging.String("user_id", event.UserID.String()),
		logging.String("event_type", string(event.EventType)),
		logging.String("service_name", event.ServiceName))

	return nil
}

func (r *userEventRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.UserEvent, error) {
	var event domain.UserEvent
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&event).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user event not found: %w", err)
		}
		r.logger.Error("Failed to get user event by ID",
			logging.Error(err),
			logging.String("event_id", id.String()))
		return nil, fmt.Errorf("failed to get user event: %w", err)
	}

	return &event, nil
}

func (r *userEventRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.UserEvent, int64, error) {
	var events []*domain.UserEvent
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&domain.UserEvent{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		r.logger.Error("Failed to count user events",
			logging.Error(err),
			logging.String("user_id", userID.String()))
		return nil, 0, fmt.Errorf("failed to count user events: %w", err)
	}

	// Get events
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("event_time DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error; err != nil {
		r.logger.Error("Failed to get user events",
			logging.Error(err),
			logging.String("user_id", userID.String()))
		return nil, 0, fmt.Errorf("failed to get user events: %w", err)
	}

	return events, total, nil
}

func (r *userEventRepository) GetByEventType(ctx context.Context, eventType domain.EventType, limit, offset int) ([]*domain.UserEvent, int64, error) {
	var events []*domain.UserEvent
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&domain.UserEvent{}).
		Where("event_type = ?", eventType).
		Count(&total).Error; err != nil {
		r.logger.Error("Failed to count events by type",
			logging.Error(err),
			logging.String("event_type", string(eventType)))
		return nil, 0, fmt.Errorf("failed to count events: %w", err)
	}

	// Get events
	if err := r.db.WithContext(ctx).
		Where("event_type = ?", eventType).
		Order("event_time DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error; err != nil {
		r.logger.Error("Failed to get events by type",
			logging.Error(err),
			logging.String("event_type", string(eventType)))
		return nil, 0, fmt.Errorf("failed to get events: %w", err)
	}

	return events, total, nil
}

func (r *userEventRepository) GetByUserIDAndEventType(ctx context.Context, userID uuid.UUID, eventType domain.EventType, limit, offset int) ([]*domain.UserEvent, int64, error) {
	var events []*domain.UserEvent
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&domain.UserEvent{}).
		Where("user_id = ? AND event_type = ?", userID, eventType).
		Count(&total).Error; err != nil {
		r.logger.Error("Failed to count user events by type",
			logging.Error(err),
			logging.String("user_id", userID.String()),
			logging.String("event_type", string(eventType)))
		return nil, 0, fmt.Errorf("failed to count user events: %w", err)
	}

	// Get events
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND event_type = ?", userID, eventType).
		Order("event_time DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error; err != nil {
		r.logger.Error("Failed to get user events by type",
			logging.Error(err),
			logging.String("user_id", userID.String()),
			logging.String("event_type", string(eventType)))
		return nil, 0, fmt.Errorf("failed to get user events: %w", err)
	}

	return events, total, nil
}

func (r *userEventRepository) GetByDateRange(ctx context.Context, from, to time.Time, limit, offset int) ([]*domain.UserEvent, int64, error) {
	var events []*domain.UserEvent
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&domain.UserEvent{}).
		Where("event_time >= ? AND event_time <= ?", from, to).
		Count(&total).Error; err != nil {
		r.logger.Error("Failed to count events by date range",
			logging.Error(err))
		return nil, 0, fmt.Errorf("failed to count events: %w", err)
	}

	// Get events
	if err := r.db.WithContext(ctx).
		Where("event_time >= ? AND event_time <= ?", from, to).
		Order("event_time DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error; err != nil {
		r.logger.Error("Failed to get events by date range",
			logging.Error(err))
		return nil, 0, fmt.Errorf("failed to get events: %w", err)
	}

	return events, total, nil
}

func (r *userEventRepository) GetByServiceName(ctx context.Context, serviceName string, limit, offset int) ([]*domain.UserEvent, int64, error) {
	var events []*domain.UserEvent
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&domain.UserEvent{}).
		Where("service_name = ?", serviceName).
		Count(&total).Error; err != nil {
		r.logger.Error("Failed to count events by service",
			logging.Error(err),
			logging.String("service_name", serviceName))
		return nil, 0, fmt.Errorf("failed to count events: %w", err)
	}

	// Get events
	if err := r.db.WithContext(ctx).
		Where("service_name = ?", serviceName).
		Order("event_time DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error; err != nil {
		r.logger.Error("Failed to get events by service",
			logging.Error(err),
			logging.String("service_name", serviceName))
		return nil, 0, fmt.Errorf("failed to get events: %w", err)
	}

	return events, total, nil
}

func (r *userEventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&domain.UserEvent{}, id).Error; err != nil {
		r.logger.Error("Failed to delete user event",
			logging.Error(err),
			logging.String("event_id", id.String()))
		return fmt.Errorf("failed to delete user event: %w", err)
	}

	r.logger.Info("User event deleted successfully",
		logging.String("event_id", id.String()))

	return nil
}

func (r *userEventRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.UserEvent{}).Count(&count).Error; err != nil {
		r.logger.Error("Failed to count user events", logging.Error(err))
		return 0, fmt.Errorf("failed to count user events: %w", err)
	}

	return count, nil
}

func (r *userEventRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.UserEvent{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		r.logger.Error("Failed to count user events",
			logging.Error(err),
			logging.String("user_id", userID.String()))
		return 0, fmt.Errorf("failed to count user events: %w", err)
	}

	return count, nil
}
