package postgres

import (
	"context"
	"fmt"
	"time"

	"auth-service/src/domain/entities"
	"auth-service/src/domain/repositories"
	"auth-service/src/domain/valueObjects"
	"auth-service/src/infrastructure/persistence/postgres/mappers"
	"auth-service/src/infrastructure/persistence/postgres/models"
	"backend-core/database/gorm"
	"backend-core/logging"
)

// PostgresUserRepository implements the UserRepository interface using backend-core GORM repository
type PostgresUserRepository struct {
	*gorm.GormRepository[models.User]
	mapper *mappers.UserMapper
	logger *logging.Logger
}

// NewPostgresUserRepository creates a new PostgresUserRepository using backend-core base repository
func NewPostgresUserRepository(db gorm.Database, logger *logging.Logger) repositories.UserRepository {
	// Create base repository using backend-core
	baseRepo := gorm.NewGormRepository[models.User](db, "user", logger)

	return &PostgresUserRepository{
		GormRepository: baseRepo,
		mapper:         mappers.NewUserMapper(),
		logger:         logger,
	}
}

// Save saves a user entity using base repository
func (r *PostgresUserRepository) Save(ctx context.Context, user *entities.User) error {
	model := r.mapper.ToModel(user)

	// Use base repository Update method (which handles upsert)
	return r.GormRepository.Update(ctx, model)
}

// FindByID finds a user by ID using base repository
func (r *PostgresUserRepository) FindByID(ctx context.Context, id valueObjects.UserID) (*entities.User, error) {
	model, err := r.GormRepository.GetByID(ctx, id.String())
	if err != nil {
		r.logger.Error("Failed to find user by ID", logging.Error(err))
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if model == nil {
		return nil, fmt.Errorf("user not found")
	}

	user, err := r.mapper.ToEntity(model)
	if err != nil {
		r.logger.Error("Failed to map user model to entity", logging.Error(err), logging.String("user_id", id.String()))
		return nil, fmt.Errorf("failed to map user: %w", err)
	}

	return user, nil
}

// FindByEmail finds a user by email using backend-core query builder
func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email valueObjects.Email) (*entities.User, error) {
	// Use backend-core query builder
	query := gorm.Query{
		Filter: map[string]interface{}{
			"email": email.String(),
		},
	}

	models, err := r.GormRepository.Find(ctx, query)
	if err != nil {
		r.logger.Error("Failed to find user by email", logging.Error(err))
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if len(models) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return r.mapper.ToEntity(models[0])
}

// FindByUsername finds a user by username using backend-core query builder
func (r *PostgresUserRepository) FindByUsername(ctx context.Context, username valueObjects.Username) (*entities.User, error) {
	// Use backend-core query builder
	query := gorm.Query{
		Filter: map[string]interface{}{
			"username": username.String(),
		},
	}

	models, err := r.GormRepository.Find(ctx, query)
	if err != nil {
		r.logger.Error("Failed to find user by username", logging.Error(err))
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if len(models) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return r.mapper.ToEntity(models[0])
}

// FindAll finds all users with pagination using backend-core
func (r *PostgresUserRepository) FindAll(ctx context.Context, offset, limit int) ([]*entities.User, error) {
	// Use backend-core GetAll method with pagination
	pagination := gorm.Pagination{
		Page:     offset/limit + 1,
		PageSize: limit,
	}

	models, err := r.GormRepository.GetAll(ctx, map[string]interface{}{}, pagination)
	if err != nil {
		r.logger.Error("Failed to find all users", logging.Error(err))
		return nil, fmt.Errorf("failed to find users: %w", err)
	}

	// Convert models to entities
	entities := make([]*entities.User, len(models))
	for i, model := range models {
		entity, err := r.mapper.ToEntity(model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to entity: %w", err)
		}
		entities[i] = entity
	}

	return entities, nil
}

// Count returns the total number of users using base repository
func (r *PostgresUserRepository) Count(ctx context.Context) (int64, error) {
	return r.GormRepository.Count(ctx, map[string]interface{}{})
}

// Delete deletes a user by ID using base repository
func (r *PostgresUserRepository) Delete(ctx context.Context, id valueObjects.UserID) error {
	return r.GormRepository.Delete(ctx, id.String())
}

// ExistsByEmail checks if a user exists with the given email using backend-core
func (r *PostgresUserRepository) ExistsByEmail(ctx context.Context, email valueObjects.Email) (bool, error) {
	filter := map[string]interface{}{
		"email": email.String(),
	}

	return r.GormRepository.Exists(ctx, filter)
}

// ExistsByUsername checks if a user exists with the given username using backend-core
func (r *PostgresUserRepository) ExistsByUsername(ctx context.Context, username valueObjects.Username) (bool, error) {
	filter := map[string]interface{}{
		"username": username.String(),
	}

	return r.GormRepository.Exists(ctx, filter)
}

// UpdateLastLogin updates the last login time for a user using backend-core
func (r *PostgresUserRepository) UpdateLastLogin(ctx context.Context, id valueObjects.UserID) error {
	// Get the user first
	user, err := r.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Update the last login time in the entity
	// Note: This would require adding a method to update last login in the User entity
	// For now, we'll use a direct update approach
	model := r.mapper.ToModel(user)
	model.LastLoginAt = &time.Time{}
	*model.LastLoginAt = time.Now()

	// Use backend-core Update method
	return r.GormRepository.Update(ctx, model)
}

// UpdateLoginAttempts updates the login attempts for a user using backend-core
func (r *PostgresUserRepository) UpdateLoginAttempts(ctx context.Context, id valueObjects.UserID, attempts int) error {
	// Get the user first
	user, err := r.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Update the login attempts in the entity
	// Note: This would require adding a method to update login attempts in the User entity
	// For now, we'll use a direct update approach
	model := r.mapper.ToModel(user)
	model.LoginAttempts = attempts

	// Use backend-core Update method
	return r.GormRepository.Update(ctx, model)
}
