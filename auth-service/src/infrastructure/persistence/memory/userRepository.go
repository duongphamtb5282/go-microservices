package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"auth-service/src/domain/entities"
	domainErrors "auth-service/src/domain/errors"
	"auth-service/src/domain/repositories"
	"auth-service/src/domain/valueObjects"
	"backend-core/logging"
)

// MemoryUserRepository implements the UserRepository interface using in-memory storage
type MemoryUserRepository struct {
	users         map[string]*entities.User
	emailIndex    map[string]string
	usernameIndex map[string]string
	mutex         sync.RWMutex
	logger        *logging.Logger
}

// NewMemoryUserRepository creates a new MemoryUserRepository
func NewMemoryUserRepository(logger *logging.Logger) repositories.UserRepository {
	return &MemoryUserRepository{
		users:         make(map[string]*entities.User),
		emailIndex:    make(map[string]string),
		usernameIndex: make(map[string]string),
		logger:        logger,
	}
}

// Save saves a user entity
func (r *MemoryUserRepository) Save(ctx context.Context, user *entities.User) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	id := user.ID().String()

	if existing, ok := r.users[id]; ok {
		r.removeFromIndexes(existing)
	}

	stored := r.rehydrateUser(user, user.LastLoginAt(), user.LoginAttempts())
	r.users[id] = stored
	r.addToIndexes(stored)

	r.logInfo("User saved to memory", logging.String("user_id", id))
	return nil
}

// FindByID finds a user by ID
func (r *MemoryUserRepository) FindByID(ctx context.Context, id valueObjects.UserID) (*entities.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mutex.RLock()
	user, exists := r.users[id.String()]
	r.mutex.RUnlock()
	if !exists {
		return nil, domainErrors.ErrUserNotFound
	}

	return r.rehydrateUser(user, user.LastLoginAt(), user.LoginAttempts()), nil
}

// FindByEmail finds a user by email
func (r *MemoryUserRepository) FindByEmail(ctx context.Context, email valueObjects.Email) (*entities.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mutex.RLock()
	userID, exists := r.emailIndex[email.String()]
	if !exists {
		r.mutex.RUnlock()
		return nil, domainErrors.ErrUserNotFound
	}
	user := r.users[userID]
	r.mutex.RUnlock()

	return r.rehydrateUser(user, user.LastLoginAt(), user.LoginAttempts()), nil
}

// FindByUsername finds a user by username
func (r *MemoryUserRepository) FindByUsername(ctx context.Context, username valueObjects.Username) (*entities.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mutex.RLock()
	userID, exists := r.usernameIndex[username.String()]
	if !exists {
		r.mutex.RUnlock()
		return nil, domainErrors.ErrUserNotFound
	}
	user := r.users[userID]
	r.mutex.RUnlock()

	return r.rehydrateUser(user, user.LastLoginAt(), user.LoginAttempts()), nil
}

// FindAll finds all users with pagination
func (r *MemoryUserRepository) FindAll(ctx context.Context, offset, limit int) ([]*entities.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if limit <= 0 {
		return []*entities.User{}, nil
	}

	if offset < 0 {
		offset = 0
	}

	r.mutex.RLock()
	internal := make([]*entities.User, 0, len(r.users))
	for _, user := range r.users {
		select {
		case <-ctx.Done():
			r.mutex.RUnlock()
			return nil, ctx.Err()
		default:
		}
		internal = append(internal, user)
	}
	r.mutex.RUnlock()

	sort.Slice(internal, func(i, j int) bool {
		ai := internal[i].AuditInfo().CreatedAt
		aj := internal[j].AuditInfo().CreatedAt
		if ai.Equal(aj) {
			return internal[i].ID().String() < internal[j].ID().String()
		}
		return ai.Before(aj)
	})

	if offset >= len(internal) {
		return []*entities.User{}, nil
	}

	end := offset + limit
	if end > len(internal) {
		end = len(internal)
	}

	result := make([]*entities.User, 0, end-offset)
	for _, user := range internal[offset:end] {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		result = append(result, r.rehydrateUser(user, user.LastLoginAt(), user.LoginAttempts()))
	}

	return result, nil
}

