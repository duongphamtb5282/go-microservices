package core

import (
	"context"
	"net/http"
	"time"
)

// Request represents an HTTP request with additional context
type Request struct {
	*http.Request
	Context  context.Context
	Metadata map[string]interface{}
	User     *User
	Session  *Session
	Cache    *CacheInfo
	Metrics  *MetricsInfo
}

// Response represents an HTTP response with additional context
type Response struct {
	*http.Response
	Context  context.Context
	Metadata map[string]interface{}
	Cache    *CacheInfo
	Metrics  *MetricsInfo
	Error    error
}

// User represents an authenticated user
type User struct {
	ID          string
	Username    string
	Email       string
	Roles       []string
	Permissions []string
	Attributes  map[string]interface{}
}

// Session represents a user session
type Session struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
	Data      map[string]interface{}
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CacheInfo represents cache information
type CacheInfo struct {
	Key      string
	TTL      time.Duration
	Strategy string
	Hit      bool
	Miss     bool
	Size     int64
	Tags     []string
}

// MetricsInfo represents metrics information
type MetricsInfo struct {
	Duration     time.Duration
	StatusCode   int
	RequestSize  int64
	ResponseSize int64
	StartTime    time.Time
	EndTime      time.Time
	Tags         map[string]string
}

// MiddlewareConfig represents middleware configuration
type MiddlewareConfig struct {
	Name     string
	Enabled  bool
	Priority int
	Config   map[string]interface{}
}
