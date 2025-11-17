package health

import (
	"fmt"
	"time"
)

// MigrationImpl represents a database migration implementation
type MigrationImpl struct {
	id          string
	version     string
	description string
	appliedAt   time.Time
	checksum    string
}

// NewMigration creates a new migration instance
func NewMigration(id, version, description string, appliedAt time.Time, checksum string) *MigrationImpl {
	return &MigrationImpl{
		id:          id,
		version:     version,
		description: description,
		appliedAt:   appliedAt,
		checksum:    checksum,
	}
}

// GetID returns the migration ID
func (m *MigrationImpl) GetID() string {
	return m.id
}

// GetVersion returns the migration version
func (m *MigrationImpl) GetVersion() string {
	return m.version
}

// GetDescription returns the migration description
func (m *MigrationImpl) GetDescription() string {
	return m.description
}

// GetAppliedAt returns when the migration was applied
func (m *MigrationImpl) GetAppliedAt() time.Time {
	return m.appliedAt
}

// GetChecksum returns the migration checksum
func (m *MigrationImpl) GetChecksum() string {
	return m.checksum
}

// SetID sets the migration ID
func (m *MigrationImpl) SetID(id string) {
	m.id = id
}

// SetVersion sets the migration version
func (m *MigrationImpl) SetVersion(version string) {
	m.version = version
}

// SetDescription sets the migration description
func (m *MigrationImpl) SetDescription(description string) {
	m.description = description
}

// SetAppliedAt sets when the migration was applied
func (m *MigrationImpl) SetAppliedAt(appliedAt time.Time) {
	m.appliedAt = appliedAt
}

// SetChecksum sets the migration checksum
func (m *MigrationImpl) SetChecksum(checksum string) {
	m.checksum = checksum
}

// IsApplied returns true if the migration has been applied
func (m *MigrationImpl) IsApplied() bool {
	return !m.appliedAt.IsZero()
}

// GetAge returns the age of the migration since it was applied
func (m *MigrationImpl) GetAge() time.Duration {
	if m.appliedAt.IsZero() {
		return 0
	}
	return time.Since(m.appliedAt)
}

// String returns a string representation of the migration
func (m *MigrationImpl) String() string {
	return fmt.Sprintf("Migration{id=%s, version=%s, description=%s, appliedAt=%v, checksum=%s}",
		m.id, m.version, m.description, m.appliedAt, m.checksum)
}

// MigrationResultImpl represents the result of a migration check implementation
type MigrationResultImpl struct {
	success    bool
	err        error
	migrations []Migration
	duration   time.Duration
}

// NewMigrationResult creates a new migration result instance
func NewMigrationResult(success bool, err error, migrations []Migration, duration time.Duration) *MigrationResultImpl {
	return &MigrationResultImpl{
		success:    success,
		err:        err,
		migrations: migrations,
		duration:   duration,
	}
}

// GetSuccess returns whether the migration check was successful
func (r *MigrationResultImpl) GetSuccess() bool {
	return r.success
}

// GetError returns the error if any
func (r *MigrationResultImpl) GetError() error {
	return r.err
}

// GetMigrations returns the list of migrations
func (r *MigrationResultImpl) GetMigrations() []Migration {
	return r.migrations
}

// GetDuration returns the duration of the migration check
func (r *MigrationResultImpl) GetDuration() time.Duration {
	return r.duration
}

// SetSuccess sets the success status
func (r *MigrationResultImpl) SetSuccess(success bool) {
	r.success = success
}

// SetError sets the error
func (r *MigrationResultImpl) SetError(err error) {
	r.err = err
}

// SetMigrations sets the migrations list
func (r *MigrationResultImpl) SetMigrations(migrations []Migration) {
	r.migrations = migrations
}

// SetDuration sets the duration
func (r *MigrationResultImpl) SetDuration(duration time.Duration) {
	r.duration = duration
}

// AddMigration adds a migration to the result
func (r *MigrationResultImpl) AddMigration(migration Migration) {
	r.migrations = append(r.migrations, migration)
}

// GetMigrationCount returns the number of migrations
func (r *MigrationResultImpl) GetMigrationCount() int {
	return len(r.migrations)
}

// GetAppliedMigrations returns only the applied migrations
func (r *MigrationResultImpl) GetAppliedMigrations() []Migration {
	var applied []Migration
	for _, migration := range r.migrations {
		if impl, ok := migration.(*MigrationImpl); ok && impl.IsApplied() {
			applied = append(applied, migration)
		}
	}
	return applied
}

// GetPendingMigrations returns only the pending migrations
func (r *MigrationResultImpl) GetPendingMigrations() []Migration {
	var pending []Migration
	for _, migration := range r.migrations {
		if impl, ok := migration.(*MigrationImpl); ok && !impl.IsApplied() {
			pending = append(pending, migration)
		}
	}
	return pending
}

// HasErrors returns true if there are any errors
func (r *MigrationResultImpl) HasErrors() bool {
	return r.err != nil
}

// String returns a string representation of the migration result
func (r *MigrationResultImpl) String() string {
	if r.HasErrors() {
		return fmt.Sprintf("MigrationResult{success=%t, error=%v, migrations=%d, duration=%v}",
			r.success, r.err, len(r.migrations), r.duration)
	}
	return fmt.Sprintf("MigrationResult{success=%t, migrations=%d, duration=%v}",
		r.success, len(r.migrations), r.duration)
}