// Count returns the total number of users
func (r *MemoryUserRepository) Count(ctx context.Context) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	r.mutex.RLock()
	count := int64(len(r.users))
	r.mutex.RUnlock()
	return count, nil
}

// Delete deletes a user by ID
func (r *MemoryUserRepository) Delete(ctx context.Context, id valueObjects.UserID) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if existing, ok := r.users[id.String()]; ok {
		r.removeFromIndexes(existing)
	}

	delete(r.users, id.String())
	r.logInfo("User deleted from memory", logging.String("user_id", id.String()))
	return nil
}

// ExistsByEmail checks if a user exists with the given email
func (r *MemoryUserRepository) ExistsByEmail(ctx context.Context, email valueObjects.Email) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	r.mutex.RLock()
	_, exists := r.emailIndex[email.String()]
	r.mutex.RUnlock()
	return exists, nil
}

// ExistsByUsername checks if a user exists with the given username
func (r *MemoryUserRepository) ExistsByUsername(ctx context.Context, username valueObjects.Username) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	r.mutex.RLock()
	_, exists := r.usernameIndex[username.String()]
	r.mutex.RUnlock()
	return exists, nil
}

// UpdateLastLogin updates the last login time for a user
func (r *MemoryUserRepository) UpdateLastLogin(ctx context.Context, id valueObjects.UserID) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	user, exists := r.users[id.String()]
	if !exists {
		return domainErrors.ErrUserNotFound
	}

	updatedLastLogin := time.Now()
	r.users[id.String()] = r.rehydrateUser(user, &updatedLastLogin, user.LoginAttempts())

	r.logInfo("User last login updated in memory", logging.String("user_id", id.String()))
	return nil
}

// UpdateLoginAttempts updates the login attempts for a user
func (r *MemoryUserRepository) UpdateLoginAttempts(ctx context.Context, id valueObjects.UserID, attempts int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if attempts < 0 {
		attempts = 0
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	user, exists := r.users[id.String()]
	if !exists {
		return domainErrors.ErrUserNotFound
	}

	r.users[id.String()] = r.rehydrateUser(user, user.LastLoginAt(), attempts)

	r.logInfo(
		"User login attempts updated in memory",
		logging.String("user_id", id.String()),
		logging.Int("attempts", attempts),
	)
	return nil
}

func (r *MemoryUserRepository) rehydrateUser(user *entities.User, lastLogin *time.Time, loginAttempts int) *entities.User {
	var lastLoginCopy *time.Time
	if lastLogin != nil {
		copyValue := *lastLogin
		lastLoginCopy = &copyValue
	}

	auditInfo := user.AuditInfo()
	auditCopy := entities.AuditInfo{
		CreatedBy: auditInfo.CreatedBy,
		CreatedAt: auditInfo.CreatedAt,
	}

	clone := entities.NewUserFromRepository(
		user.ID(),
		user.Username(),
		user.Email(),
		user.Password(),
		user.IsActive(),
		lastLoginCopy,
		loginAttempts,
		auditCopy,
	)

	if profile := user.Profile(); profile != nil {
		cloneProfile := *profile
		clone.UpdateProfile(&cloneProfile)
	} else {
		clone.UpdateProfile(nil)
	}

	return clone
}

func (r *MemoryUserRepository) addToIndexes(user *entities.User) {
	userID := user.ID().String()
	r.emailIndex[user.Email().String()] = userID
	r.usernameIndex[user.Username().String()] = userID
}

func (r *MemoryUserRepository) removeFromIndexes(user *entities.User) {
	delete(r.emailIndex, user.Email().String())
	delete(r.usernameIndex, user.Username().String())
}

func (r *MemoryUserRepository) logInfo(msg string, params ...interface{}) {
	if r.logger == nil {
		return
	}
	r.logger.Info(msg, params...)
}
