package gorm

import (
	"sync"

	"backend-core/config"
	"backend-core/logging"

	"gorm.io/gorm"
)

// RepositoryFactory creates and manages repositories
type RepositoryFactory struct {
	config       *config.DatabaseConfig
	gormDB       *gorm.DB
	logger       *logging.Logger
	repositories map[string]interface{}
	mu           sync.RWMutex
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(config *config.DatabaseConfig, gormDB *gorm.DB, logger *logging.Logger) *RepositoryFactory {
	return &RepositoryFactory{
		config:       config,
		gormDB:       gormDB,
		logger:       logger,
		repositories: make(map[string]interface{}),
	}
}

// GetRepository creates or returns a cached repository
func (f *RepositoryFactory) GetRepository(entityType string) interface{} {
	f.mu.RLock()
	if repo, exists := f.repositories[entityType]; exists {
		f.mu.RUnlock()
		return repo
	}
	f.mu.RUnlock()

	f.mu.Lock()
	defer f.mu.Unlock()

	// Double-check after acquiring write lock
	if repo, exists := f.repositories[entityType]; exists {
		return repo
	}

	// Create new repository - simplified for now
	repo := &GormRepository[interface{}]{
		database:   &GormDatabase{},
		entityType: entityType,
		logger:     f.logger,
		tableName:  entityType,
	}

	f.repositories[entityType] = repo
	return repo
}

// ClearCache clears the repository cache
func (f *RepositoryFactory) ClearCache() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.repositories = make(map[string]interface{})
}

// GetCacheSize returns the number of cached repositories
func (f *RepositoryFactory) GetCacheSize() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.repositories)
}
