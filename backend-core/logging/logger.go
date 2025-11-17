package logging

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"backend-core/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// Logger wraps zapcore.Core with additional functionality
type Logger struct {
	level zapcore.Level
	zapcore.Core
}

// NewLogger creates a new logger instance using LoggingConfig (legacy)
func NewLogger(cfg *config.LoggingConfig) (*Logger, error) {
	// Note: SetDefaults removed - all values must be explicitly configured

	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// Create encoder
	var encoder zapcore.Encoder
	switch cfg.Format {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case "console":
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Create core
	var core zapcore.Core
	level := getZapLevel(cfg.Level)

	switch cfg.Output {
	case "stdout":
		core = zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
	case "stderr":
		core = zapcore.NewCore(encoder, zapcore.AddSync(os.Stderr), level)
	case "file":
		// Ensure log directory exists
		if err := os.MkdirAll(filepath.Dir(cfg.FilePath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// Create file writer with rotation
		fileWriter := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}

		core = zapcore.NewCore(encoder, zapcore.AddSync(fileWriter), level)
	default:
		core = zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
	}

	return &Logger{level: level, Core: core}, nil
}

// NewZapLogger creates a new logger instance using Zap config (recommended)
func NewZapLogger(cfg *config.Zap) (*Logger, error) {
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid zap configuration: %w", err)
	}

	// Ensure log directory exists
	if err := os.MkdirAll(cfg.Directory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create cores
	cores := []zapcore.Core{}
	level := getZapLevel(cfg.Level)

	// Console core (if enabled)
	if cfg.LogInConsole {
		consoleCore := zapcore.NewCore(
			cfg.Encoder(),
			zapcore.AddSync(os.Stdout),
			level,
		)
		cores = append(cores, consoleCore)
	}

	// File core (always enabled for persistence)
	fileWriter := createFileWriter(cfg)

	fileCore := zapcore.NewCore(
		cfg.Encoder(),
		zapcore.AddSync(fileWriter),
		level,
	)
	cores = append(cores, fileCore)

	// Combine cores
	core := zapcore.NewTee(cores...)

	return &Logger{level: level, Core: core}, nil
}

// NewLoggerFromConfig creates a logger from logging config (legacy)
func NewLoggerFromConfig(cfg *config.LoggingConfig) (*Logger, error) {
	return NewLogger(cfg)
}

// NewLoggerFromZapConfig creates a logger from zap config (recommended)
func NewLoggerFromZapConfig(cfg *config.Zap) (*Logger, error) {
	return NewZapLogger(cfg)
}

// WithContext adds context fields to the logger and returns a zap.Logger
func (l *Logger) WithContext(ctx context.Context) *zap.Logger {
	fields := []zap.Field{}

	// Add correlation ID if present
	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		fields = append(fields, zap.Any("correlation_id", correlationID))
	}

	// Add user ID if present
	if userID := ctx.Value("user_id"); userID != nil {
		fields = append(fields, zap.Any("user_id", userID))
	}

	// Add request ID if present
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields = append(fields, zap.Any("request_id", requestID))
	}

	// Add service name if present
	if serviceName := ctx.Value("service_name"); serviceName != nil {
		fields = append(fields, zap.Any("service_name", serviceName))
	}

	// Create a new zap.Logger with the core and fields
	return zap.New(l.Core, zap.AddCaller()).With(fields...)
}

// WithFields creates a new zap.Logger with fields
func (l *Logger) WithFields(fields map[string]interface{}) *zap.Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}
	return zap.New(l.Core, zap.AddCaller()).With(zapFields...)
}

// WithField creates a new zap.Logger with a single field
func (l *Logger) WithField(key string, value interface{}) *zap.Logger {
	return zap.New(l.Core, zap.AddCaller()).With(zap.Any(key, value))
}

// WithError creates a new zap.Logger with an error field
func (l *Logger) WithError(err error) *zap.Logger {
	return zap.New(l.Core, zap.AddCaller()).With(zap.Error(err))
}

// Legacy methods for backward compatibility (deprecated)
func (l *Logger) DebugLegacy(msg string, fields ...zap.Field) {
	if l.Enabled(zapcore.DebugLevel) {
		entry := zapcore.Entry{
			Level:   zapcore.DebugLevel,
			Time:    time.Now(),
			Message: msg,
		}
		l.Write(entry, fields)
	}
}

func (l *Logger) InfoLegacy(msg string, fields ...zap.Field) {
	if l.Enabled(zapcore.InfoLevel) {
		entry := zapcore.Entry{
			Level:   zapcore.InfoLevel,
			Time:    time.Now(),
			Message: msg,
		}
		l.Write(entry, fields)
	}
}

func (l *Logger) WarnLegacy(msg string, fields ...zap.Field) {
	if l.Enabled(zapcore.WarnLevel) {
		entry := zapcore.Entry{
			Level:   zapcore.WarnLevel,
			Time:    time.Now(),
			Message: msg,
		}
		l.Write(entry, fields)
	}
}

func (l *Logger) ErrorLegacy(msg string, fields ...zap.Field) {
	if l.Enabled(zapcore.ErrorLevel) {
		entry := zapcore.Entry{
			Level:   zapcore.ErrorLevel,
			Time:    time.Now(),
			Message: msg,
		}
		l.Write(entry, fields)
	}
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, params ...interface{}) {
	if !l.Enabled(zapcore.FatalLevel) {
		return
	}

	fields := l.convertToFields(params...)

	entry := zapcore.Entry{
		Level:   zapcore.FatalLevel,
		Time:    time.Now(),
		Message: msg,
	}
	l.Write(entry, fields)
	os.Exit(1)
}

