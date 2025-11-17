# Data Transformers Package

This package provides data transformation capabilities for cache operations, following the Single Responsibility Principle (SRP).

## Package Structure

```
backend-core/cache/redis/transformers/
├── transformer_interface.go        # DataTransformer interface
├── transformer_manager.go         # DataTransformerManager
├── json_transformer.go           # JSONTransformer
├── encryption_transformer.go     # EncryptionTransformer
├── compression_transformer.go    # CompressionTransformer
├── validation_transformer.go     # ValidationTransformer
├── metadata_transformer.go       # MetadataTransformer
├── conditional_transformer.go     # ConditionalTransformer
└── README.md                     # This file
```

## Transformer Interface

All transformers implement the `DataTransformer` interface:

```go
type DataTransformer interface {
    GetName() string
    GetDescription() string
    Transform(ctx context.Context, key string, data interface{}) (interface{}, error)
    ShouldTransform(ctx context.Context, key string, data interface{}) bool
    GetPriority() int
}
```

## Available Transformers

### 1. JSONTransformer (`json_transformer.go`)

**Purpose**: Format data as JSON  
**Priority**: High (1) - should run first  
**Use Case**: Standardize data format

```go
transformer := NewJSONTransformer("json-formatter", 1)
```

### 2. EncryptionTransformer (`encryption_transformer.go`)

**Purpose**: Encrypt sensitive data  
**Priority**: High (2) - after JSON formatting  
**Use Case**: Secure sensitive data (passwords, PII)

```go
encryptFunc := func(data interface{}) (interface{}, error) {
    // Encryption logic
    return encryptedData, nil
}
transformer := NewEncryptionTransformer("encryptor", 2, encryptFunc)
```

### 3. CompressionTransformer (`compression_transformer.go`)

**Purpose**: Compress large data  
**Priority**: Medium (3) - after encryption  
**Use Case**: Reduce memory usage for large datasets

```go
compressFunc := func(data interface{}) (interface{}, error) {
    // Compression logic
    return compressedData, nil
}
transformer := NewCompressionTransformer("compressor", 3, compressFunc)
```

### 4. ValidationTransformer (`validation_transformer.go`)

**Purpose**: Validate data integrity  
**Priority**: Medium (4) - before storage  
**Use Case**: Ensure data quality and integrity

```go
validateFunc := func(data interface{}) (interface{}, error) {
    // Validation logic
    return validatedData, nil
}
transformer := NewValidationTransformer("validator", 4, validateFunc)
```

### 5. MetadataTransformer (`metadata_transformer.go`)

**Purpose**: Add metadata to data  
**Priority**: Low (5) - final processing  
**Use Case**: Add timestamps, version info, etc.

```go
transformer := NewMetadataTransformer("metadata", 5)
```

### 6. ConditionalTransformer (`conditional_transformer.go`)

**Purpose**: Apply transformations based on conditions  
**Priority**: Variable (6) - based on conditions  
**Use Case**: Apply transformations based on data content

```go
condition := func(ctx context.Context, key string, data interface{}) bool {
    return strings.Contains(key, "sensitive")
}
transform := func(ctx context.Context, key string, data interface{}) (interface{}, error) {
    // Conditional transformation logic
    return transformedData, nil
}
transformer := NewConditionalTransformer("conditional", 6, condition, transform)
```

## Transformer Manager

The `DataTransformerManager` manages multiple transformers:

```go
manager := NewDataTransformerManager()

// Add transformers (automatically sorted by priority)
manager.AddTransformer(jsonTransformer)
manager.AddTransformer(encryptionTransformer)
manager.AddTransformer(compressionTransformer)
manager.AddTransformer(validationTransformer)
manager.AddTransformer(metadataTransformer)
manager.AddTransformer(conditionalTransformer)

// Transform data using all applicable transformers
transformedData, err := manager.TransformData(ctx, key, data)
```

## Usage Examples

### Basic Usage

```go
package main

import (
    "context"
    "fmt"

    "backend-core/cache/redis/transformers"
)

func main() {
    // Create transformer manager
    manager := transformers.NewDataTransformerManager()

    // Add JSON transformer
    jsonTransformer := transformers.NewJSONTransformer("json", 1)
    manager.AddTransformer(jsonTransformer)

    // Add encryption transformer
    encryptFunc := func(data interface{}) (interface{}, error) {
        // Simple encryption simulation
        if dataMap, ok := data.(map[string]interface{}); ok {
            if password, exists := dataMap["password"]; exists {
                dataMap["password"] = fmt.Sprintf("encrypted_%s", password)
            }
        }
        return data, nil
    }
    encryptionTransformer := transformers.NewEncryptionTransformer("encrypt", 2, encryptFunc)
    manager.AddTransformer(encryptionTransformer)

    // Transform data
    ctx := context.Background()
    userData := map[string]interface{}{
        "id": 1,
        "username": "john_doe",
        "password": "secret123",
    }

    transformedData, err := manager.TransformData(ctx, "user:1", userData)
    if err != nil {
        fmt.Printf("Transformation error: %v\n", err)
        return
    }

    fmt.Printf("Transformed data: %+v\n", transformedData)
}
```

