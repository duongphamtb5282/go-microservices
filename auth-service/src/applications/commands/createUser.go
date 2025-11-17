package commands

// CreateUserCommand represents a command to create a user
type CreateUserCommand struct {
	Username  string `json:"username" validate:"required,min=3,max=20"`
	Email     string `json:"email" validate:"required,email,max=254"`
	Password  string `json:"password" validate:"required,min=8,max=100"`
	CreatedBy string `json:"created_by" validate:"required"`
}

// CommandType returns the command type
func (c CreateUserCommand) CommandType() string {
	return "CreateUser"
}

// NewCreateUserCommand creates a new CreateUserCommand
func NewCreateUserCommand(username, email, password, createdBy string) CreateUserCommand {
	return CreateUserCommand{
		Username:  username,
		Email:     email,
		Password:  password,
		CreatedBy: createdBy,
	}
}
