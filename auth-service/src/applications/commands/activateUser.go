package commands

// ActivateUserCommand represents a command to activate a user
type ActivateUserCommand struct {
	UserID string `json:"user_id" validate:"required"`
}

// CommandType returns the command type
func (c ActivateUserCommand) CommandType() string {
	return "ActivateUser"
}

// NewActivateUserCommand creates a new ActivateUserCommand
func NewActivateUserCommand(userID string) ActivateUserCommand {
	return ActivateUserCommand{
		UserID: userID,
	}
}
