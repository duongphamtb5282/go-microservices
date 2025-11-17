package events

import (
	"time"
)

// Config holds event sourcing configuration
type Config struct {
	// Event store settings
	Enabled           bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	TableName         string `mapstructure:"table_name" json:"table_name" yaml:"table_name"`
	SnapshotTableName string `mapstructure:"snapshot_table_name" json:"snapshot_table_name" yaml:"snapshot_table_name"`

	// Snapshot settings
	SnapshotEnabled   bool          `mapstructure:"snapshot_enabled" json:"snapshot_enabled" yaml:"snapshot_enabled"`
	SnapshotInterval  int           `mapstructure:"snapshot_interval" json:"snapshot_interval" yaml:"snapshot_interval"`
	SnapshotRetention time.Duration `mapstructure:"snapshot_retention" json:"snapshot_retention" yaml:"snapshot_retention"`

	// Event processing
	BatchSize         int           `mapstructure:"batch_size" json:"batch_size" yaml:"batch_size"`
	ProcessingTimeout time.Duration `mapstructure:"processing_timeout" json:"processing_timeout" yaml:"processing_timeout"`

	// Serialization
	SerializerType  string `mapstructure:"serializer_type" json:"serializer_type" yaml:"serializer_type"`
	CompressionType string `mapstructure:"compression_type" json:"compression_type" yaml:"compression_type"`
}

// DefaultConfig returns a default events configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:           true,
		TableName:         "events",
		SnapshotTableName: "snapshots",
		SnapshotEnabled:   true,
		SnapshotInterval:  100,
		SnapshotRetention: 30 * 24 * time.Hour, // 30 days
		BatchSize:         100,
		ProcessingTimeout: 30 * time.Second,
		SerializerType:    "json",
		CompressionType:   "gzip",
	}
}
