package config

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	zapcore "go.uber.org/zap/zapcore"
)

// Zap holds Zap logger configuration
type Zap struct {
	Level         string `mapstructure:"level" json:"level" yaml:"level" validate:"required,oneof=debug info warn error fatal panic"`                                  // Log level
	Prefix        string `mapstructure:"prefix" json:"prefix" yaml:"prefix"`                                                                                           // Log prefix
	Format        string `mapstructure:"format" json:"format" yaml:"format" validate:"required,oneof=json console"`                                                    // Output format
	Directory     string `mapstructure:"directory" json:"directory" yaml:"directory" validate:"required"`                                                              // Log directory
	EncodeLevel   string `mapstructure:"encode-level" json:"encode-level" yaml:"encode-level" validate:"required,oneof=Lowercase LowercaseColor Capital CapitalColor"` // Encode level
	StacktraceKey string `mapstructure:"stacktrace-key" json:"stacktrace-key" yaml:"stacktrace-key"`                                                                   // Stack trace key
	ShowLine      bool   `mapstructure:"show-line" json:"show-line" yaml:"show-line"`                                                                                  // Show line numbers
	LogInConsole  bool   `mapstructure:"log-in-console" json:"log-in-console" yaml:"log-in-console"`                                                                   // Log to console
	RetentionDay  int    `mapstructure:"retention-day" json:"retention-day" yaml:"retention-day" validate:"min=0"`                                                     // Log retention days

	// Log rotation settings
	MaxSize      int    `mapstructure:"max-size" json:"max-size" yaml:"max-size" validate:"required,min=1"`                               // Max file size in MB
	MaxBackups   int    `mapstructure:"max-backups" json:"max-backups" yaml:"max-backups" validate:"required,min=0"`                      // Number of backup files
	MaxAge       int    `mapstructure:"max-age" json:"max-age" yaml:"max-age" validate:"required,min=0"`                                  // Max age in days
	Compress     bool   `mapstructure:"compress" json:"compress" yaml:"compress"`                                                         // Compress old files
	LocalTime    bool   `mapstructure:"local-time" json:"local-time" yaml:"local-time"`                                                   // Use local time
	RotationType string `mapstructure:"rotation-type" json:"rotation-type" yaml:"rotation-type" validate:"required,oneof=size time both"` // Rotation type: size, time, both
}

// NewZapConfig creates a new Zap configuration with defaults
func NewZapConfig() *Zap {
	return &Zap{
		Level:         "info",
		Prefix:        "[microservices]",
		Format:        "json",
		Directory:     "logs",
		EncodeLevel:   "Lowercase",
		StacktraceKey: "stacktrace",
		ShowLine:      true,
		LogInConsole:  true,
		RetentionDay:  7,

		// Log rotation defaults
		MaxSize:      100,    // 100 MB
		MaxBackups:   5,      // Keep 5 backup files
		MaxAge:       30,     // Keep files for 30 days
		Compress:     true,   // Compress old files
		LocalTime:    true,   // Use local time
		RotationType: "both", // Rotate by both size and time
	}
}

// Levels returns all log levels from the configured level to FatalLevel
func (z *Zap) Levels() []zapcore.Level {
	levels := make([]zapcore.Level, 0, 7)
	level, err := zapcore.ParseLevel(z.Level)
	if err != nil {
		level = zapcore.DebugLevel
	}
	for ; level <= zapcore.FatalLevel; level++ {
		levels = append(levels, level)
	}
	return levels
}

// Encoder returns the zapcore.Encoder based on the configured format
func (z *Zap) Encoder() zapcore.Encoder {
	config := zapcore.EncoderConfig{
		TimeKey:       "time",
		NameKey:       "name",
		LevelKey:      "level",
		CallerKey:     "caller",
		MessageKey:    "message",
		StacktraceKey: z.StacktraceKey,
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeTime: func(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString(z.Prefix + t.Format("2006-01-02 15:04:05.000"))
		},
		EncodeLevel:    z.LevelEncoder(),
		EncodeCaller:   zapcore.FullCallerEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}
	if z.Format == "json" {
		return zapcore.NewJSONEncoder(config)
	}
	return zapcore.NewConsoleEncoder(config)
}

// LevelEncoder returns the zapcore.LevelEncoder based on the configured encode level
func (z *Zap) LevelEncoder() zapcore.LevelEncoder {
	switch z.EncodeLevel {
	case "Lowercase":
		return zapcore.LowercaseLevelEncoder
	case "LowercaseColor":
		return zapcore.LowercaseColorLevelEncoder
	case "Capital":
		return zapcore.CapitalLevelEncoder
	case "CapitalColor":
		return zapcore.CapitalColorLevelEncoder
	default:
		return zapcore.LowercaseLevelEncoder
	}
}

// Validation methods

// IsValidLevel checks if the configured log level is valid
func (z *Zap) IsValidLevel() bool {
	switch z.Level {
	case "debug", "info", "warn", "error", "fatal", "panic":
		return true
	default:
		return false
	}
}

