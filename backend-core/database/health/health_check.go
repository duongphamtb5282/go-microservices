package health

import (
	"fmt"
	"time"
)

// HealthCheck represents a single health check result
type HealthCheck struct {
	timestamp    time.Time
	status       HealthStatus
	responseTime time.Duration
	err          error
	details      map[string]interface{}
}

// NewHealthCheck creates a new health check result
func NewHealthCheck(status HealthStatus, responseTime time.Duration, err error, details map[string]interface{}) *HealthCheck {
	return &HealthCheck{
		timestamp:    time.Now(),
		status:       status,
		responseTime: responseTime,
		err:          err,
		details:      details,
	}
}

// GetStatus returns the health status
func (h *HealthCheck) GetStatus() HealthStatus {
	return h.status
}

// GetResponseTime returns the response time
func (h *HealthCheck) GetResponseTime() time.Duration {
	return h.responseTime
}

// GetError returns the error if any
func (h *HealthCheck) GetError() error {
	return h.err
}

// GetDetails returns the details map
func (h *HealthCheck) GetDetails() map[string]interface{} {
	return h.details
}

// GetTimestamp returns the timestamp
func (h *HealthCheck) GetTimestamp() time.Time {
	return h.timestamp
}

// SetStatus sets the health status
func (h *HealthCheck) SetStatus(status HealthStatus) {
	h.status = status
}

// SetResponseTime sets the response time
func (h *HealthCheck) SetResponseTime(responseTime time.Duration) {
	h.responseTime = responseTime
}

// SetError sets the error
func (h *HealthCheck) SetError(err error) {
	h.err = err
}

// SetDetails sets the details
func (h *HealthCheck) SetDetails(details map[string]interface{}) {
	h.details = details
}

// AddDetail adds a detail to the details map
func (h *HealthCheck) AddDetail(key string, value interface{}) {
	if h.details == nil {
		h.details = make(map[string]interface{})
	}
	h.details[key] = value
}

// IsHealthy returns true if the health check is healthy
func (h *HealthCheck) IsHealthy() bool {
	return h.status == HealthStatusHealthy
}

// IsDegraded returns true if the health check is degraded
func (h *HealthCheck) IsDegraded() bool {
	return h.status == HealthStatusDegraded
}

// IsUnhealthy returns true if the health check is unhealthy
func (h *HealthCheck) IsUnhealthy() bool {
	return h.status == HealthStatusUnhealthy
}

// HasError returns true if the health check has an error
func (h *HealthCheck) HasError() bool {
	return h.err != nil
}

// GetDuration returns the duration since the health check was created
func (h *HealthCheck) GetDuration() time.Duration {
	return time.Since(h.timestamp)
}

// String returns a string representation of the health check
func (h *HealthCheck) String() string {
	if h.HasError() {
		return fmt.Sprintf("HealthCheck{status=%s, error=%v, responseTime=%v}",
			h.status.String(), h.err, h.responseTime)
	}
	return fmt.Sprintf("HealthCheck{status=%s, responseTime=%v}",
		h.status.String(), h.responseTime)
}
