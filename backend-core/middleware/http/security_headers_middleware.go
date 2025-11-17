package http

import (
	"backend-core/logging"

	"github.com/gin-gonic/gin"
)

// SecurityHeadersConfig holds security headers configuration
type SecurityHeadersConfig struct {
	ContentTypeOptions      string
	FrameOptions            string
	XSSProtection           string
	ReferrerPolicy          string
	PermissionsPolicy       string
	StrictTransportSecurity string
	ContentSecurityPolicy   string
}

// DefaultSecurityHeadersConfig returns default security headers configuration
func DefaultSecurityHeadersConfig() *SecurityHeadersConfig {
	return &SecurityHeadersConfig{
		ContentTypeOptions:      "nosniff",
		FrameOptions:            "DENY",
		XSSProtection:           "1; mode=block",
		ReferrerPolicy:          "strict-origin-when-cross-origin",
		PermissionsPolicy:       "geolocation=(), microphone=(), camera=()",
		StrictTransportSecurity: "max-age=31536000; includeSubDomains",
		ContentSecurityPolicy:   "default-src 'self'",
	}
}

// SecurityHeadersMiddleware provides security headers functionality
type SecurityHeadersMiddleware struct {
	config *SecurityHeadersConfig
	logger *logging.Logger
}

// NewSecurityHeadersMiddleware creates a new security headers middleware
func NewSecurityHeadersMiddleware(config *SecurityHeadersConfig, logger *logging.Logger) *SecurityHeadersMiddleware {
	if config == nil {
		config = DefaultSecurityHeadersConfig()
	}
	return &SecurityHeadersMiddleware{
		config: config,
		logger: logger,
	}
}

// Handler returns the security headers middleware handler
func (s *SecurityHeadersMiddleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Set security headers
		if s.config.ContentTypeOptions != "" {
			ctx.Header("X-Content-Type-Options", s.config.ContentTypeOptions)
		}

		if s.config.FrameOptions != "" {
			ctx.Header("X-Frame-Options", s.config.FrameOptions)
		}

		if s.config.XSSProtection != "" {
			ctx.Header("X-XSS-Protection", s.config.XSSProtection)
		}

		if s.config.ReferrerPolicy != "" {
			ctx.Header("Referrer-Policy", s.config.ReferrerPolicy)
		}

		if s.config.PermissionsPolicy != "" {
			ctx.Header("Permissions-Policy", s.config.PermissionsPolicy)
		}

		// Only set HSTS for HTTPS requests
		if ctx.Request.TLS != nil && s.config.StrictTransportSecurity != "" {
			ctx.Header("Strict-Transport-Security", s.config.StrictTransportSecurity)
		}

		if s.config.ContentSecurityPolicy != "" {
			ctx.Header("Content-Security-Policy", s.config.ContentSecurityPolicy)
		}

		s.logger.Debug("Security headers applied",
			logging.String("method", ctx.Request.Method),
			logging.String("path", ctx.Request.URL.Path))

		ctx.Next()
	}
}
