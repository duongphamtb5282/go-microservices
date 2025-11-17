package masking

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// MaskedLoggerFactory creates loggers with sensitive data masking
type MaskedLoggerFactory struct {
	masker SensitiveDataMasker
}

// NewMaskedLoggerFactory creates a new masked logger factory
func NewMaskedLoggerFactory(masker SensitiveDataMasker) *MaskedLoggerFactory {
	return &MaskedLoggerFactory{
		masker: masker,
	}
}

// CreateLogger creates a new logger with masking
func (f *MaskedLoggerFactory) CreateLogger(config *LoggerConfig) (*zap.Logger, error) {
	// Create encoder
	encoder, err := f.createEncoder(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create encoder: %w", err)
	}

	// Create core
	core, err := f.createCore(config, encoder)
	if err != nil {
		return nil, fmt.Errorf("failed to create core: %w", err)
	}

	// Create logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger, nil
}

// LoggerConfig holds configuration for creating a logger
type LoggerConfig struct {
	Level        zapcore.Level `json:"level" yaml:"level"`
	Format       string        `json:"format" yaml:"format"`
	Output       string        `json:"output" yaml:"output"`
	FilePath     string        `json:"file_path" yaml:"file_path"`
	MaxSize      int           `json:"max_size" yaml:"max_size"`
	MaxBackups   int           `json:"max_backups" yaml:"max_backups"`
	MaxAge       int           `json:"max_age" yaml:"max_age"`
	Compress     bool          `json:"compress" yaml:"compress"`
	LocalTime    bool          `json:"local_time" yaml:"local_time"`
	RotationType string        `json:"rotation_type" yaml:"rotation_type"`
	ShowCaller   bool          `json:"show_caller" yaml:"show_caller"`
	ShowStack    bool          `json:"show_stack" yaml:"show_stack"`
	StackLevel   zapcore.Level `json:"stack_level" yaml:"stack_level"`
}

// createEncoder creates a zap encoder with masking
func (f *MaskedLoggerFactory) createEncoder(config *LoggerConfig) (zapcore.Encoder, error) {
	// Create base encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// Create base encoder
	var baseEncoder zapcore.Encoder
	switch config.Format {
	case "json":
		baseEncoder = zapcore.NewJSONEncoder(encoderConfig)
	case "console":
		baseEncoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		return nil, fmt.Errorf("unsupported format: %s", config.Format)
	}

	// Wrap with masking encoder
	maskingEncoder := NewMaskingEncoder(baseEncoder, f.masker)

	return maskingEncoder, nil
}

// createCore creates a zap core with masking
func (f *MaskedLoggerFactory) createCore(config *LoggerConfig, encoder zapcore.Encoder) (zapcore.Core, error) {
	// Create writer
	writer, err := f.createWriter(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create writer: %w", err)
	}

	// Create base core
	core := zapcore.NewCore(encoder, writer, config.Level)

	// Wrap with masking core
	maskingCore := NewMaskingCore(core, f.masker)

	return maskingCore, nil
}

// createWriter creates a zap writer
func (f *MaskedLoggerFactory) createWriter(config *LoggerConfig) (zapcore.WriteSyncer, error) {
	switch config.Output {
	case "stdout":
		return zapcore.AddSync(os.Stdout), nil
	case "stderr":
		return zapcore.AddSync(os.Stderr), nil
	case "file":
		return f.createFileWriter(config)
	default:
		return nil, fmt.Errorf("unsupported output: %s", config.Output)
	}
}

// createFileWriter creates a file writer with rotation
func (f *MaskedLoggerFactory) createFileWriter(config *LoggerConfig) (zapcore.WriteSyncer, error) {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(config.FilePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create lumberjack logger
	lj := &lumberjack.Logger{
		Filename:   config.FilePath,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
		LocalTime:  config.LocalTime,
	}

	// Apply rotation strategy
	switch config.RotationType {
	case "size":
		// Size-based rotation only
		lj.MaxAge = 0 // Disable time-based rotation
	case "time":
		// Time-based rotation only
		lj.MaxSize = 0 // Disable size-based rotation
	case "both":
		// Both size and time-based rotation (default)
	default:
		// Default to both rotation types
	}

	return zapcore.AddSync(lj), nil
}

// CreateDevelopmentLogger creates a development logger with masking
func (f *MaskedLoggerFactory) CreateDevelopmentLogger() (*zap.Logger, error) {
	config := &LoggerConfig{
		Level:      zapcore.DebugLevel,
		Format:     "console",
		Output:     "stdout",
		ShowCaller: true,
		ShowStack:  true,
		StackLevel: zapcore.ErrorLevel,
	}

	return f.CreateLogger(config)
}

// CreateProductionLogger creates a production logger with masking
func (f *MaskedLoggerFactory) CreateProductionLogger(filePath string) (*zap.Logger, error) {
	config := &LoggerConfig{
		Level:        zapcore.InfoLevel,
		Format:       "json",
		Output:       "file",
		FilePath:     filePath,
		MaxSize:      100,
		MaxBackups:   5,
		MaxAge:       30,
		Compress:     true,
		LocalTime:    true,
		RotationType: "both",
		ShowCaller:   false,
		ShowStack:    true,
		StackLevel:   zapcore.ErrorLevel,
	}

	return f.CreateLogger(config)
}

// CreateTestLogger creates a test logger with masking
func (f *MaskedLoggerFactory) CreateTestLogger() (*zap.Logger, error) {
	config := &LoggerConfig{
		Level:      zapcore.DebugLevel,
		Format:     "console",
		Output:     "stdout",
		ShowCaller: true,
		ShowStack:  false,
	}

	return f.CreateLogger(config)
}
