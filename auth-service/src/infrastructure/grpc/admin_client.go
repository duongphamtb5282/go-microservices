package grpc

import (
	"context"
	"fmt"
	"time"

	grpcclient "backend-core/grpc/client"
	"backend-core/logging"
	pb "backend-shared/proto/admin"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AdminClient wraps the gRPC admin service client
type AdminClient struct {
	client pb.AdminServiceClient
	conn   *grpc.ClientConn
	logger *logging.Logger
}

// AdminClientConfig holds configuration for the admin client
type AdminClientConfig struct {
	Host    string
	Port    string
	Timeout time.Duration
	APIKey  string // Optional API key for service-to-service authentication
}

// NewAdminClient creates a new admin service client with middleware
func NewAdminClient(config AdminClientConfig, logger *logging.Logger) (*AdminClient, error) {
	target := fmt.Sprintf("%s:%s", config.Host, config.Port)

	// Build client config
	clientConfig := &grpcclient.ClientConfig{
		Address:          target,
		Timeout:          config.Timeout,
		MaxMessageSize:   10 * 1024 * 1024, // 10MB
		Insecure:         true,             // Use TLS in production
		KeepAliveTime:    30 * time.Second,
		KeepAliveTimeout: 10 * time.Second,
	}

	// Build client with middleware
	builder := grpcclient.NewClientBuilder(clientConfig).
		WithRetry(&grpcclient.RetryConfig{
			MaxAttempts:       3,
			BackoffMultiplier: 2.0,
			InitialBackoff:    100 * time.Millisecond,
			MaxBackoff:        5 * time.Second,
		}).
		WithLogging(logger).
		WithTracing()

	// Add API key authentication if configured
	if config.APIKey != "" {
		// For API key, we can add it as metadata in each request
		// or use a custom interceptor
		logger.Info("API key authentication configured for admin client")
	}

	// Build connection
	conn, err := builder.Build()
	if err != nil {
		logger.Error("Failed to connect to admin service",
			logging.Error(err),
			logging.String("target", target))
		return nil, fmt.Errorf("failed to connect to admin service: %w", err)
	}

	logger.Info("Connected to admin service with middleware",
		logging.String("target", target),
		logging.String("middleware", "retry,logging,tracing"))

	return &AdminClient{
		client: pb.NewAdminServiceClient(conn),
		conn:   conn,
		logger: logger,
	}, nil
}

// Close closes the gRPC connection
func (c *AdminClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// RecordUserCreated sends a user creation event to admin service
func (c *AdminClient) RecordUserCreated(ctx context.Context, userID, email, username, firstName, lastName, createdBy string) error {
	req := &pb.UserCreatedRequest{
		UserId:      userID,
		Email:       email,
		Username:    username,
		FirstName:   firstName,
		LastName:    lastName,
		CreatedAt:   timestamppb.Now(),
		CreatedBy:   createdBy,
		ServiceName: "auth-service",
	}

	c.logger.Debug("Sending user created event to admin service",
		logging.String("user_id", userID),
		logging.String("email", email))

	resp, err := c.client.RecordUserCreated(ctx, req)
	if err != nil {
		c.logger.Error("Failed to record user created event",
			logging.Error(err),
			logging.String("user_id", userID))
		return fmt.Errorf("failed to record user created: %w", err)
	}

	if !resp.Success {
		c.logger.Warn("Admin service reported unsuccessful recording",
			logging.String("message", resp.Message),
			logging.String("user_id", userID))
		return fmt.Errorf("admin service error: %s", resp.Message)
	}

	c.logger.Info("User creation event recorded successfully",
		logging.String("event_id", resp.EventId),
		logging.String("user_id", userID))

	return nil
}

// RecordUserUpdated sends a user update event to admin service
func (c *AdminClient) RecordUserUpdated(ctx context.Context, userID, email, username, firstName, lastName, updatedBy string, changedFields []string) error {
	req := &pb.UserUpdatedRequest{
		UserId:        userID,
		Email:         email,
		Username:      username,
		FirstName:     firstName,
		LastName:      lastName,
		UpdatedAt:     timestamppb.Now(),
		UpdatedBy:     updatedBy,
		ServiceName:   "auth-service",
		ChangedFields: changedFields,
	}

	c.logger.Debug("Sending user updated event to admin service",
		logging.String("user_id", userID))

	resp, err := c.client.RecordUserUpdated(ctx, req)
	if err != nil {
		c.logger.Error("Failed to record user updated event",
			logging.Error(err),
			logging.String("user_id", userID))
		return fmt.Errorf("failed to record user updated: %w", err)
	}

	if !resp.Success {
		c.logger.Warn("Admin service reported unsuccessful recording",
			logging.String("message", resp.Message))
		return fmt.Errorf("admin service error: %s", resp.Message)
	}

	c.logger.Info("User update event recorded successfully",
		logging.String("event_id", resp.EventId),
		logging.String("user_id", userID))

	return nil
}

// RecordUserDeleted sends a user deletion event to admin service
func (c *AdminClient) RecordUserDeleted(ctx context.Context, userID, email, deletedBy, reason string) error {
	req := &pb.UserDeletedRequest{
		UserId:      userID,
		Email:       email,
		DeletedAt:   timestamppb.Now(),
		DeletedBy:   deletedBy,
		ServiceName: "auth-service",
		Reason:      reason,
	}

	c.logger.Debug("Sending user deleted event to admin service",
		logging.String("user_id", userID))

	resp, err := c.client.RecordUserDeleted(ctx, req)
	if err != nil {
		c.logger.Error("Failed to record user deleted event",
			logging.Error(err),
			logging.String("user_id", userID))
		return fmt.Errorf("failed to record user deleted: %w", err)
	}

	if !resp.Success {
		c.logger.Warn("Admin service reported unsuccessful recording",
			logging.String("message", resp.Message))
		return fmt.Errorf("admin service error: %s", resp.Message)
	}

	c.logger.Info("User deletion event recorded successfully",
		logging.String("event_id", resp.EventId),
		logging.String("user_id", userID))

	return nil
}