// Panic logs a panic message and panics
func (l *Logger) Panic(msg string, params ...interface{}) {
	if !l.Enabled(zapcore.PanicLevel) {
		return
	}

	fields := l.convertToFields(params...)

	entry := zapcore.Entry{
		Level:   zapcore.PanicLevel,
		Time:    time.Now(),
		Message: msg,
	}
	l.Write(entry, fields)
	panic(msg)
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Core.Sync()
}

// GetLevel returns the current log level
func (l *Logger) GetLevel() zapcore.Level {
	return l.level
}

// SetLevel sets the log level
func (l *Logger) SetLevel(level zapcore.Level) {
	l.level = level
}

// Enabled checks if a level is enabled
func (l *Logger) Enabled(level zapcore.Level) bool {
	return l.level.Enabled(level)
}

// Write writes a log entry
func (l *Logger) Write(entry zapcore.Entry, fields []zap.Field) error {
	return l.Core.Write(entry, fields)
}

// createFileWriter creates a file writer based on rotation configuration
func createFileWriter(cfg *config.Zap) *lumberjack.Logger {
	// Base configuration
	writer := &lumberjack.Logger{
		Filename:  filepath.Join(cfg.Directory, "app.log"),
		Compress:  cfg.Compress,
		LocalTime: cfg.LocalTime,
	}

	// Apply rotation strategy based on rotation type
	switch cfg.RotationType {
	case "size":
		// Size-based rotation only
		writer.MaxSize = cfg.MaxSize
		writer.MaxBackups = cfg.MaxBackups
		// Don't set MaxAge for size-only rotation
	case "time":
		// Time-based rotation only
		writer.MaxAge = cfg.MaxAge
		// Don't set MaxSize for time-only rotation
	case "both":
		// Both size and time-based rotation
		writer.MaxSize = cfg.MaxSize
		writer.MaxBackups = cfg.MaxBackups
		writer.MaxAge = cfg.MaxAge
	default:
		// Default to both rotation types
		writer.MaxSize = cfg.MaxSize
		writer.MaxBackups = cfg.MaxBackups
		writer.MaxAge = cfg.MaxAge
	}

	return writer
}

// getZapLevel converts string level to zap level
func getZapLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	case "panic":
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}

// Helper functions for creating zap.Field values
func String(key, val string) zap.Field {
	return zap.String(key, val)
}

func Error(err error) zap.Field {
	return zap.Error(err)
}

func Any(key string, val interface{}) zap.Field {
	return zap.Any(key, val)
}

func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

func Bool(key string, val bool) zap.Field {
	return zap.Bool(key, val)
}

func Duration(key string, val time.Duration) zap.Field {
	return zap.Duration(key, val)
}

func Time(key string, val time.Time) zap.Field {
	return zap.Time(key, val)
}

func Int64(key string, val int64) zap.Field {
	return zap.Int64(key, val)
}

// Field is an alias for zap.Field for convenience
type Field = zap.Field

// convertToFields converts various input types to zap fields efficiently
func (l *Logger) convertToFields(params ...interface{}) []zap.Field {
	if len(params) == 0 {
		return nil
	}

	// Fast path: if single map parameter, convert directly
	if len(params) == 1 {
		if fieldsMap, ok := params[0].(map[string]interface{}); ok {
			return l.mapToFields(fieldsMap)
		}
		if zapFields, ok := params[0].([]zap.Field); ok {
			return zapFields
		}
	}

	// Handle mixed parameters: key-value pairs, maps, and zap.Field instances
	fields := make([]zap.Field, 0, len(params))

	for i := 0; i < len(params); i++ {
		switch param := params[i].(type) {
		case map[string]interface{}:
			// Direct map input
			fields = append(fields, l.mapToFields(param)...)
		case []zap.Field:
			// Direct zap fields
			fields = append(fields, param...)
		case zap.Field:
			// Single zap field
			fields = append(fields, param)
		case string:
			// Key-value pair: string key followed by value
			if i+1 < len(params) {
				fields = append(fields, zap.Any(param, params[i+1]))
				i++ // Skip next parameter as it's consumed
			}
		default:
			// Treat as anonymous field with index
			fields = append(fields, zap.Any(fmt.Sprintf("param_%d", i), param))
		}
	}

	return fields
}

// mapToFields converts a map to zap fields efficiently
func (l *Logger) mapToFields(m map[string]interface{}) []zap.Field {
	if len(m) == 0 {
		return nil
	}

	fields := make([]zap.Field, 0, len(m))
	for k, v := range m {
		fields = append(fields, zap.Any(k, v))
	}
	return fields
}

// Log provides flexible logging that accepts various parameter types
func (l *Logger) Log(level zapcore.Level, msg string, params ...interface{}) {
	if !l.Enabled(level) {
		return
	}

	fields := l.convertToFields(params...)

	entry := zapcore.Entry{
		Level:   level,
		Time:    time.Now(),
		Message: msg,
	}
	l.Write(entry, fields)
}

// Simple convenience methods
func (l *Logger) Debug(msg string, params ...interface{}) {
	l.Log(zapcore.DebugLevel, msg, params...)
}

func (l *Logger) Info(msg string, params ...interface{}) {
	l.Log(zapcore.InfoLevel, msg, params...)
}

func (l *Logger) Warn(msg string, params ...interface{}) {
	l.Log(zapcore.WarnLevel, msg, params...)
}

func (l *Logger) Error(msg string, params ...interface{}) {
	l.Log(zapcore.ErrorLevel, msg, params...)
}
