package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeOrderCreated   NotificationType = "ORDER_CREATED"
	NotificationTypeOrderUpdated   NotificationType = "ORDER_UPDATED"
	NotificationTypeOrderShipped   NotificationType = "ORDER_SHIPPED"
	NotificationTypeOrderDelivered NotificationType = "ORDER_DELIVERED"
	NotificationTypeOrderCancelled NotificationType = "ORDER_CANCELLED"
	NotificationTypeWelcome        NotificationType = "WELCOME"
	NotificationTypePromotion      NotificationType = "PROMOTION"
)

// Notification represents a notification in the system
type Notification struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"userId" json:"userId"`
	Type      NotificationType   `bson:"type" json:"type"`
	Title     string             `bson:"title" json:"title"`
	Message   string             `bson:"message" json:"message"`
	Read      bool               `bson:"read" json:"read"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// NewNotification creates a new notification entity
func NewNotification(userID primitive.ObjectID, notificationType NotificationType, title, message string) *Notification {
	now := time.Now()
	return &Notification{
		UserID:    userID,
		Type:      notificationType,
		Title:     title,
		Message:   message,
		Read:      false,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// GetID returns the notification ID as string
func (n *Notification) GetID() string {
	return n.ID.Hex()
}

// GetUserID returns the user ID as string
func (n *Notification) GetUserID() string {
	return n.UserID.Hex()
}

// MarkAsRead marks the notification as read
func (n *Notification) MarkAsRead() {
	n.Read = true
	n.UpdatedAt = time.Now()
}

// MarkAsUnread marks the notification as unread
func (n *Notification) MarkAsUnread() {
	n.Read = false
	n.UpdatedAt = time.Now()
}

// Update updates notification fields
func (n *Notification) Update(title, message string) {
	n.Title = title
	n.Message = message
	n.UpdatedAt = time.Now()
}
