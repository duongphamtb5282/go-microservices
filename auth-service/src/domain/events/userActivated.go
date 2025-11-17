package events

import (
	"auth-service/src/domain/valueObjects"
	"backend-shared/events"
)

// UserActivated represents a user activated domain event
type UserActivated struct {
	*events.Event
}

// NewUserActivated creates a new UserActivated event
func NewUserActivated(userID valueObjects.UserID, username valueObjects.Username, email valueObjects.Email) UserActivated {
	data := map[string]interface{}{
		"user_id":  userID.String(),
		"username": username.String(),
		"email":    email.String(),
	}

	event := UserActivated{
		Event: events.NewEvent("UserActivated", "auth-service", data),
	}

	return event
}

// UserID returns the user ID from the event data
func (e UserActivated) UserID() string {
	if data, ok := e.Data.(map[string]interface{}); ok {
		if userID, exists := data["user_id"]; exists {
			return userID.(string)
		}
	}
	return ""
}

// Username returns the username from the event data
func (e UserActivated) Username() string {
	if data, ok := e.Data.(map[string]interface{}); ok {
		if username, exists := data["username"]; exists {
			return username.(string)
		}
	}
	return ""
}

// Email returns the email from the event data
func (e UserActivated) Email() string {
	if data, ok := e.Data.(map[string]interface{}); ok {
		if email, exists := data["email"]; exists {
			return email.(string)
		}
	}
	return ""
}
