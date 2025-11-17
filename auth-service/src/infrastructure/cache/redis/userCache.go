package redis

import (
	"context"
	"fmt"
	"time"

	"auth-service/src/domain/entities"
	"auth-service/src/domain/valueObjects"
	"backend-core/cache"
	"backend-core/config"
	"backend-core/logging"
)

// RedisUserCache implements user caching using backend-core Redis cache
type RedisUserCache struct {
	cache  cache.Cache
	logger *logging.Logger
}

// NewRedisUserCache creates a new Redis user cache using backend-core
func NewRedisUserCache(cfg *config.RedisConfig, logger *logging.Logger) *RedisUserCache {
	redisCache := cache.NewRedisCache(cfg)
	return &RedisUserCache{
		cache:  redisCache,
		logger: logger,
	}
}

// CacheUser caches a user entity
func (c *RedisUserCache) CacheUser(ctx context.Context, user *entities.User) error {
	key := fmt.Sprintf("user:%s", user.ID().String())

	// Convert user to cache-friendly format
	userData := map[string]interface{}{
		"id":         user.ID().String(),
		"username":   user.Username().String(),
		"email":      user.Email().String(),
		"is_active":  user.IsActive(),
		"created_at": user.AuditInfo().CreatedAt,
		"updated_at": user.AuditInfo().CreatedAt, // Use CreatedAt as fallback
	}

	// Cache for 1 hour using backend-core cache
	err := c.cache.Set(ctx, key, userData, time.Hour)
	if err != nil {
		c.logger.Error("Failed to cache user", logging.Error(err), logging.String("user_id", user.ID().String()))
		return fmt.Errorf("failed to cache user: %w", err)
	}

	c.logger.Info("User cached successfully", logging.String("user_id", user.ID().String()))
	return nil
}

// GetUser retrieves a user from cache
func (c *RedisUserCache) GetUser(ctx context.Context, userID valueObjects.UserID) (*entities.User, error) {
	key := fmt.Sprintf("user:%s", userID.String())

	var userData map[string]interface{}
	err := c.cache.Get(ctx, key, &userData)
	if err != nil {
		if err == cache.ErrCacheMiss {
			return nil, fmt.Errorf("user not found in cache")
		}
		return nil, fmt.Errorf("failed to get user from cache: %w", err)
	}

	// Convert cached data back to domain entity
	username, _ := userData["username"].(string)
	email, _ := userData["email"].(string)
	isActive, _ := userData["is_active"].(bool)
	createdAt, _ := userData["created_at"].(string)

	// Create value objects
	usernameVO, err := valueObjects.NewUsername(username)
	if err != nil {
		return nil, fmt.Errorf("invalid username in cache: %w", err)
	}

	emailVO, err := valueObjects.NewEmail(email)
	if err != nil {
		return nil, fmt.Errorf("invalid email in cache: %w", err)
	}

	// Parse created at time
	createdAtTime, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		createdAtTime = time.Now() // Fallback to current time
	}

	// Create audit info
	auditInfo := entities.AuditInfo{
		CreatedBy: "system",
		CreatedAt: createdAtTime,
	}

	// Create user entity
	user := entities.NewUserFromCache(
		userID,
		usernameVO,
		emailVO,
		isActive,
		auditInfo,
	)

	c.logger.Info("User retrieved from cache successfully", logging.String("user_id", userID.String()))
	return user, nil
}

// DeleteUser removes a user from cache
func (c *RedisUserCache) DeleteUser(ctx context.Context, userID valueObjects.UserID) error {
	key := fmt.Sprintf("user:%s", userID.String())

	err := c.cache.Delete(ctx, key)
	if err != nil {
		c.logger.Error("Failed to delete user from cache", logging.Error(err), logging.String("user_id", userID.String()))
		return fmt.Errorf("failed to delete user from cache: %w", err)
	}

	c.logger.Info("User deleted from cache", logging.String("user_id", userID.String()))
	return nil
}

// CacheUserByEmail caches a user by email
func (c *RedisUserCache) CacheUserByEmail(ctx context.Context, email valueObjects.Email, user *entities.User) error {
	key := fmt.Sprintf("user:email:%s", email.String())

	// Store user ID for email lookup
	err := c.cache.Set(ctx, key, user.ID().String(), time.Hour)
	if err != nil {
		c.logger.Error("Failed to cache user by email", logging.Error(err), logging.String("email", email.String()))
		return fmt.Errorf("failed to cache user by email: %w", err)
	}

	// Also cache the full user data
	return c.CacheUser(ctx, user)
}

// GetUserByEmail retrieves a user by email from cache
func (c *RedisUserCache) GetUserByEmail(ctx context.Context, email valueObjects.Email) (*entities.User, error) {
	key := fmt.Sprintf("user:email:%s", email.String())

	var userIDStr string
	err := c.cache.Get(ctx, key, &userIDStr)
	if err != nil {
		if err == cache.ErrCacheMiss {
			return nil, fmt.Errorf("user not found in cache by email")
		}
		return nil, fmt.Errorf("failed to get user by email from cache: %w", err)
	}

	userID, err := valueObjects.NewUserIDFromString(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in cache: %w", err)
	}

	return c.GetUser(ctx, userID)
}
