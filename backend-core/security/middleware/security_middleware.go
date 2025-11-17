package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SecurityMiddleware provides comprehensive security features
type SecurityMiddleware struct {
	config      *SecurityConfig
	logger      *zap.Logger
	rateLimiter *RateLimiter
	ipBlocker   *IPBlocker
}

// SecurityConfig holds security middleware configuration
type SecurityConfig struct {
	// Rate limiting
	RateLimit struct {
		Enabled         bool          `mapstructure:"enabled"`
		RequestsPerMin  int           `mapstructure:"requests_per_minute"`
		BurstLimit      int           `mapstructure:"burst_limit"`
		CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
	} `mapstructure:"rate_limit"`

	// Security headers
	SecurityHeaders struct {
		Enabled            bool   `mapstructure:"enabled"`
		HSTS               bool   `mapstructure:"hsts"`
		ContentTypeOptions bool   `mapstructure:"content_type_options"`
		FrameOptions       string `mapstructure:"frame_options"`
		XSSProtection      bool   `mapstructure:"xss_protection"`
		ReferrerPolicy     string `mapstructure:"referrer_policy"`
	} `mapstructure:"security_headers"`

	// IP restrictions
	AllowedIPs           []string `mapstructure:"allowed_ips"`
	BlockedCountries     []string `mapstructure:"blocked_countries"`
	AllowPrivateNetworks bool     `mapstructure:"allow_private_networks"`

	// Audit
	Audit struct {
		Enabled                     bool `mapstructure:"enabled"`
		LogAllAuthEvents            bool `mapstructure:"log_all_auth_events"`
		LogAdminActions             bool `mapstructure:"log_admin_actions"`
		RetentionDays               int  `mapstructure:"retention_days"`
		SuspiciousActivityThreshold int  `mapstructure:"suspicious_activity_threshold"`
	} `mapstructure:"audit"`

	// API Security
	APISecurity struct {
		RequireAPIKey       bool     `mapstructure:"require_api_key"`
		ValidateContentType bool     `mapstructure:"validate_content_type"`
		MaxRequestSize      string   `mapstructure:"max_request_size"`
		AllowedContentTypes []string `mapstructure:"allowed_content_types"`
	} `mapstructure:"api_security"`
}

// RateLimiter manages request rate limiting
type RateLimiter struct {
	requests map[string]*ClientRequests
	mutex    sync.RWMutex
}

// ClientRequests tracks requests per client
type ClientRequests struct {
	count        int
	windowStart  time.Time
	blocked      bool
	blockedUntil time.Time
}

