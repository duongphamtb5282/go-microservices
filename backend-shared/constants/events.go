package constants

// Event type constants
const (
	// User events
	EventUserCreated         = "user.created"
	EventUserUpdated         = "user.updated"
	EventUserDeleted         = "user.deleted"
	EventUserActivated       = "user.activated"
	EventUserDeactivated     = "user.deactivated"
	EventUserVerified        = "user.verified"
	EventUserUnverified      = "user.unverified"
	EventUserLogin           = "user.login"
	EventUserLogout          = "user.logout"
	EventUserPasswordChanged = "user.password_changed"
	EventUserEmailChanged    = "user.email_changed"

	// Product events
	EventProductCreated      = "product.created"
	EventProductUpdated      = "product.updated"
	EventProductDeleted      = "product.deleted"
	EventProductActivated    = "product.activated"
	EventProductDeactivated  = "product.deactivated"
	EventProductStockUpdated = "product.stock_updated"
	EventProductPriceChanged = "product.price_changed"

	// Order events
	EventOrderCreated   = "order.created"
	EventOrderUpdated   = "order.updated"
	EventOrderCancelled = "order.cancelled"
	EventOrderConfirmed = "order.confirmed"
	EventOrderShipped   = "order.shipped"
	EventOrderDelivered = "order.delivered"
	EventOrderReturned  = "order.returned"
	EventOrderRefunded  = "order.refunded"

	// Payment events
	EventPaymentCreated   = "payment.created"
	EventPaymentCompleted = "payment.completed"
	EventPaymentFailed    = "payment.failed"
	EventPaymentRefunded  = "payment.refunded"
	EventPaymentCancelled = "payment.cancelled"

	// Category events
	EventCategoryCreated = "category.created"
	EventCategoryUpdated = "category.updated"
	EventCategoryDeleted = "category.deleted"

	// Review events
	EventReviewCreated  = "review.created"
	EventReviewUpdated  = "review.updated"
	EventReviewDeleted  = "review.deleted"
	EventReviewApproved = "review.approved"
	EventReviewRejected = "review.rejected"

	// Comment events
	EventCommentCreated  = "comment.created"
	EventCommentUpdated  = "comment.updated"
	EventCommentDeleted  = "comment.deleted"
	EventCommentApproved = "comment.approved"
	EventCommentRejected = "comment.rejected"

	// Message events
	EventMessageCreated   = "message.created"
	EventMessageUpdated   = "message.updated"
	EventMessageDeleted   = "message.deleted"
	EventMessageSent      = "message.sent"
	EventMessageDelivered = "message.delivered"
	EventMessageRead      = "message.read"

	// File events
	EventFileUploaded   = "file.uploaded"
	EventFileDownloaded = "file.downloaded"
	EventFileDeleted    = "file.deleted"
	EventFileProcessed  = "file.processed"

	// Image events
	EventImageUploaded  = "image.uploaded"
	EventImageProcessed = "image.processed"
	EventImageDeleted   = "image.deleted"
	EventImageResized   = "image.resized"

	// Video events
	EventVideoUploaded   = "video.uploaded"
	EventVideoProcessed  = "video.processed"
	EventVideoDeleted    = "video.deleted"
	EventVideoTranscoded = "video.transcoded"

	// Document events
	EventDocumentUploaded  = "document.uploaded"
	EventDocumentProcessed = "document.processed"
	EventDocumentDeleted   = "document.deleted"
	EventDocumentConverted = "document.converted"

	// Audit events
	EventAuditCreated = "audit.created"
	EventAuditUpdated = "audit.updated"
	EventAuditDeleted = "audit.deleted"
	EventAuditViewed  = "audit.viewed"

	// System events
	EventSystemStarted     = "system.started"
	EventSystemStopped     = "system.stopped"
	EventSystemRestarted   = "system.restarted"
	EventSystemMaintenance = "system.maintenance"
	EventSystemError       = "system.error"
	EventSystemWarning     = "system.warning"
	EventSystemInfo        = "system.info"

	// Notification events
	EventNotificationCreated   = "notification.created"
	EventNotificationSent      = "notification.sent"
	EventNotificationDelivered = "notification.delivered"
	EventNotificationFailed    = "notification.failed"
	EventNotificationRead      = "notification.read"
	EventNotificationUnread    = "notification.unread"

	// Cache events
	EventCacheHit     = "cache.hit"
	EventCacheMiss    = "cache.miss"
	EventCacheSet     = "cache.set"
	EventCacheDel     = "cache.del"
	EventCacheExpired = "cache.expired"
	EventCacheCleared = "cache.cleared"

	// Database events
	EventDatabaseConnected    = "database.connected"
	EventDatabaseDisconnected = "database.disconnected"
	EventDatabaseError        = "database.error"
	EventDatabaseSlowQuery    = "database.slow_query"
	EventDatabaseLockTimeout  = "database.lock_timeout"

	// API events
	EventAPIRequest   = "api.request"
	EventAPIResponse  = "api.response"
	EventAPIError     = "api.error"
	EventAPITimeout   = "api.timeout"
	EventAPIRateLimit = "api.rate_limit"

	// Security events
	EventSecurityLoginAttempt       = "security.login_attempt"
	EventSecurityLoginSuccess       = "security.login_success"
	EventSecurityLoginFailure       = "security.login_failure"
	EventSecurityLogout             = "security.logout"
	EventSecurityPasswordReset      = "security.password_reset"
	EventSecurityAccountLocked      = "security.account_locked"
	EventSecurityAccountUnlocked    = "security.account_unlocked"
	EventSecuritySuspiciousActivity = "security.suspicious_activity"
	EventSecurityDataBreach         = "security.data_breach"

	// Performance events
	EventPerformanceSlowRequest = "performance.slow_request"
	EventPerformanceHighCPU     = "performance.high_cpu"
	EventPerformanceHighMemory  = "performance.high_memory"
	EventPerformanceHighDisk    = "performance.high_disk"
	EventPerformanceHighNetwork = "performance.high_network"

	// Health events
	EventHealthCheck    = "health.check"
	EventHealthUp       = "health.up"
	EventHealthDown     = "health.down"
	EventHealthDegraded = "health.degraded"

	// Deployment events
	EventDeploymentStarted   = "deployment.started"
	EventDeploymentCompleted = "deployment.completed"
	EventDeploymentFailed    = "deployment.failed"
	EventDeploymentRollback  = "deployment.rollback"

	// Backup events
	EventBackupStarted   = "backup.started"
	EventBackupCompleted = "backup.completed"
	EventBackupFailed    = "backup.failed"
	EventBackupRestored  = "backup.restored"

	// Monitoring events
	EventMonitoringAlert     = "monitoring.alert"
	EventMonitoringRecovery  = "monitoring.recovery"
	EventMonitoringThreshold = "monitoring.threshold"
	EventMonitoringAnomaly   = "monitoring.anomaly"
)

