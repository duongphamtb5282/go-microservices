package services

import (
	"context"

	"auth-service/src/domain/entities"
	"auth-service/src/domain/errors"
	"auth-service/src/domain/repositories"
	"auth-service/src/domain/valueObjects"
)

// UserDomainService handles complex domain logic for users
type UserDomainService struct {
	userRepo repositories.UserRepository
}

// NewUserDomainService creates a new UserDomainService
func NewUserDomainService(userRepo repositories.UserRepository) *UserDomainService {
	return &UserDomainService{
		userRepo: userRepo,
	}
}

// CreateUser creates a new user with business rule validation
func (s *UserDomainService) CreateUser(ctx context.Context, username, email, password, createdBy string) (*entities.User, error) {
	// Check if user already exists
	emailVO, err := valueObjects.NewEmail(email)
	if err != nil {
		return nil, err
	}

	exists, err := s.userRepo.ExistsByEmail(ctx, emailVO)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.ErrEmailAlreadyExists
	}

	usernameVO, err := valueObjects.NewUsername(username)
	if err != nil {
		return nil, err
	}

	exists, err = s.userRepo.ExistsByUsername(ctx, usernameVO)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.ErrUsernameAlreadyExists
	}

	// Create user entity
	user, err := entities.NewUser(username, email, password, createdBy)
	if err != nil {
		return nil, err
	}

	// Save user
	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// AuthenticateUser authenticates a user with email and password
func (s *UserDomainService) AuthenticateUser(ctx context.Context, email, password string) (*entities.User, error) {
	emailVO, err := valueObjects.NewEmail(email)
	if err != nil {
		return nil, err
	}

	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, emailVO)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	// Check if user is active
	if !user.IsActive() {
		return nil, errors.ErrUserNotFound
	}

	// Check if user is locked
	if user.IsLocked() {
		return nil, errors.ErrUserLocked
	}

	// Verify password
	if !user.Password().Verify(password) {
		// Record failed login attempt
		user.RecordFailedLogin()
		s.userRepo.UpdateLoginAttempts(ctx, user.ID(), user.LoginAttempts())
		return nil, errors.ErrInvalidCredentials
	}

	// Record successful login
	user.RecordLogin()
	s.userRepo.UpdateLastLogin(ctx, user.ID())

	return user, nil
}

// ActivateUser activates a user account
func (s *UserDomainService) ActivateUser(ctx context.Context, userID valueObjects.UserID) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := user.Activate(); err != nil {
		return err
	}

	return s.userRepo.Save(ctx, user)
}

// DeactivateUser deactivates a user account
func (s *UserDomainService) DeactivateUser(ctx context.Context, userID valueObjects.UserID) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := user.Deactivate(); err != nil {
		return err
	}

	return s.userRepo.Save(ctx, user)
}

// ChangePassword changes a user's password
func (s *UserDomainService) ChangePassword(ctx context.Context, userID valueObjects.UserID, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := user.ChangePassword(newPassword); err != nil {
		return err
	}

	return s.userRepo.Save(ctx, user)
}
