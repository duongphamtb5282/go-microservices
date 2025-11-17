package models

import (
	"time"
)

// APIResponse represents a standard API response
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(message string, data interface{}) APIResponse {
	return APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(message string, err *APIError) APIResponse {
	return APIResponse{
		Success:   false,
		Message:   message,
		Error:     err,
		Timestamp: time.Now(),
	}
}

// SetRequestID sets the request ID
func (r *APIResponse) SetRequestID(requestID string) {
	r.RequestID = requestID
}

// APIError represents an API error
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Field   string `json:"field,omitempty"`
}

// NewAPIError creates a new API error
func NewAPIError(code, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}
}

// SetDetails sets the error details
func (e *APIError) SetDetails(details string) {
	e.Details = details
}

// SetField sets the error field
func (e *APIError) SetField(field string) {
	e.Field = field
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string, value interface{}) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	}
}

// ValidationResponse represents a validation response
type ValidationResponse struct {
	Valid   bool              `json:"valid"`
	Errors  []ValidationError `json:"errors,omitempty"`
	Message string            `json:"message,omitempty"`
}

// NewValidationResponse creates a new validation response
func NewValidationResponse(valid bool, errors []ValidationError, message string) ValidationResponse {
	return ValidationResponse{
		Valid:   valid,
		Errors:  errors,
		Message: message,
	}
}

// AddError adds a validation error
func (v *ValidationResponse) AddError(error ValidationError) {
	v.Errors = append(v.Errors, error)
}

// HasErrors checks if there are validation errors
func (v *ValidationResponse) HasErrors() bool {
	return len(v.Errors) > 0
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string                   `json:"status"`
	Timestamp time.Time                `json:"timestamp"`
	Services  map[string]ServiceHealth `json:"services,omitempty"`
	Version   string                   `json:"version,omitempty"`
	Uptime    time.Duration            `json:"uptime,omitempty"`
}

// NewHealthResponse creates a new health response
func NewHealthResponse(status string) HealthResponse {
	return HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Services:  make(map[string]ServiceHealth),
	}
}

// AddService adds a service to the health response
func (h *HealthResponse) AddService(name string, health ServiceHealth) {
	h.Services[name] = health
}

// SetVersion sets the version
func (h *HealthResponse) SetVersion(version string) {
	h.Version = version
}

// SetUptime sets the uptime
func (h *HealthResponse) SetUptime(uptime time.Duration) {
	h.Uptime = uptime
}

// ServiceHealth represents the health of a service
type ServiceHealth struct {
	Status    string                 `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewServiceHealth creates a new service health
func NewServiceHealth(status, message string) ServiceHealth {
	return ServiceHealth{
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
}

// AddMetadata adds metadata to the service health
func (s *ServiceHealth) AddMetadata(key string, value interface{}) {
	if s.Metadata == nil {
		s.Metadata = make(map[string]interface{})
	}
	s.Metadata[key] = value
}

// MetricsResponse represents a metrics response
type MetricsResponse struct {
	Metrics   map[string]interface{} `json:"metrics"`
	Timestamp time.Time              `json:"timestamp"`
	Period    string                 `json:"period,omitempty"`
}

// NewMetricsResponse creates a new metrics response
func NewMetricsResponse(metrics map[string]interface{}) MetricsResponse {
	return MetricsResponse{
		Metrics:   metrics,
		Timestamp: time.Now(),
	}
}

// SetPeriod sets the metrics period
func (m *MetricsResponse) SetPeriod(period string) {
	m.Period = period
}

// AddMetric adds a metric
func (m *MetricsResponse) AddMetric(key string, value interface{}) {
	if m.Metrics == nil {
		m.Metrics = make(map[string]interface{})
	}
	m.Metrics[key] = value
}

// CacheResponse represents a cache response
type CacheResponse struct {
	Hit       bool        `json:"hit"`
	Key       string      `json:"key"`
	Value     interface{} `json:"value,omitempty"`
	TTL       int64       `json:"ttl,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewCacheResponse creates a new cache response
func NewCacheResponse(hit bool, key string, value interface{}) CacheResponse {
	return CacheResponse{
		Hit:       hit,
		Key:       key,
		Value:     value,
		Timestamp: time.Now(),
	}
}

// SetTTL sets the TTL
func (c *CacheResponse) SetTTL(ttl int64) {
	c.TTL = ttl
}

// AuditResponse represents an audit response
type AuditResponse struct {
	EntityID   string                 `json:"entity_id"`
	EntityType string                 `json:"entity_type"`
	Action     string                 `json:"action"`
	UserID     string                 `json:"user_id"`
	Timestamp  time.Time              `json:"timestamp"`
	Changes    map[string]interface{} `json:"changes,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// NewAuditResponse creates a new audit response
func NewAuditResponse(entityID, entityType, action, userID string) AuditResponse {
	return AuditResponse{
		EntityID:   entityID,
		EntityType: entityType,
		Action:     action,
		UserID:     userID,
		Timestamp:  time.Now(),
		Changes:    make(map[string]interface{}),
		Metadata:   make(map[string]interface{}),
	}
}

// AddChange adds a change to the audit response
func (a *AuditResponse) AddChange(field string, oldValue, newValue interface{}) {
	if a.Changes == nil {
		a.Changes = make(map[string]interface{})
	}
	a.Changes[field] = map[string]interface{}{
		"old_value": oldValue,
		"new_value": newValue,
	}
}

// AddMetadata adds metadata to the audit response
func (a *AuditResponse) AddMetadata(key string, value interface{}) {
	if a.Metadata == nil {
		a.Metadata = make(map[string]interface{})
	}
	a.Metadata[key] = value
}
