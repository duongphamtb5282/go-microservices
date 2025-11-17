package events

import (
	"auth-service/src/domain/valueObjects"
	"backend-shared/events"
)

// UserCreated represents a user created domain event
type UserCreated struct {
	*events.Event
}

// NewUserCreated creates a new UserCreated event
func NewUserCreated(userID valueObjects.UserID, username valueObjects.Username, email valueObjects.Email) UserCreated {
	data := map[string]interface{}{
		"user_id":  userID.String(),
		"username": username.String(),
		"email":    email.String(),
	}

	event := UserCreated{
		Event: events.NewEvent("UserCreated", "auth-service", data),
	}

	return event
}

// UserID returns the user ID from the event data
func (e UserCreated) UserID() string {
	if data, ok := e.Data.(map[string]interface{}); ok {
		if userID, exists := data["user_id"]; exists {
			return userID.(string)
		}
	}
	return ""
}

// Username returns the username from the event data
func (e UserCreated) Username() string {
	if data, ok := e.Data.(map[string]interface{}); ok {
		if username, exists := data["username"]; exists {
			return username.(string)
		}
	}
	return ""
}

// Email returns the email from the event data
func (e UserCreated) Email() string {
	if data, ok := e.Data.(map[string]interface{}); ok {
		if email, exists := data["email"]; exists {
			return email.(string)
		}
	}
	return ""
}