// Event source constants
const (
	SourceAuthService         = "auth-service"
	SourceUserService         = "user-service"
	SourceProductService      = "product-service"
	SourceOrderService        = "order-service"
	SourcePaymentService      = "payment-service"
	SourceNotificationService = "notification-service"
	SourceFileService         = "file-service"
	SourceAuditService        = "audit-service"
	SourceSystemService       = "system-service"
	SourceMonitoringService   = "monitoring-service"
	SourceBackupService       = "backup-service"
	SourceDeploymentService   = "deployment-service"
)

// Event priority constants
const (
	EventPriorityLow      = "low"
	EventPriorityMedium   = "medium"
	EventPriorityHigh     = "high"
	EventPriorityCritical = "critical"
)

// Event category constants
const (
	EventCategoryUser         = "user"
	EventCategoryProduct      = "product"
	EventCategoryOrder        = "order"
	EventCategoryPayment      = "payment"
	EventCategorySystem       = "system"
	EventCategorySecurity     = "security"
	EventCategoryPerformance  = "performance"
	EventCategoryHealth       = "health"
	EventCategoryDeployment   = "deployment"
	EventCategoryBackup       = "backup"
	EventCategoryMonitoring   = "monitoring"
	EventCategoryAudit        = "audit"
	EventCategoryNotification = "notification"
	EventCategoryCache        = "cache"
	EventCategoryDatabase     = "database"
	EventCategoryAPI          = "api"
)
