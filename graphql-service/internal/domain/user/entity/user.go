package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the system
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username  string             `bson:"username" json:"username"`
	Email     string             `bson:"email" json:"email"`
	FirstName string             `bson:"firstName" json:"firstName"`
	LastName  string             `bson:"lastName" json:"lastName"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// NewUser creates a new user entity
func NewUser(username, email, firstName, lastName string) *User {
	now := time.Now()
	return &User{
		Username:  username,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// GetID returns the user ID as string
func (u *User) GetID() string {
	return u.ID.Hex()
}

// GetFullName returns the user's full name
func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

// Update updates user fields
func (u *User) Update(username, email, firstName, lastName string) {
	u.Username = username
	u.Email = email
	u.FirstName = firstName
	u.LastName = lastName
	u.UpdatedAt = time.Now()
}