// IsValidFormat checks if the configured log format is valid
func (z *Zap) IsValidFormat() bool {
	switch z.Format {
	case "json", "console":
		return true
	default:
		return false
	}
}

// IsValidEncodeLevel checks if the configured encode level is valid
func (z *Zap) IsValidEncodeLevel() bool {
	switch z.EncodeLevel {
	case "Lowercase", "LowercaseColor", "Capital", "CapitalColor":
		return true
	default:
		return false
	}
}

// IsValidRotationType checks if the configured rotation type is valid
func (z *Zap) IsValidRotationType() bool {
	switch z.RotationType {
	case "size", "time", "both":
		return true
	default:
		return false
	}
}

// Validate validates the Zap configuration
func (z *Zap) Validate() error {
	validate := validator.New()
	if err := validate.Struct(z); err != nil {
		return fmt.Errorf("zap configuration validation failed: %w", err)
	}
	return nil
}

// Functional Options Pattern

// ZapOption defines a function that configures a Zap instance
type ZapOption func(*Zap)

// WithLevel sets the log level
func WithLevel(level string) ZapOption {
	return func(z *Zap) {
		z.Level = level
	}
}

// WithPrefix sets the log prefix
func WithPrefix(prefix string) ZapOption {
	return func(z *Zap) {
		z.Prefix = prefix
	}
}

// WithFormat sets the log format
func WithFormat(format string) ZapOption {
	return func(z *Zap) {
		z.Format = format
	}
}

// WithDirectory sets the log directory
func WithDirectory(directory string) ZapOption {
	return func(z *Zap) {
		z.Directory = directory
	}
}

// WithEncodeLevel sets the encode level
func WithEncodeLevel(encodeLevel string) ZapOption {
	return func(z *Zap) {
		z.EncodeLevel = encodeLevel
	}
}

// WithStacktraceKey sets the stacktrace key
func WithStacktraceKey(stacktraceKey string) ZapOption {
	return func(z *Zap) {
		z.StacktraceKey = stacktraceKey
	}
}

// WithShowLine sets whether to show line numbers
func WithShowLine(showLine bool) ZapOption {
	return func(z *Zap) {
		z.ShowLine = showLine
	}
}

// WithLogInConsole sets whether to log to console
func WithLogInConsole(logInConsole bool) ZapOption {
	return func(z *Zap) {
		z.LogInConsole = logInConsole
	}
}

// WithRetentionDay sets the log retention days
func WithRetentionDay(retentionDay int) ZapOption {
	return func(z *Zap) {
		z.RetentionDay = retentionDay
	}
}

// WithMaxSize sets the maximum file size in MB
func WithMaxSize(maxSize int) ZapOption {
	return func(z *Zap) {
		z.MaxSize = maxSize
	}
}

// WithMaxBackups sets the maximum number of backup files
func WithMaxBackups(maxBackups int) ZapOption {
	return func(z *Zap) {
		z.MaxBackups = maxBackups
	}
}

// WithMaxAge sets the maximum age of log files in days
func WithMaxAge(maxAge int) ZapOption {
	return func(z *Zap) {
		z.MaxAge = maxAge
	}
}

// WithCompress sets whether to compress old log files
func WithCompress(compress bool) ZapOption {
	return func(z *Zap) {
		z.Compress = compress
	}
}

// WithLocalTime sets whether to use local time for rotation
func WithLocalTime(localTime bool) ZapOption {
	return func(z *Zap) {
		z.LocalTime = localTime
	}
}

// WithRotationType sets the rotation type
func WithRotationType(rotationType string) ZapOption {
	return func(z *Zap) {
		z.RotationType = rotationType
	}
}

// NewZapConfigWithOptions creates a new Zap configuration with options
func NewZapConfigWithOptions(opts ...ZapOption) *Zap {
	z := NewZapConfig()
	for _, opt := range opts {
		opt(z)
	}
	return z
}

// Convenience methods for common configurations

// NewDevelopmentConfig creates a configuration suitable for development
func NewDevelopmentConfig() *Zap {
	return NewZapConfigWithOptions(
		WithLevel("debug"),
		WithFormat("console"),
		WithEncodeLevel("LowercaseColor"),
		WithShowLine(true),
		WithLogInConsole(true),
	)
}

// NewProductionConfig creates a configuration suitable for production
func NewProductionConfig() *Zap {
	return NewZapConfigWithOptions(
		WithLevel("info"),
		WithFormat("json"),
		WithEncodeLevel("Lowercase"),
		WithShowLine(false),
		WithLogInConsole(false),
		WithRetentionDay(30),
	)
}

// NewTestConfig creates a configuration suitable for testing
func NewTestConfig() *Zap {
	return NewZapConfigWithOptions(
		WithLevel("error"),
		WithFormat("console"),
		WithEncodeLevel("Lowercase"),
		WithShowLine(false),
		WithLogInConsole(false),
		WithRetentionDay(1),
	)
}
