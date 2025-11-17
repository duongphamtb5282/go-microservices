package migration

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"backend-core/logging"
)

// MigrationLoader loads migrations from files
type MigrationLoader struct {
	migrationsPath string
	logger         *logging.Logger
}

// NewMigrationLoader creates a new migration loader
func NewMigrationLoader(migrationsPath string, logger *logging.Logger) *MigrationLoader {
	return &MigrationLoader{
		migrationsPath: migrationsPath,
		logger:         logger,
	}
}

// LoadMigrations loads all migration files from the migrations directory
func (l *MigrationLoader) LoadMigrations() ([]*MigrationImpl, error) {
	var migrations []*MigrationImpl

	err := filepath.WalkDir(l.migrationsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		// Parse migration file
		migration, err := l.parseMigrationFile(path)
		if err != nil {
			l.logger.Error("Failed to parse migration file",
				logging.String("path", path),
				logging.Error(err))
			return err
		}

		if migration != nil {
			migrations = append(migrations, migration)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load migrations: %w", err)
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version() < migrations[j].Version()
	})

	l.logger.Info("Loaded migrations",
		logging.Int("count", len(migrations)))

	return migrations, nil
}

// parseMigrationFile parses a single migration file
func (l *MigrationLoader) parseMigrationFile(filePath string) (*MigrationImpl, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Extract version and name from filename
	// Format: 001_create_users_table.sql
	filename := filepath.Base(filePath)
	parts := strings.Split(filename, "_")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid migration filename format: %s", filename)
	}

	versionStr := parts[0]
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		return nil, fmt.Errorf("invalid version in filename %s: %w", filename, err)
	}

	// Extract name (everything after the first underscore, without .sql)
	name := strings.TrimSuffix(strings.Join(parts[1:], "_"), ".sql")

	// Parse SQL content
	upSQL, downSQL, err := l.parseSQLContent(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse SQL content in %s: %w", filename, err)
	}

	// Calculate checksum
	checksum := l.calculateChecksum(string(content))

	migration := &MigrationImpl{
		VersionNum: version,
		Name:       name,
		UpSQL:      upSQL,
		DownSQL:    downSQL,
		Checksum:   checksum,
	}

	l.logger.Debug("Parsed migration",
		logging.Int("version", migration.Version()),
		logging.String("name", migration.Name),
		logging.String("file", filename))

	return migration, nil
}

// parseSQLContent parses SQL content to extract up and down migrations
func (l *MigrationLoader) parseSQLContent(content string) (string, string, error) {
	// Split content by -- +++++ UP and -- +++++ DOWN markers
	upMarker := "-- +++++ UP"
	downMarker := "-- +++++ DOWN"

	upIndex := strings.Index(content, upMarker)
	downIndex := strings.Index(content, downMarker)

	if upIndex == -1 {
		return "", "", fmt.Errorf("up migration marker not found")
	}

	if downIndex == -1 {
		return "", "", fmt.Errorf("down migration marker not found")
	}

	if downIndex <= upIndex {
		return "", "", fmt.Errorf("down migration marker must come after up migration marker")
	}

	// Extract up migration
	upStart := upIndex + len(upMarker)
	upEnd := downIndex
	upSQL := strings.TrimSpace(content[upStart:upEnd])

	// Extract down migration
	downStart := downIndex + len(downMarker)
	downSQL := strings.TrimSpace(content[downStart:])

	if upSQL == "" {
		return "", "", fmt.Errorf("up migration is empty")
	}

	if downSQL == "" {
		return "", "", fmt.Errorf("down migration is empty")
	}

	return upSQL, downSQL, nil
}

// calculateChecksum calculates SHA256 checksum of the migration content
func (l *MigrationLoader) calculateChecksum(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)
}