// IPBlocker manages IP-based blocking
type IPBlocker struct {
	blockedIPs map[string]time.Time
	mutex      sync.RWMutex
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(config *SecurityConfig, logger *zap.Logger) *SecurityMiddleware {
	sm := &SecurityMiddleware{
		config:      config,
		logger:      logger,
		rateLimiter: NewRateLimiter(),
		ipBlocker:   NewIPBlocker(),
	}

	// Initialize rate limiter cleanup
	if config.RateLimit.Enabled {
		go sm.startRateLimitCleanup()
	}

	return sm
}

// Handler returns the security middleware handler
func (sm *SecurityMiddleware) Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Apply security measures in order of priority

			// 1. IP-based restrictions
			if !sm.checkIPRestrictions(r) {
				sm.logger.Warn("Request blocked by IP restrictions",
					zap.String("ip", sm.getClientIP(r)),
					zap.String("path", r.URL.Path),
				)
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}

			// 2. Rate limiting
			if sm.config.RateLimit.Enabled {
				if !sm.checkRateLimit(r) {
					sm.logger.Warn("Request blocked by rate limiting",
						zap.String("ip", sm.getClientIP(r)),
						zap.String("path", r.URL.Path),
					)
					http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
					return
				}
			}

			// 3. Content type validation
			if sm.config.APISecurity.ValidateContentType {
				if !sm.validateContentType(r) {
					sm.logger.Warn("Request blocked by content type validation",
						zap.String("ip", sm.getClientIP(r)),
						zap.String("content_type", r.Header.Get("Content-Type")),
					)
					http.Error(w, "Invalid content type", http.StatusBadRequest)
					return
				}
			}

			// 4. Request size validation
			if !sm.validateRequestSize(r) {
				sm.logger.Warn("Request blocked by size validation",
					zap.String("ip", sm.getClientIP(r)),
					zap.String("content_length", r.Header.Get("Content-Length")),
				)
				http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
				return
			}

			// 5. Add security headers
			if sm.config.SecurityHeaders.Enabled {
				sm.addSecurityHeaders(w)
			}

			// 6. Log security events
			if sm.config.Audit.Enabled {
				sm.logSecurityEvent(r)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// checkIPRestrictions validates IP-based access
func (sm *SecurityMiddleware) checkIPRestrictions(r *http.Request) bool {
	clientIP := sm.getClientIP(r)

	// Check if IP is explicitly blocked
	if sm.ipBlocker.isBlocked(clientIP) {
		return false
	}

	// Check allowed IPs (if specified)
	if len(sm.config.AllowedIPs) > 0 {
		allowed := false
		for _, allowedIP := range sm.config.AllowedIPs {
			if sm.ipMatches(clientIP, allowedIP) {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}

	// Check private networks
	if !sm.config.AllowPrivateNetworks {
		if sm.isPrivateIP(clientIP) {
			return false
		}
	}

	return true
}

// checkRateLimit validates rate limiting
func (sm *SecurityMiddleware) checkRateLimit(r *http.Request) bool {
	clientIP := sm.getClientIP(r)
	return sm.rateLimiter.Allow(clientIP, sm.config.RateLimit.RequestsPerMin, sm.config.RateLimit.BurstLimit)
}

// validateContentType validates request content type
func (sm *SecurityMiddleware) validateContentType(r *http.Request) bool {
	if r.Method == "GET" || r.Method == "HEAD" {
		return true // No content type validation for GET/HEAD
	}

	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		return false
	}

	for _, allowedType := range sm.config.APISecurity.AllowedContentTypes {
		if strings.Contains(contentType, allowedType) {
			return true
		}
	}

	return false
}

// validateRequestSize validates request size
func (sm *SecurityMiddleware) validateRequestSize(r *http.Request) bool {
	if r.ContentLength > sm.parseMaxSize(sm.config.APISecurity.MaxRequestSize) {
		return false
	}
	return true
}

// addSecurityHeaders adds security headers to response
func (sm *SecurityMiddleware) addSecurityHeaders(w http.ResponseWriter) {
	if sm.config.SecurityHeaders.HSTS {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	}

	if sm.config.SecurityHeaders.ContentTypeOptions {
		w.Header().Set("X-Content-Type-Options", "nosniff")
	}

	if sm.config.SecurityHeaders.FrameOptions != "" {
		w.Header().Set("X-Frame-Options", sm.config.SecurityHeaders.FrameOptions)
	}

	if sm.config.SecurityHeaders.XSSProtection {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
	}

	if sm.config.SecurityHeaders.ReferrerPolicy != "" {
		w.Header().Set("Referrer-Policy", sm.config.SecurityHeaders.ReferrerPolicy)
	}

	// Additional security headers
	w.Header().Set("X-DNS-Prefetch-Control", "off")
	w.Header().Set("X-Download-Options", "noopen")
	w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
}

// logSecurityEvent logs security events for audit
func (sm *SecurityMiddleware) logSecurityEvent(r *http.Request) {
	sm.logger.Info("Security event",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("ip", sm.getClientIP(r)),
		zap.String("user_agent", r.Header.Get("User-Agent")),
		zap.String("content_type", r.Header.Get("Content-Type")),
		zap.Int64("content_length", r.ContentLength),
		zap.Time("timestamp", time.Now()),
	)
}

// Helper methods

func (sm *SecurityMiddleware) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP in the chain
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to remote address
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func (sm *SecurityMiddleware) ipMatches(ip, pattern string) bool {
	if pattern == "*" {
		return true
	}

	if strings.Contains(pattern, "/") {
		// CIDR notation
		_, cidr, _ := net.ParseCIDR(pattern)
		return cidr.Contains(net.ParseIP(ip))
	}

	return ip == pattern
}

func (sm *SecurityMiddleware) isPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// Check for private IP ranges
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"::1/128",
		"fe80::/10",
		"fc00::/7",
	}

	for _, rangeStr := range privateRanges {
		_, cidr, err := net.ParseCIDR(rangeStr)
		if err == nil && cidr.Contains(parsedIP) {
			return true
		}
	}

	return false
}

func (sm *SecurityMiddleware) parseMaxSize(sizeStr string) int64 {
	// Parse size strings like "16MB", "1GB", etc.
	// Simplified implementation
	switch {
	case strings.HasSuffix(sizeStr, "KB"):
		return parseSize(sizeStr, 1024)
	case strings.HasSuffix(sizeStr, "MB"):
		return parseSize(sizeStr, 1024*1024)
	case strings.HasSuffix(sizeStr, "GB"):
		return parseSize(sizeStr, 1024*1024*1024)
	default:
		return 16 * 1024 * 1024 // Default 16MB
	}
}

func parseSize(sizeStr string, multiplier int64) int64 {
	// Remove suffix and parse number
	numStr := strings.TrimSuffix(sizeStr, "KB")
	numStr = strings.TrimSuffix(numStr, "MB")
	numStr = strings.TrimSuffix(numStr, "GB")

	var num int64
	fmt.Sscanf(numStr, "%d", &num)
	return num * multiplier
}

func (sm *SecurityMiddleware) startRateLimitCleanup() {
	ticker := time.NewTicker(sm.config.RateLimit.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		sm.rateLimiter.Cleanup()
	}
}

// Rate limiter implementation
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		requests: make(map[string]*ClientRequests),
	}
}

