package masking

import (
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

// MaskingEncoder wraps a zapcore.Encoder to add sensitive data masking
type MaskingEncoder struct {
	zapcore.Encoder
	masker SensitiveDataMasker
}

// NewMaskingEncoder creates a new masking encoder
func NewMaskingEncoder(encoder zapcore.Encoder, masker SensitiveDataMasker) *MaskingEncoder {
	return &MaskingEncoder{
		Encoder: encoder,
		masker:  masker,
	}
}

// EncodeEntry encodes a log entry with sensitive data masking
func (e *MaskingEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	// Mask sensitive fields
	maskedFields := e.masker.MaskZapFields(fields)

	// Encode with masked fields
	return e.Encoder.EncodeEntry(entry, maskedFields)
}

// Clone creates a copy of the encoder
func (e *MaskingEncoder) Clone() zapcore.Encoder {
	return &MaskingEncoder{
		Encoder: e.Encoder.Clone(),
		masker:  e.masker,
	}
}

// MaskingCore wraps a zapcore.Core to add sensitive data masking
type MaskingCore struct {
	zapcore.Core
	masker SensitiveDataMasker
}

// NewMaskingCore creates a new masking core
func NewMaskingCore(core zapcore.Core, masker SensitiveDataMasker) *MaskingCore {
	return &MaskingCore{
		Core:   core,
		masker: masker,
	}
}

// Write writes a log entry with sensitive data masking
func (c *MaskingCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// Mask sensitive fields
	maskedFields := c.masker.MaskZapFields(fields)

	// Write with masked fields
	return c.Core.Write(entry, maskedFields)
}

// With adds structured context to the core
func (c *MaskingCore) With(fields []zapcore.Field) zapcore.Core {
	// Mask sensitive fields
	maskedFields := c.masker.MaskZapFields(fields)

	// Create new core with masked fields
	return &MaskingCore{
		Core:   c.Core.With(maskedFields),
		masker: c.masker,
	}
}

// Check determines whether the core should log at the given level
func (c *MaskingCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return c.Core.Check(entry, checkedEntry)
}

// Sync flushes any buffered logs
func (c *MaskingCore) Sync() error {
	return c.Core.Sync()
}

// Enabled returns whether the core is enabled for the given level
func (c *MaskingCore) Enabled(level zapcore.Level) bool {
	return c.Core.Enabled(level)
}
