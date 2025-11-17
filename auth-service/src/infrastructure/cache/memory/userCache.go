package memory

import (
	"context"
	"fmt"
	"sync"

	"auth-service/src/domain/entities"
	"auth-service/src/domain/valueObjects"
	"backend-core/logging"
)

// MemoryUserCache implements user caching using in-memory storage
type MemoryUserCache struct {
	users  map[string]*entities.User
	emails map[string]string // email -> userID mapping
	mutex  sync.RWMutex
	logger *logging.Logger
}

// NewMemoryUserCache creates a new memory user cache
func NewMemoryUserCache(logger *logging.Logger) *MemoryUserCache {
	return &MemoryUserCache{
		users:  make(map[string]*entities.User),
		emails: make(map[string]string),
		logger: logger,
	}
}

// CacheUser caches a user entity
func (c *MemoryUserCache) CacheUser(ctx context.Context, user *entities.User) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	userID := user.ID().String()
	c.users[userID] = user
	c.emails[user.Email().String()] = userID

	c.logger.Info("User cached successfully", logging.String("user_id", userID))
	return nil
}

// GetUser retrieves a user from cache
func (c *MemoryUserCache) GetUser(ctx context.Context, userID valueObjects.UserID) (*entities.User, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	user, exists := c.users[userID.String()]
	if !exists {
		return nil, fmt.Errorf("user not found in cache")
	}

	return user, nil
}

// DeleteUser removes a user from cache
func (c *MemoryUserCache) DeleteUser(ctx context.Context, userID valueObjects.UserID) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	userIDStr := userID.String()
	user, exists := c.users[userIDStr]
	if !exists {
		return fmt.Errorf("user not found in cache")
	}

	// Remove from both maps
	delete(c.users, userIDStr)
	delete(c.emails, user.Email().String())

	c.logger.Info("User deleted from cache", logging.String("user_id", userIDStr))
	return nil
}

// CacheUserByEmail caches a user by email
func (c *MemoryUserCache) CacheUserByEmail(ctx context.Context, email valueObjects.Email, user *entities.User) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	userID := user.ID().String()
	c.users[userID] = user
	c.emails[email.String()] = userID

	c.logger.Info("User cached by email successfully",
		logging.String("user_id", userID),
		logging.String("email", email.String()))
	return nil
}

// GetUserByEmail retrieves a user by email from cache
func (c *MemoryUserCache) GetUserByEmail(ctx context.Context, email valueObjects.Email) (*entities.User, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	userID, exists := c.emails[email.String()]
	if !exists {
		return nil, fmt.Errorf("user not found in cache by email")
	}

	user, exists := c.users[userID]
	if !exists {
		return nil, fmt.Errorf("user not found in cache")
	}

	return user, nil
}