func (rl *RateLimiter) Allow(clientIP string, requestsPerMin, burstLimit int) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	client, exists := rl.requests[clientIP]

	if !exists {
		client = &ClientRequests{
			count:       1,
			windowStart: now,
		}
		rl.requests[clientIP] = client
		return true
	}

	// Check if blocked
	if client.blocked && now.Before(client.blockedUntil) {
		return false
	}

	// Reset if window expired
	if now.Sub(client.windowStart) > time.Minute {
		client.count = 1
		client.windowStart = now
		client.blocked = false
		return true
	}

	// Check rate limit
	if client.count >= requestsPerMin+burstLimit {
		// Block for 1 minute
		client.blocked = true
		client.blockedUntil = now.Add(time.Minute)
		return false
	}

	client.count++
	return true
}

func (rl *RateLimiter) Cleanup() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	for ip, client := range rl.requests {
		// Remove old entries (older than 5 minutes)
		if now.Sub(client.windowStart) > 5*time.Minute {
			delete(rl.requests, ip)
		}
	}
}

// IP blocker implementation
func NewIPBlocker() *IPBlocker {
	return &IPBlocker{
		blockedIPs: make(map[string]time.Time),
	}
}

func (ib *IPBlocker) isBlocked(ip string) bool {
	ib.mutex.RLock()
	defer ib.mutex.RUnlock()

	unblockTime, exists := ib.blockedIPs[ip]
	if !exists {
		return false
	}

	if time.Now().After(unblockTime) {
		// Clean up expired blocks
		go func() {
			ib.mutex.Lock()
			delete(ib.blockedIPs, ip)
			ib.mutex.Unlock()
		}()
		return false
	}

	return true
}

func (ib *IPBlocker) BlockIP(ip string, duration time.Duration) {
	ib.mutex.Lock()
	defer ib.mutex.Unlock()

	ib.blockedIPs[ip] = time.Now().Add(duration)
}