### Advanced Usage with Custom Transformers

```go
package main

import (
    "context"
    "strings"
    "time"

    "backend-core/cache/redis/transformers"
)

func main() {
    manager := transformers.NewDataTransformerManager()

    // Add conditional transformer for sensitive data
    condition := func(ctx context.Context, key string, data interface{}) bool {
        return strings.Contains(key, "sensitive") || strings.Contains(key, "password")
    }

    transform := func(ctx context.Context, key string, data interface{}) (interface{}, error) {
        if dataMap, ok := data.(map[string]interface{}); ok {
            // Mask sensitive fields
            if password, exists := dataMap["password"]; exists {
                dataMap["password"] = "***masked***"
            }
            if token, exists := dataMap["token"]; exists {
                dataMap["token"] = "***masked***"
            }
        }
        return data, nil
    }

    conditionalTransformer := transformers.NewConditionalTransformer("mask-sensitive", 1, condition, transform)
    manager.AddTransformer(conditionalTransformer)

    // Add metadata transformer
    metadataTransformer := transformers.NewMetadataTransformer("metadata", 2)
    manager.AddTransformer(metadataTransformer)

    // Transform sensitive data
    ctx := context.Background()
    sensitiveData := map[string]interface{}{
        "id": 1,
        "username": "john_doe",
        "password": "secret123",
        "token": "abc123",
    }

    transformedData, err := manager.TransformData(ctx, "sensitive:user:1", sensitiveData)
    if err != nil {
        fmt.Printf("Transformation error: %v\n", err)
        return
    }

    fmt.Printf("Transformed data: %+v\n", transformedData)
}
```

## Transformer Pipeline

Transformers are applied in priority order (lower number = higher priority):

1. **JSONTransformer** (Priority 1) - Format data
2. **EncryptionTransformer** (Priority 2) - Encrypt sensitive data
3. **CompressionTransformer** (Priority 3) - Compress large data
4. **ValidationTransformer** (Priority 4) - Validate data
5. **MetadataTransformer** (Priority 5) - Add metadata
6. **ConditionalTransformer** (Priority 6) - Apply conditional logic

## Best Practices

### 1. Priority Order

- **High Priority (1-2)**: Format and security transformations
- **Medium Priority (3-4)**: Performance and validation transformations
- **Low Priority (5-6)**: Metadata and conditional transformations

### 2. Error Handling

```go
transformedData, err := manager.TransformData(ctx, key, data)
if err != nil {
    // Handle transformation errors
    log.Errorf("Transformation failed: %v", err)
    return originalData, nil
}
```

### 3. Performance Considerations

- Use transformers only when necessary
- Avoid expensive operations in high-priority transformers
- Consider caching transformed data

### 4. Testing

```go
func TestJSONTransformer(t *testing.T) {
    transformer := NewJSONTransformer("test", 1)

    ctx := context.Background()
    data := map[string]interface{}{"key": "value"}

    result, err := transformer.Transform(ctx, "test:key", data)
    assert.NoError(t, err)
    assert.Equal(t, data, result)
}
```

## Migration from Monolithic File

The transformers have been separated from the monolithic `transformers.go` file:

**Before:**

```go
// All transformers in one file (345 lines)
import "backend-core/cache/redis/transformers"
```

**After:**

```go
// Individual transformer files
import "backend-core/cache/redis/transformers"
// All transformers are now in separate files but same package
```

**Benefits:**

- ✅ **Single Responsibility**: Each file has one transformer
- ✅ **Better Maintainability**: Easier to modify individual transformers
- ✅ **Improved Testing**: Test transformers independently
- ✅ **Cleaner Code**: Smaller, focused files (30-50 lines each)
- ✅ **Better Organization**: Logical file structure

## Future Enhancements

### 1. Additional Transformers

- **Base64Transformer**: Base64 encoding/decoding
- **XMLTransformer**: XML format transformation
- **CSVTransformer**: CSV format transformation
- **YAMLTransformer**: YAML format transformation

### 2. Advanced Features

- **Transformer Metrics**: Track transformer performance
- **Dynamic Configuration**: Runtime transformer configuration
- **Transformer Validation**: Validate transformer configurations
- **Transformer Composition**: Combine multiple transformers

### 3. Performance Optimizations

- **Lazy Loading**: Load transformers on demand
- **Caching**: Cache transformation results
- **Parallel Processing**: Transform data in parallel
- **Memory Optimization**: Optimize memory usage
