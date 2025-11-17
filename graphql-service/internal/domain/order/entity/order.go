package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusConfirmed OrderStatus = "CONFIRMED"
	OrderStatusShipped   OrderStatus = "SHIPPED"
	OrderStatusDelivered OrderStatus = "DELIVERED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

// Order represents an order in the system
type Order struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"userId" json:"userId"`
	Items       []OrderItem        `bson:"items" json:"items"`
	Status      OrderStatus        `bson:"status" json:"status"`
	TotalAmount float64            `bson:"totalAmount" json:"totalAmount"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProductID primitive.ObjectID `bson:"productId" json:"productId"`
	Quantity  int                `bson:"quantity" json:"quantity"`
	Price     float64            `bson:"price" json:"price"`
	Subtotal  float64            `bson:"subtotal" json:"subtotal"`
}

// NewOrder creates a new order entity
func NewOrder(userID primitive.ObjectID, items []OrderItem) *Order {
	now := time.Now()

	// Calculate total amount
	var totalAmount float64
	for _, item := range items {
		item.Subtotal = item.Price * float64(item.Quantity)
		totalAmount += item.Subtotal
	}

	return &Order{
		UserID:      userID,
		Items:       items,
		Status:      OrderStatusPending,
		TotalAmount: totalAmount,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// GetID returns the order ID as string
func (o *Order) GetID() string {
	return o.ID.Hex()
}

// GetUserID returns the user ID as string
func (o *Order) GetUserID() string {
	return o.UserID.Hex()
}

// UpdateStatus updates the order status
func (o *Order) UpdateStatus(status OrderStatus) {
	o.Status = status
	o.UpdatedAt = time.Now()
}

// Cancel cancels the order
func (o *Order) Cancel() {
	o.Status = OrderStatusCancelled
	o.UpdatedAt = time.Now()
}
