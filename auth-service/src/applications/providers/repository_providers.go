package providers

import (
	"fmt"

	"auth-service/src/domain/authorization"
	"auth-service/src/domain/repositories"
	memoryCache "auth-service/src/infrastructure/cache/memory"
	redisCache "auth-service/src/infrastructure/cache/redis"
	authorizationRepo "auth-service/src/infrastructure/persistence/authorization"
	memoryRepo "auth-service/src/infrastructure/persistence/memory"
	postgresRepo "auth-service/src/infrastructure/persistence/postgres"
	"backend-core/config"
	"backend-core/database"
	"backend-core/database/gorm"
	"backend-core/database/postgresql"
	"backend-core/logging"
)

// UserRepositoryProvider creates a user repository based on database availability
func UserRepositoryProvider(db database.Database, logger *logging.Logger) repositories.UserRepository {
	logger.Info("🔍 UserRepositoryProvider called", logging.String("db_type", fmt.Sprintf("%T", db)))

	if db != nil {
		// Try to cast to PostgreSQL database wrapper to get GORM database
		if postgresWrapper, ok := db.(*postgresql.DatabaseWrapper); ok {
			logger.Info("✅ Successfully cast to PostgreSQL DatabaseWrapper")
			// Get the underlying PostgreSQL database
			postgresDB := postgresWrapper.PostgreSQLDatabase
			// Get the GORM database from PostgreSQL database
			gormDB := postgresDB.GetGormDB()
			if gormDB != nil {
				logger.Info("✅ Using PostgreSQL UserRepository")
				// Get the config and create a GORM database adapter
				config := postgresDB.GetConfig()
				gormDatabase := gorm.NewGormDatabase(&config.DatabaseConfig, gormDB, logger)
				return postgresRepo.NewPostgresUserRepository(gormDatabase, logger)
			}
			logger.Warn("⚠️ GORM DB is nil, falling back to memory")
		} else {
			logger.Warn("⚠️ Failed to cast to PostgreSQL DatabaseWrapper, falling back to memory",
				logging.String("actual_type", fmt.Sprintf("%T", db)))
		}
		// Use memory repository as fallback when database type is not supported or GORM DB is nil
		return memoryRepo.NewMemoryUserRepository(logger)
	}
	logger.Warn("⚠️ Database is nil, falling back to memory")
	// Use memory repository as fallback when no database
	return memoryRepo.NewMemoryUserRepository(logger)
}

// UserCacheProvider creates a user cache based on database availability
func UserCacheProvider(db database.Database, logger *logging.Logger) repositories.UserCache {
	if db != nil {
		// Use Redis cache when database is available
		redisConfig := &config.RedisConfig{
			Name:     "auth-service-cache",
			Addr:     "localhost:6379", // Use localhost for simplicity
			Password: "",
			DB:       0,
			PoolSize: 10,
		}
		return redisCache.NewRedisUserCache(redisConfig, logger)
	}
	// Use memory cache when database is not available
	return memoryCache.NewMemoryUserCache(logger)
}

// RoleRepositoryProvider creates a role repository
func RoleRepositoryProvider(db database.Database, logger *logging.Logger) authorization.RoleRepository {
	if db == nil {
		logger.Warn("Database is nil, role repository will not function")
		return nil
	}

	return authorizationRepo.NewRoleRepository(db, logger)
}

// PermissionRepositoryProvider creates a permission repository
func PermissionRepositoryProvider(db database.Database, logger *logging.Logger) authorization.PermissionRepository {
	if db == nil {
		logger.Warn("Database is nil, permission repository will not function")
		return nil
	}

	return authorizationRepo.NewPermissionRepository(db, logger)
}
