package authorization

import (
	"context"
	"fmt"
	"time"

	"auth-service/src/domain/authorization"
	"backend-core/database/gorm"
	"backend-core/logging"

	"github.com/google/uuid"
	gormio "gorm.io/gorm"
)

// roleRepository implements authorization.RoleRepository
type roleRepository struct {
	*gorm.GormRepository[authorization.Role]
	logger *logging.Logger
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(database interface{}, logger *logging.Logger) authorization.RoleRepository {
	// Type assert to get the gorm database methods
	var gormDB gorm.Database
	if db, ok := database.(gorm.Database); ok {
		gormDB = db
	} else {
		// Fallback: try to get GORM DB from core database interface
		logger.Warn("Database does not implement gorm.Database interface, attempting fallback")
		if coreDB, ok := database.(interface{ GetGormDB() interface{} }); ok {
			if gdb, ok := coreDB.GetGormDB().(*gormio.DB); ok {
				// Create a simple wrapper
				gormDB = &simpleGormWrapper{db: gdb, logger: logger}
			}
		}
		if gormDB == nil {
			logger.Error("Failed to extract GORM database from provided database interface")
			return nil
		}
	}

	baseRepo := gorm.NewGormRepository[authorization.Role](gormDB, "Role", logger)

	return &roleRepository{
		GormRepository: baseRepo,
		logger:         logger,
	}
}

// simpleGormWrapper wraps a GORM DB to implement gorm.Database interface
type simpleGormWrapper struct {
	db     *gormio.DB
	logger *logging.Logger
}

func (w *simpleGormWrapper) GetGormDB() *gormio.DB {
	return w.db
}

// Implement the remaining gorm.Database interface methods
func (w *simpleGormWrapper) Connect(ctx context.Context) error {
	sqlDB, err := w.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (w *simpleGormWrapper) Disconnect(ctx context.Context) error {
	sqlDB, err := w.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (w *simpleGormWrapper) IsConnected() bool {
	sqlDB, err := w.db.DB()
	if err != nil {
		return false
	}
	return sqlDB.Ping() == nil
}

func (w *simpleGormWrapper) HealthCheck(ctx context.Context) error {
	return w.Connect(ctx)
}

func (w *simpleGormWrapper) GetRepository(entityType string) interface{} {
	return nil
}

func (w *simpleGormWrapper) AutoMigrate(ctx context.Context, models ...interface{}) error {
	return w.db.WithContext(ctx).AutoMigrate(models...)
}

func (w *simpleGormWrapper) Migrate(ctx context.Context, models ...interface{}) error {
	return w.AutoMigrate(ctx, models...)
}

func (w *simpleGormWrapper) DropTable(ctx context.Context, models ...interface{}) error {
	return w.db.WithContext(ctx).Migrator().DropTable(models...)
}

func (w *simpleGormWrapper) HasTable(ctx context.Context, model interface{}) bool {
	return w.db.WithContext(ctx).Migrator().HasTable(model)
}

func (w *simpleGormWrapper) GetMigrationStatus(ctx context.Context) ([]gorm.MigrationInfo, error) {
	return nil, fmt.Errorf("not implemented")
}

func (w *simpleGormWrapper) BeginTransaction(ctx context.Context) (gorm.Transaction, error) {
	tx := w.db.WithContext(ctx).Begin()
	return &gormTransactionWrapper{tx: tx}, tx.Error
}

func (w *simpleGormWrapper) WithTransaction(ctx context.Context, fn func(gorm.Transaction) error) error {
	return w.db.WithContext(ctx).Transaction(func(tx *gormio.DB) error {
		return fn(&gormTransactionWrapper{tx: tx})
	})
}

func (w *simpleGormWrapper) GetConfig() gorm.DatabaseConfig {
	return gorm.DatabaseConfig{}
}

func (w *simpleGormWrapper) GetLogger() *logging.Logger {
	return w.logger
}

// gormTransactionWrapper wraps gorm transaction
type gormTransactionWrapper struct {
	tx *gormio.DB
}

func (t *gormTransactionWrapper) Commit() error {
	return t.tx.Commit().Error
}

func (t *gormTransactionWrapper) Rollback() error {
	return t.tx.Rollback().Error
}

func (t *gormTransactionWrapper) GetGormDB() *gormio.DB {
	return t.tx
}

func (r *roleRepository) Create(ctx context.Context, role *authorization.Role) error {
	role.SetUUID(uuid.New())
	role.CreatedAt = time.Now()
	role.ModifiedAt = time.Now()

	if err := r.GormRepository.Create(ctx, role); err != nil {
		r.logger.Error("Failed to create role",
			logging.Error(err),
			logging.String("role_name", role.Name))
		return fmt.Errorf("failed to create role: %w", err)
	}

	r.logger.Info("Role created successfully",
		logging.String("role_id", role.ID.String()),
		logging.String("role_name", role.Name))

	return nil
}

func (r *roleRepository) GetByID(ctx context.Context, id uuid.UUID) (*authorization.Role, error) {
	role, err := r.GormRepository.GetByID(ctx, id.String())
	if err != nil {
		r.logger.Error("Failed to get role by ID",
			logging.Error(err),
			logging.String("role_id", id.String()))
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return nil, fmt.Errorf("role not found")
	}

	return role, nil
}

func (r *roleRepository) GetByName(ctx context.Context, name string) (*authorization.Role, error) {
	query := gorm.Query{
		Filter: map[string]interface{}{
			"name": name,
		},
	}

	roles, err := r.GormRepository.Find(ctx, query)
	if err != nil {
		r.logger.Error("Failed to get role by name",
			logging.Error(err),
			logging.String("role_name", name))
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	if len(roles) == 0 {
		return nil, fmt.Errorf("role not found")
	}

	return roles[0], nil
}

func (r *roleRepository) GetAll(ctx context.Context) ([]*authorization.Role, error) {
	roles, err := r.GormRepository.GetAll(ctx, map[string]interface{}{}, gorm.Pagination{})
	if err != nil {
		r.logger.Error("Failed to get all roles", logging.Error(err))
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	return roles, nil
}

func (r *roleRepository) GetActiveRoles(ctx context.Context) ([]*authorization.Role, error) {
	filter := map[string]interface{}{
		"is_active": true,
	}

	roles, err := r.GormRepository.GetAll(ctx, filter, gorm.Pagination{})
	if err != nil {
		r.logger.Error("Failed to get active roles", logging.Error(err))
		return nil, fmt.Errorf("failed to get active roles: %w", err)
	}

	return roles, nil
}

func (r *roleRepository) Update(ctx context.Context, role *authorization.Role) error {
	role.UpdateAudit(role.ModifiedBy)

	if err := r.GormRepository.Update(ctx, role); err != nil {
		r.logger.Error("Failed to update role",
			logging.Error(err),
			logging.String("role_id", role.ID.String()))
		return fmt.Errorf("failed to update role: %w", err)
	}

	r.logger.Info("Role updated successfully",
		logging.String("role_id", role.ID.String()),
		logging.String("role_name", role.Name))

	return nil
}

func (r *roleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.GormRepository.Delete(ctx, id.String()); err != nil {
		r.logger.Error("Failed to delete role",
			logging.Error(err),
			logging.String("role_id", id.String()))
		return fmt.Errorf("failed to delete role: %w", err)
	}

	r.logger.Info("Role deleted successfully",
		logging.String("role_id", id.String()))

	return nil
}

func (r *roleRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*authorization.Role, error) {
	var roles []*authorization.Role

	// Use GORM Joins to get roles for a user through user_roles table
	err := r.GetGormDB().WithContext(ctx).
		Table("roles").
		Select("roles.*").
		Joins("INNER JOIN user_roles ur ON roles.id = ur.role_id").
		Where("ur.user_id = ? AND roles.is_active = ?", userID.String(), true).
		Order("roles.name").
		Find(&roles).Error

	if err != nil {
		r.logger.Error("Failed to get user roles",
			logging.Error(err),
			logging.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	return roles, nil
}

func (r *roleRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID, assignedBy string) error {
	userRole := &authorization.UserRole{
		UserID:     userID,
		RoleID:     roleID,
		AssignedAt: time.Now(),
		AssignedBy: assignedBy,
	}
	userRole.SetUUID(uuid.New())
	userRole.CreatedAt = time.Now()
	userRole.ModifiedAt = time.Now()

	// Use GormRepository's underlying database connection to create UserRole
	db := r.GetGormDB().WithContext(ctx)
	if err := db.Create(userRole).Error; err != nil {
		r.logger.Error("Failed to assign role to user",
			logging.Error(err),
			logging.String("user_id", userID.String()),
			logging.String("role_id", roleID.String()))
		return fmt.Errorf("failed to assign role: %w", err)
	}

	r.logger.Info("Role assigned to user successfully",
		logging.String("user_id", userID.String()),
		logging.String("role_id", roleID.String()),
		logging.String("assigned_by", assignedBy))

	return nil
}

func (r *roleRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	// Use GormRepository's underlying database connection to delete UserRole
	db := r.GetGormDB().WithContext(ctx)
	if err := db.Where("user_id = ? AND role_id = ?", userID.String(), roleID.String()).
		Delete(&authorization.UserRole{}).Error; err != nil {
		r.logger.Error("Failed to remove role from user",
			logging.Error(err),
			logging.String("user_id", userID.String()),
			logging.String("role_id", roleID.String()))
		return fmt.Errorf("failed to remove role: %w", err)
	}

	r.logger.Info("Role removed from user successfully",
		logging.String("user_id", userID.String()),
		logging.String("role_id", roleID.String()))

	return nil
}
