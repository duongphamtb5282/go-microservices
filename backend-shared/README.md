# Backend-Shared Module

## Overview

This module contains shared components, utilities, and common structures that can be reused across microservices. It provides audit entities, event-driven architecture support, shared models, utilities, and constants.

## Directory Structure

```
backend-shared/
├── README.md                    # This file
├── go.mod                       # Go module file
├── audit/                       # Audit entities and utilities
│   ├── audit.go                 # Base audit entity
│   ├── audit_interface.go       # Audit interface
│   └── audit_utils.go           # Audit utilities
├── events/                      # Event-driven architecture
│   ├── event.go                 # Base event structure
│   ├── event_bus.go             # Event bus interface
│   ├── event_handler.go         # Event handler interface
│   └── payload.go               # Shared event payloads
├── models/                      # Shared models
│   ├── base_model.go            # Base model with audit fields
│   ├── pagination.go            # Pagination models
│   └── response.go              # Common response models
├── utils/                       # Shared utilities
│   ├── validation.go            # Validation utilities
│   ├── crypto.go                # Cryptographic utilities
│   └── time.go                 # Time utilities
└── constants/                   # Shared constants
    ├── status.go                # Status constants
    ├── errors.go                # Error constants
    └── events.go                # Event constants
```

## Core Components

### 1. Audit Entities (`audit/`)

- Base audit entity with CreatedBy, CreatedAt, ModifiedBy, ModifiedAt
- Audit interface for different entity types
- Audit utilities for tracking changes

### 2. Event-Driven Architecture (`events/`)

- Base event structure
- Event bus interface for publishing/subscribing
- Event handler interface
- Shared event payloads

### 3. Shared Models (`models/`)

- Base model with audit fields
- Pagination models
- Common response models

### 4. Utilities (`utils/`)

- Validation utilities
- Cryptographic utilities
- Time utilities

### 5. Constants (`constants/`)

- Status constants
- Error constants
- Event constants

## Usage

### 1. Import the module

```go
import "backend-shared/audit"
import "backend-shared/events"
import "backend-shared/models"
import "backend-shared/utils"
import "backend-shared/constants"
```

### 2. Use audit entities

```go
type User struct {
    audit.AuditEntity
    ID       string `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
}
```

### 3. Use event-driven architecture

```go
// Create event
event := events.NewEvent("user.created", payload)

// Publish event
eventBus.Publish(event)

// Subscribe to events
eventBus.Subscribe("user.created", handler)
```

### 4. Use shared models

```go
// Use base model
type Product struct {
    models.BaseModel
    Name  string `json:"name"`
    Price float64 `json:"price"`
}

// Use pagination
pagination := models.NewPagination(1, 10)
```

## Benefits

### 1. Consistency

- Consistent audit fields across all entities
- Standardized event payloads
- Common response formats

### 2. Reusability

- Shared utilities across microservices
- Common constants and models
- Event-driven architecture support

### 3. Maintainability

- Centralized shared components
- Easy to update across services
- Consistent patterns

### 4. Scalability

- Event-driven architecture support
- Shared payloads for inter-service communication
- Extensible audit system

## Implementation

### Phase 1: Core Components

1. Implement audit entities
2. Create event-driven architecture
3. Add shared models
4. Create utilities

### Phase 2: Integration

1. Integrate with existing microservices
2. Add event handlers
3. Implement audit tracking
4. Add validation

### Phase 3: Advanced Features

1. Event sourcing
2. CQRS support
3. Saga patterns
4. Event replay

## Examples

See the `examples/` directory for complete usage examples:

- `audit_example.go` - Audit entity usage
- `event_example.go` - Event-driven architecture
- `model_example.go` - Shared model usage
- `integration_example.go` - Full integration example

## Best Practices

### 1. Audit Entities

- Always include audit fields
- Use consistent field names
- Track all changes
- Implement audit interface

### 2. Event-Driven Architecture

- Use consistent event naming
- Include correlation IDs
- Handle event failures
- Implement event versioning

### 3. Shared Models

- Use base models
- Implement validation
- Include metadata
- Support pagination

### 4. Utilities

- Keep utilities pure
- Avoid side effects
- Include tests
- Document usage

## Troubleshooting

### Common Issues

1. **Import Errors**: Check module path
2. **Audit Fields**: Ensure all entities include audit fields
3. **Event Handling**: Check event handler registration
4. **Validation**: Verify validation rules

### Debug Commands

```bash
# Check module dependencies
go mod tidy

# Run tests
go test ./...

# Check imports
go list -m all
```

## Next Steps

1. **Implementation**: Implement core components
2. **Integration**: Integrate with microservices
3. **Testing**: Add comprehensive tests
4. **Documentation**: Maintain detailed documentation
5. **Performance**: Monitor and optimize performance
