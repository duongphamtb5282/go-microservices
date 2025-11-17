package transformers

import "context"

// EncryptionTransformer encrypts sensitive data
type EncryptionTransformer struct {
	name     string
	priority int
	encrypt  func(interface{}) (interface{}, error)
}

// NewEncryptionTransformer creates a new encryption transformer
func NewEncryptionTransformer(name string, priority int, encrypt func(interface{}) (interface{}, error)) *EncryptionTransformer {
	return &EncryptionTransformer{
		name:     name,
		priority: priority,
		encrypt:  encrypt,
	}
}

func (t *EncryptionTransformer) GetName() string {
	return t.name
}

func (t *EncryptionTransformer) GetDescription() string {
	return "Encrypts sensitive data"
}

func (t *EncryptionTransformer) Transform(ctx context.Context, key string, data interface{}) (interface{}, error) {
	if t.encrypt != nil {
		return t.encrypt(data)
	}
	return data, nil
}

func (t *EncryptionTransformer) ShouldTransform(ctx context.Context, key string, data interface{}) bool {
	// Only transform sensitive keys
	return key == "password" || key == "token" || key == "secret"
}

func (t *EncryptionTransformer) GetPriority() int {
	return t.priority
}
