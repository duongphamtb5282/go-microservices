package constants

// Status constants
const (
	// General status
	StatusActive    = "active"
	StatusInactive  = "inactive"
	StatusPending   = "pending"
	StatusApproved  = "approved"
	StatusRejected  = "rejected"
	StatusCancelled = "cancelled"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusSuccess   = "success"
	StatusError     = "error"
	StatusWarning   = "warning"
	StatusInfo      = "info"

	// User status
	UserStatusActive     = "active"
	UserStatusInactive   = "inactive"
	UserStatusSuspended  = "suspended"
	UserStatusDeleted    = "deleted"
	UserStatusPending    = "pending"
	UserStatusVerified   = "verified"
	UserStatusUnverified = "unverified"

	// Order status
	OrderStatusPending    = "pending"
	OrderStatusConfirmed  = "confirmed"
	OrderStatusProcessing = "processing"
	OrderStatusShipped    = "shipped"
	OrderStatusDelivered  = "delivered"
	OrderStatusCancelled  = "cancelled"
	OrderStatusReturned   = "returned"
	OrderStatusRefunded   = "refunded"

	// Payment status
	PaymentStatusPending   = "pending"
	PaymentStatusCompleted = "completed"
	PaymentStatusFailed    = "failed"
	PaymentStatusRefunded  = "refunded"
	PaymentStatusCancelled = "cancelled"

	// Product status
	ProductStatusActive   = "active"
	ProductStatusInactive = "inactive"
	ProductStatusDraft    = "draft"
	ProductStatusArchived = "archived"

	// Event status
	EventStatusPending   = "pending"
	EventStatusPublished = "published"
	EventStatusCancelled = "cancelled"
	EventStatusCompleted = "completed"

	// Notification status
	NotificationStatusPending   = "pending"
	NotificationStatusSent      = "sent"
	NotificationStatusDelivered = "delivered"
	NotificationStatusFailed    = "failed"
	NotificationStatusRead      = "read"
	NotificationStatusUnread    = "unread"

	// Audit status
	AuditStatusCreated = "created"
	AuditStatusUpdated = "updated"
	AuditStatusDeleted = "deleted"
	AuditStatusViewed  = "viewed"

	// Cache status
	CacheStatusHit  = "hit"
	CacheStatusMiss = "miss"
	CacheStatusSet  = "set"
	CacheStatusDel  = "del"

	// Health status
	HealthStatusHealthy   = "healthy"
	HealthStatusUnhealthy = "unhealthy"
	HealthStatusDegraded  = "degraded"
	HealthStatusUnknown   = "unknown"
)

// Priority constants
const (
	PriorityLow      = "low"
	PriorityMedium   = "medium"
	PriorityHigh     = "high"
	PriorityCritical = "critical"
)

// Level constants
const (
	LevelDebug   = "debug"
	LevelInfo    = "info"
	LevelWarning = "warning"
	LevelError   = "error"
	LevelFatal   = "fatal"
)

// Action constants
const (
	ActionCreate  = "create"
	ActionRead    = "read"
	ActionUpdate  = "update"
	ActionDelete  = "delete"
	ActionList    = "list"
	ActionSearch  = "search"
	ActionLogin   = "login"
	ActionLogout  = "logout"
	ActionView    = "view"
	ActionEdit    = "edit"
	ActionApprove = "approve"
	ActionReject  = "reject"
	ActionCancel  = "cancel"
	ActionResume  = "resume"
	ActionPause   = "pause"
	ActionStop    = "stop"
	ActionStart   = "start"
	ActionRestart = "restart"
)

// Entity type constants
const (
	EntityTypeUser         = "user"
	EntityTypeProduct      = "product"
	EntityTypeOrder        = "order"
	EntityTypePayment      = "payment"
	EntityTypeCategory     = "category"
	EntityTypeReview       = "review"
	EntityTypeComment      = "comment"
	EntityTypeMessage      = "message"
	EntityTypeFile         = "file"
	EntityTypeImage        = "image"
	EntityTypeVideo        = "video"
	EntityTypeDocument     = "document"
	EntityTypeAudit        = "audit"
	EntityTypeEvent        = "event"
	EntityTypeNotification = "notification"
)

// Channel constants
const (
	ChannelEmail    = "email"
	ChannelSMS      = "sms"
	ChannelPush     = "push"
	ChannelWebhook  = "webhook"
	ChannelSlack    = "slack"
	ChannelDiscord  = "discord"
	ChannelTelegram = "telegram"
	ChannelWhatsApp = "whatsapp"
)

// Sort direction constants
const (
	SortAsc  = "asc"
	SortDesc = "desc"
)

// Default values
const (
	DefaultPageSize   = 10
	MaxPageSize       = 100
	DefaultTimeout    = 30 // seconds
	DefaultRetryCount = 3
	DefaultCacheTTL   = 300   // seconds
	DefaultSessionTTL = 3600  // seconds
	DefaultTokenTTL   = 86400 // seconds
)

// HTTP status codes
const (
	HTTPStatusOK                  = 200
	HTTPStatusCreated             = 201
	HTTPStatusAccepted            = 202
	HTTPStatusNoContent           = 204
	HTTPStatusBadRequest          = 400
	HTTPStatusUnauthorized        = 401
	HTTPStatusForbidden           = 403
	HTTPStatusNotFound            = 404
	HTTPStatusMethodNotAllowed    = 405
	HTTPStatusConflict            = 409
	HTTPStatusUnprocessableEntity = 422
	HTTPStatusTooManyRequests     = 429
	HTTPStatusInternalServerError = 500
	HTTPStatusBadGateway          = 502
	HTTPStatusServiceUnavailable  = 503
	HTTPStatusGatewayTimeout      = 504
)
