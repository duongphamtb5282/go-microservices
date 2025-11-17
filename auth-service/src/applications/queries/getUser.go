package queries

// GetUserQuery represents a query to get a user
type GetUserQuery struct {
	UserID string `json:"user_id" validate:"required"`
}

// QueryType returns the query type
func (q GetUserQuery) QueryType() string {
	return "GetUser"
}

// NewGetUserQuery creates a new GetUserQuery
func NewGetUserQuery(userID string) GetUserQuery {
	return GetUserQuery{
		UserID: userID,
	}
}
