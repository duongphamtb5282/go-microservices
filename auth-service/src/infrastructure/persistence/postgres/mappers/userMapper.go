package mappers

import (
	"fmt"

	"auth-service/src/domain/entities"
	"auth-service/src/domain/valueObjects"
	"auth-service/src/infrastructure/persistence/postgres/models"
)

// UserMapper handles transformation between domain entities and database models
type UserMapper struct{}

// NewUserMapper creates a new UserMapper
func NewUserMapper() *UserMapper {
	return &UserMapper{}
}

// ToModel converts a domain entity to a database model
func (m *UserMapper) ToModel(user *entities.User) *models.User {
	return &models.User{
		CreatedBy:     user.AuditInfo().CreatedBy,
		CreatedAt:     user.AuditInfo().CreatedAt,
		ModifiedBy:    user.AuditInfo().CreatedBy, // For now, set to same as created
		ModifiedAt:    user.AuditInfo().CreatedAt,
		ID:            user.ID().String(),
		Username:      user.Username().String(),
		Email:         user.Email().String(),
		PasswordHash:  user.Password().Hash(),
		IsActive:      user.IsActive(),
		LastLoginAt:   user.LastLoginAt(),
		LoginAttempts: user.LoginAttempts(),
	}
}

// ToEntity converts a database model to a domain entity
func (m *UserMapper) ToEntity(model *models.User) (*entities.User, error) {
	// Create value objects
	userID, err := valueObjects.NewUserIDFromString(model.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID '%s': %w", model.ID, err)
	}

	username, err := valueObjects.NewUsername(model.Username)
	if err != nil {
		return nil, fmt.Errorf("invalid username '%s': %w", model.Username, err)
	}

	email, err := valueObjects.NewEmail(model.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email '%s': %w", model.Email, err)
	}

	// Create password from bcrypt hash (no salt needed - bcrypt includes salt in hash)
	password, err := valueObjects.NewPasswordFromHash(model.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("invalid password hash (length=%d): %w", len(model.PasswordHash), err)
	}

	// Create audit info
	auditInfo := entities.AuditInfo{
		CreatedBy: model.CreatedBy,
		CreatedAt: model.CreatedAt,
	}

	// Create user entity using the repository constructor
	user := entities.NewUserFromRepository(
		userID,
		username,
		email,
		password,
		model.IsActive,
		model.LastLoginAt,
		model.LoginAttempts,
		auditInfo,
	)

	return user, nil
}

// ToEntitySlice converts a slice of database models to domain entities
func (m *UserMapper) ToEntitySlice(models []*models.User) ([]*entities.User, error) {
	entities := make([]*entities.User, len(models))
	for i, model := range models {
		entity, err := m.ToEntity(model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert model at index %d: %w", i, err)
		}
		entities[i] = entity
	}
	return entities, nil
}
