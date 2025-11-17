package http

import (
	"backend-core/logging"
)

// HTTPMiddlewareFactory creates HTTP middleware instances
type HTTPMiddlewareFactory struct {
	logger *logging.Logger
}

// NewHTTPMiddlewareFactory creates a new HTTP middleware factory
func NewHTTPMiddlewareFactory(logger *logging.Logger) *HTTPMiddlewareFactory {
	return &HTTPMiddlewareFactory{
		logger: logger,
	}
}

// CreateCORSMiddleware creates a CORS middleware
func (f *HTTPMiddlewareFactory) CreateCORSMiddleware(config *CORSConfig) *CORSMiddleware {
	return NewCORSMiddleware(config, f.logger)
}

// CreateSecurityHeadersMiddleware creates a security headers middleware
func (f *HTTPMiddlewareFactory) CreateSecurityHeadersMiddleware(config *SecurityHeadersConfig) *SecurityHeadersMiddleware {
	return NewSecurityHeadersMiddleware(config, f.logger)
}

// CreateRequestCorrelationMiddleware creates a unified request/correlation middleware
func (f *HTTPMiddlewareFactory) CreateRequestCorrelationMiddleware(config *RequestCorrelationConfig) *RequestCorrelationMiddleware {
	return NewRequestCorrelationMiddleware(config, f.logger)
}

// CreateDefaultCORSMiddleware creates a CORS middleware with default config
func (f *HTTPMiddlewareFactory) CreateDefaultCORSMiddleware() *CORSMiddleware {
	return NewCORSMiddleware(DefaultCORSConfig(), f.logger)
}

// CreateDefaultSecurityHeadersMiddleware creates a security headers middleware with default config
func (f *HTTPMiddlewareFactory) CreateDefaultSecurityHeadersMiddleware() *SecurityHeadersMiddleware {
	return NewSecurityHeadersMiddleware(DefaultSecurityHeadersConfig(), f.logger)
}

// CreateDefaultRequestCorrelationMiddleware creates a unified request/correlation middleware with default config
func (f *HTTPMiddlewareFactory) CreateDefaultRequestCorrelationMiddleware() *RequestCorrelationMiddleware {
	return NewRequestCorrelationMiddleware(DefaultRequestCorrelationConfig(), f.logger)
}
