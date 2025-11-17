package authorization

import (
	"context"
	"fmt"
	"time"

	"auth-service/src/domain/authorization"
	"backend-core/cache"
	"backend-core/logging"
	"backend-core/telemetry"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	gormio "gorm.io/gorm"
)

// permissionRepository implements authorization.PermissionRepository
type permissionRepository struct {
	db        *gormio.DB
	logger    *logging.Logger
	cacheMgr  *cache.CacheManager
	telemetry *telemetry.Telemetry
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(database interface{}, logger *logging.Logger) authorization.PermissionRepository {
	db := extractGormDB(database, logger)
	if db == nil {
		return nil
	}
	return &permissionRepository{
		db:     db,
		logger: logger,
	}
}

// NewPermissionRepositoryWithCache creates a new permission repository with caching enabled
func NewPermissionRepositoryWithCache(database interface{}, logger *logging.Logger, cacheMgr *cache.CacheManager) authorization.PermissionRepository {
	db := extractGormDB(database, logger)
	if db == nil {
		return nil
	}
	return &permissionRepository{
		db:       db,
		logger:   logger,
		cacheMgr: cacheMgr,
	}
}

// NewPermissionRepositoryWithTelemetry creates a new permission repository with telemetry enabled
func NewPermissionRepositoryWithTelemetry(database interface{}, logger *logging.Logger, telemetry *telemetry.Telemetry) authorization.PermissionRepository {
	db := extractGormDB(database, logger)
	if db == nil {
		return nil
	}
	return &permissionRepository{
		db:        db,
		logger:    logger,
		telemetry: telemetry,
	}
}

// NewPermissionRepositoryWithAll creates a new permission repository with all features enabled
func NewPermissionRepositoryWithAll(database interface{}, logger *logging.Logger, cacheMgr *cache.CacheManager, telemetry *telemetry.Telemetry) authorization.PermissionRepository {
	db := extractGormDB(database, logger)
	if db == nil {
		return nil
	}
	return &permissionRepository{
		db:        db,
		logger:    logger,
		cacheMgr:  cacheMgr,
		telemetry: telemetry,
	}
}

// extractGormDB extracts *gorm.DB from various database interface types
func extractGormDB(database interface{}, logger *logging.Logger) *gormio.DB {
	// Try direct gorm database interface
	if db, ok := database.(interface{ GetGormDB() *gormio.DB }); ok {
		return db.GetGormDB()
	}

	// Try core database interface with GetGormDB method
	if coreDB, ok := database.(interface{ GetGormDB() interface{} }); ok {
		if gdb, ok := coreDB.GetGormDB().(*gormio.DB); ok {
			return gdb
		}
	}

	logger.Error("Failed to extract GORM database from provided database interface")
	return nil
}

func (r *permissionRepository) Create(ctx context.Context, permission *authorization.Permission) error {
	ctx, span := r.startSpan(ctx, "permission_repository.create")
	defer span.End()

	span.SetAttributes(
		attribute.String("permission.name", permission.Name),
		attribute.String("permission.resource", permission.Resource),
		attribute.String("permission.action", permission.Action),
	)

	startTime := time.Now()
	permission.SetUUID(uuid.New())
	permission.CreatedAt = time.Now()
	permission.ModifiedAt = time.Now()

	if err := r.db.WithContext(ctx).Create(permission).Error; err != nil {
		r.recordError(span, err)
		r.recordMetric("permission_create_errors_total", 1, map[string]string{"error": err.Error()})
		r.logger.Error("Failed to create permission",
			logging.Error(err),
			logging.String("permission_name", permission.Name))
		return fmt.Errorf("failed to create permission: %w", err)
	}

	span.SetAttributes(attribute.String("permission.id", permission.ID.String()))

	// Record success metrics
	duration := time.Since(startTime).Milliseconds()
	r.recordMetric("permission_create_duration_ms", float64(duration), nil)
	r.recordMetric("permission_create_total", 1, nil)

	// Invalidate related cache entries
	r.invalidatePermissionCache(ctx, permission)

	r.logger.Info("Permission created successfully",
		logging.String("permission_id", permission.ID.String()),
		logging.String("permission_name", permission.Name))

	return nil
}

func (r *permissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*authorization.Permission, error) {
	cacheKey := fmt.Sprintf("permission:id:%s", id.String())

	// Try cache first if cache manager is available
	if r.cacheMgr != nil {
		var permission authorization.Permission
		err := r.cacheMgr.Remember(ctx, cacheKey, &permission, func() (interface{}, error) {
			return r.getByIDFromDB(ctx, id)
		}, time.Minute*5) // Cache for 5 minutes

		if err == nil && permission.ID.String() != "" {
			return &permission, nil
		}
		// If cache miss or error, fall back to database
	}

	return r.getByIDFromDB(ctx, id)
}

// getByIDFromDB retrieves permission from database (internal method)
func (r *permissionRepository) getByIDFromDB(ctx context.Context, id uuid.UUID) (*authorization.Permission, error) {
	var permission authorization.Permission
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&permission).Error; err != nil {
		if err == gormio.ErrRecordNotFound {
			return nil, fmt.Errorf("permission not found: %w", err)
		}
		r.logger.Error("Failed to get permission by ID",
			logging.Error(err),
			logging.String("permission_id", id.String()))
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return &permission, nil
}

func (r *permissionRepository) GetByName(ctx context.Context, name string) (*authorization.Permission, error) {
	cacheKey := fmt.Sprintf("permission:name:%s", name)

	// Try cache first if cache manager is available
	if r.cacheMgr != nil {
		var permission authorization.Permission
		err := r.cacheMgr.Remember(ctx, cacheKey, &permission, func() (interface{}, error) {
			return r.getByNameFromDB(ctx, name)
		}, time.Minute*5) // Cache for 5 minutes

		if err == nil && permission.ID.String() != "" {
			return &permission, nil
		}
		// If cache miss or error, fall back to database
	}

	return r.getByNameFromDB(ctx, name)
}

// getByNameFromDB retrieves permission from database (internal method)
func (r *permissionRepository) getByNameFromDB(ctx context.Context, name string) (*authorization.Permission, error) {
	var permission authorization.Permission
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&permission).Error; err != nil {
		if err == gormio.ErrRecordNotFound {
			return nil, fmt.Errorf("permission not found: %w", err)
		}
		r.logger.Error("Failed to get permission by name",
			logging.Error(err),
			logging.String("permission_name", name))
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return &permission, nil
}

func (r *permissionRepository) GetByResourceAndAction(ctx context.Context, resource, action string) (*authorization.Permission, error) {
	var permission authorization.Permission
	if err := r.db.WithContext(ctx).
		Where("resource = ? AND action = ?", resource, action).
		First(&permission).Error; err != nil {
		if err == gormio.ErrRecordNotFound {
			return nil, fmt.Errorf("permission not found: %w", err)
		}
		r.logger.Error("Failed to get permission by resource and action",
			logging.Error(err),
			logging.String("resource", resource),
			logging.String("action", action))
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return &permission, nil
}

func (r *permissionRepository) GetAll(ctx context.Context) ([]*authorization.Permission, error) {
	var permissions []*authorization.Permission
	if err := r.db.WithContext(ctx).Find(&permissions).Error; err != nil {
		r.logger.Error("Failed to get all permissions", logging.Error(err))
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	return permissions, nil
}

func (r *permissionRepository) GetActivePermissions(ctx context.Context) ([]*authorization.Permission, error) {
	var permissions []*authorization.Permission
	if err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&permissions).Error; err != nil {
		r.logger.Error("Failed to get active permissions", logging.Error(err))
		return nil, fmt.Errorf("failed to get active permissions: %w", err)
	}

	return permissions, nil
}

func (r *permissionRepository) Update(ctx context.Context, permission *authorization.Permission) error {
	permission.UpdateAudit(permission.ModifiedBy)

	if err := r.db.WithContext(ctx).Save(permission).Error; err != nil {
		r.logger.Error("Failed to update permission",
			logging.Error(err),
			logging.String("permission_id", permission.ID.String()))
		return fmt.Errorf("failed to update permission: %w", err)
	}

	// Invalidate related cache entries
	r.invalidatePermissionCache(ctx, permission)

	r.logger.Info("Permission updated successfully",
		logging.String("permission_id", permission.ID.String()),
		logging.String("permission_name", permission.Name))

	return nil
}

func (r *permissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Get permission details before deletion for cache invalidation
	permission, err := r.getByIDFromDB(ctx, id)
	if err != nil {
		return err
	}

	if err := r.db.WithContext(ctx).Delete(&authorization.Permission{}, id).Error; err != nil {
		r.logger.Error("Failed to delete permission",
			logging.Error(err),
			logging.String("permission_id", id.String()))
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	// Invalidate related cache entries
	r.invalidatePermissionCache(ctx, permission)

	r.logger.Info("Permission deleted successfully",
		logging.String("permission_id", id.String()))

	return nil
}

func (r *permissionRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*authorization.Permission, error) {
	var permissions []*authorization.Permission

	// Use GORM Joins to get permissions for a role through role_permissions table
	err := r.db.WithContext(ctx).
		Table("permissions").
		Select("permissions.*").
		Joins("INNER JOIN role_permissions rp ON permissions.id = rp.permission_id").
		Where("rp.role_id = ? AND permissions.is_active = ?", roleID.String(), true).
		Order("permissions.resource, permissions.action").
		Find(&permissions).Error

	if err != nil {
		r.logger.Error("Failed to get role permissions",
			logging.Error(err),
			logging.String("role_id", roleID.String()))
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	return permissions, nil
}

func (r *permissionRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*authorization.Permission, error) {
	var permissions []*authorization.Permission

	// Use GORM Joins to get permissions for a user through roles and role_permissions
	err := r.db.WithContext(ctx).
		Table("permissions").
		Select("DISTINCT permissions.*").
		Joins("INNER JOIN role_permissions rp ON permissions.id = rp.permission_id").
		Joins("INNER JOIN user_roles ur ON rp.role_id = ur.role_id").
		Joins("INNER JOIN roles r ON ur.role_id = r.id").
		Where("ur.user_id = ? AND permissions.is_active = ? AND r.is_active = ?",
			userID.String(), true, true).
		Order("permissions.resource, permissions.action").
		Find(&permissions).Error

	if err != nil {
		r.logger.Error("Failed to get user permissions",
			logging.Error(err),
			logging.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	return permissions, nil
}

func (r *permissionRepository) AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID, assignedBy string) error {
	rolePermission := &authorization.RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
		AssignedAt:   time.Now(),
		AssignedBy:   assignedBy,
	}
	rolePermission.SetUUID(uuid.New())
	rolePermission.CreatedAt = time.Now()
	rolePermission.ModifiedAt = time.Now()

	if err := r.db.WithContext(ctx).Create(rolePermission).Error; err != nil {
		r.logger.Error("Failed to assign permission to role",
			logging.Error(err),
			logging.String("role_id", roleID.String()),
			logging.String("permission_id", permissionID.String()))
		return fmt.Errorf("failed to assign permission: %w", err)
	}

	r.logger.Info("Permission assigned to role successfully",
		logging.String("role_id", roleID.String()),
		logging.String("permission_id", permissionID.String()),
		logging.String("assigned_by", assignedBy))

	return nil
}

func (r *permissionRepository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&authorization.RolePermission{}).Error; err != nil {
		r.logger.Error("Failed to remove permission from role",
			logging.Error(err),
			logging.String("role_id", roleID.String()),
			logging.String("permission_id", permissionID.String()))
		return fmt.Errorf("failed to remove permission: %w", err)
	}

	r.logger.Info("Permission removed from role successfully",
		logging.String("role_id", roleID.String()),
		logging.String("permission_id", permissionID.String()))

	return nil
}

func (r *permissionRepository) CheckUserPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	var count int64

	// Use GORM Joins to check if user has permission through roles and role_permissions
	err := r.db.WithContext(ctx).
		Table("permissions").
		Select("COUNT(DISTINCT permissions.id)").
		Joins("INNER JOIN role_permissions rp ON permissions.id = rp.permission_id").
		Joins("INNER JOIN user_roles ur ON rp.role_id = ur.role_id").
		Joins("INNER JOIN roles r ON ur.role_id = r.id").
		Where("ur.user_id = ? AND permissions.resource = ? AND permissions.action = ? AND permissions.is_active = ? AND r.is_active = ?",
			userID.String(), resource, action, true, true).
		Count(&count).Error

	if err != nil {
		r.logger.Error("Failed to check user permission",
			logging.Error(err),
			logging.String("user_id", userID.String()),
			logging.String("resource", resource),
			logging.String("action", action))
		return false, fmt.Errorf("failed to check permission: %w", err)
	}

	hasPermission := count > 0

	r.logger.Debug("User permission checked",
		logging.String("user_id", userID.String()),
		logging.String("resource", resource),
		logging.String("action", action),
		logging.Bool("has_permission", hasPermission))

	return hasPermission, nil
}

// invalidatePermissionCache invalidates cache entries related to a permission
func (r *permissionRepository) invalidatePermissionCache(ctx context.Context, permission *authorization.Permission) {
	if r.cacheMgr == nil {
		return
	}

	// Invalidate specific permission caches
	cacheKeys := []string{
		fmt.Sprintf("permission:id:%s", permission.ID.String()),
		fmt.Sprintf("permission:name:%s", permission.Name),
		fmt.Sprintf("permission:resource_action:%s_%s", permission.Resource, permission.Action),
		"permission:active", // Invalidate active permissions list
	}

	for _, key := range cacheKeys {
		if err := r.cacheMgr.Forget(ctx, key); err != nil {
			r.logger.Warn("Failed to invalidate cache key",
				logging.String("cache_key", key),
				logging.Error(err))
		}
	}
}

// startSpan starts a new telemetry span if telemetry is enabled
func (r *permissionRepository) startSpan(ctx context.Context, operation string) (context.Context, trace.Span) {
	if r.telemetry == nil {
		return ctx, trace.SpanFromContext(ctx)
	}

	return r.telemetry.GetTracer().Start(ctx, operation)
}

// recordError records an error in the telemetry span
func (r *permissionRepository) recordError(span trace.Span, err error) {
	if span != nil {
		span.RecordError(err)
		span.SetStatus(1, err.Error()) // 1 = Error status
	}
}

// recordMetric records a metric if telemetry is enabled
func (r *permissionRepository) recordMetric(name string, value float64, labels map[string]string) {
	if r.telemetry == nil {
		return
	}

	// This is a simplified metric recording - in a real implementation,
	// you would use the telemetry meter to create proper metrics
	r.logger.Debug("Recording metric",
		logging.String("metric", name),
		logging.Any("value", value),
		logging.Any("labels", labels))
}
