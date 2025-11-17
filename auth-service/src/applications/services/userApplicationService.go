package services

import (
	"context"
	"fmt"
	"time"

	"auth-service/src/applications/commands"
	"auth-service/src/applications/dto"
	"auth-service/src/applications/queries"
	"auth-service/src/domain/events"
	"auth-service/src/domain/repositories"
	"auth-service/src/domain/services"
	"auth-service/src/domain/valueObjects"

	// "auth-service/src/infrastructure/telemetry" // Temporarily disabled
	"auth-service/src/infrastructure/worker"
	"backend-core/logging"
)

// UserApplicationService handles user application logic
type UserApplicationService struct {
	userRepo          repositories.UserRepository
	userCache         repositories.UserCache
	userDomainService *services.UserDomainService
	eventBus          EventBus
	adminClient       AdminClient
	workerPool        *worker.WorkerPool
	logger            *logging.Logger
	// telemetry         *telemetry.SimpleTelemetry // Temporarily disabled
	// businessMetrics   *telemetry.BusinessMetrics // Temporarily disabled
}

// AdminClient defines the interface for admin service gRPC client
type AdminClient interface {
	RecordUserCreated(ctx context.Context, userID, email, username, firstName, lastName, createdBy string) error
	RecordUserUpdated(ctx context.Context, userID, email, username, firstName, lastName, updatedBy string, changedFields []string) error
	RecordUserDeleted(ctx context.Context, userID, email, deletedBy, reason string) error
}

// EventBus defines the interface for publishing events
type EventBus interface {
	Publish(event interface{}) error
}

// NewUserApplicationService creates a new UserApplicationService
func NewUserApplicationService(
	userRepo repositories.UserRepository,
	userCache repositories.UserCache,
	userDomainService *services.UserDomainService,
	eventBus EventBus,
	adminClient AdminClient,
	workerPool *worker.WorkerPool,
	logger *logging.Logger,
	// telemetry *telemetry.SimpleTelemetry, // Temporarily disabled
	// businessMetrics *telemetry.BusinessMetrics, // Temporarily disabled
) *UserApplicationService {
	return &UserApplicationService{
		userRepo:          userRepo,
		userCache:         userCache,
		userDomainService: userDomainService,
		eventBus:          eventBus,
		adminClient:       adminClient,
		workerPool:        workerPool,
		logger:            logger,
		// telemetry:         telemetry, // Temporarily disabled
		// businessMetrics:   businessMetrics, // Temporarily disabled
	}
}

// CreateUser creates a new user with complete flow: DB -> Cache -> Kafka
func (s *UserApplicationService) CreateUser(ctx context.Context, cmd commands.CreateUserCommand) (*dto.CreateUserResponse, error) {
	// Temporarily disable telemetry
	// ctx, _ = s.telemetry.StartSpan(ctx, "user.create")
	// s.telemetry.SetSpanAttributes(ctx,
	// 	"user.username", cmd.Username,
	// 	"user.email", cmd.Email,
	// 	"operation", "create_user",
	// 	)

	// start := time.Now() // Temporarily disabled for telemetry
	s.logger.Info("Creating user with complete flow: DB -> Cache -> Kafka",
		logging.String("username", cmd.Username),
		logging.String("email", cmd.Email))

	// 1. Use domain service to create user (this saves to database)
	user, err := s.userDomainService.CreateUser(ctx, cmd.Username, cmd.Email, cmd.Password, cmd.CreatedBy)
	if err != nil {
		// s.telemetry.SetSpanError(ctx, err) // Temporarily disabled
		s.logger.Error("Failed to create user in database", "error", err)
		return nil, err
	}

	s.logger.Info("✅ User inserted into database", "user_id", user.ID().String(), "username", user.Username().String())

	// 2. Cache the user asynchronously
	cacheUserTask := &worker.Task{
		Type:       worker.TaskTypeCacheUser,
		Priority:   2, // Normal priority
		MaxRetries: 3,
		Timeout:    10 * time.Second,
		Payload: &worker.CacheUserPayload{
			UserID:   user.ID().String(),
			UserData: user,
		},
	}
	s.workerPool.SubmitTaskAsync(cacheUserTask)

	// Also cache by email for quick lookup asynchronously
	cacheUserByEmailTask := &worker.Task{
		Type:       worker.TaskTypeCacheUserByEmail,
		Priority:   2, // Normal priority
		MaxRetries: 3,
		Timeout:    10 * time.Second,
		Payload: &worker.CacheUserByEmailPayload{
			Email:    user.Email().String(),
			UserData: user,
		},
	}
	s.workerPool.SubmitTaskAsync(cacheUserByEmailTask)

	// 3. Publish domain event to Kafka asynchronously
	userCreatedEvent := events.NewUserCreated(user.ID(), user.Username(), user.Email())
	publishEventTask := &worker.Task{
		Type:       worker.TaskTypePublishEvent,
		Priority:   3, // High priority for events
		MaxRetries: 5, // More retries for critical events
		Timeout:    15 * time.Second,
		Payload: &worker.PublishEventPayload{
			EventType: "user.created",
			EventData: userCreatedEvent,
		},
	}
	s.workerPool.SubmitTaskAsync(publishEventTask)

	// 4. Record user creation in admin service via gRPC asynchronously
	if s.adminClient != nil {
		recordAdminTask := &worker.Task{
			Type:       worker.TaskTypeRecordAdminUser,
			Priority:   1, // Low priority for admin recording
			MaxRetries: 3,
			Timeout:    20 * time.Second,
			Payload: &worker.RecordAdminUserPayload{
				UserID:    user.ID().String(),
				Email:     user.Email().String(),
				Username:  user.Username().String(),
				FirstName: "", // firstName - add if available in your domain model
				LastName:  "", // lastName - add if available in your domain model
				CreatedBy: cmd.CreatedBy,
			},
		}
		s.workerPool.SubmitTaskAsync(recordAdminTask)
	}

	// Convert to DTO
	userDTO := dto.NewUserDTO(user)

	// Record business metrics
	// s.businessMetrics.RecordUserCreation(ctx, user.Username().String(), user.Email().String())

	s.logger.Info("✅ User created successfully - database insert completed, async tasks submitted",
		logging.String("user_id", user.ID().String()),
		logging.String("username", user.Username().String()),
		logging.String("async_tasks_submitted", "cache_user,cache_user_by_email,publish_event,record_admin"))

	// Temporarily disable span event
	// s.telemetry.AddSpanEvent(ctx, "user.created.successfully",
	// 	"user_id", user.ID().String(),
	// 	"duration_ms", time.Since(start).Seconds()*1000,
	// )

	return &dto.CreateUserResponse{
		User:    userDTO,
		Message: "User created successfully",
	}, nil
}

// GetUser retrieves a user by ID with cache-first strategy
func (s *UserApplicationService) GetUser(ctx context.Context, query queries.GetUserQuery) (*dto.GetUserResponse, error) {
	// Temporarily disable telemetry
	// ctx, _ = s.telemetry.StartSpan(ctx, "user.get")
	// s.telemetry.SetSpanAttributes(ctx,
	// 	"user.id", query.UserID,
	// 	"operation", "get_user",
	// 	)

	// start := time.Now() // Temporarily disabled for telemetry
	s.logger.Info("Getting user with cache-first strategy", "user_id", query.UserID)

	userID, err := valueObjects.NewUserIDFromString(query.UserID)
	if err != nil {
		// s.telemetry.SetSpanError(ctx, err) // Temporarily disabled
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// 1. Try cache first
	user, err := s.userCache.GetUser(ctx, userID)
	if err == nil {
		s.logger.Info("✅ User found in cache", "user_id", userID.String())

		userDTO := dto.NewUserDTO(user)
		return &dto.GetUserResponse{
			User: userDTO,
		}, nil
	}

	s.logger.Info("Cache miss, falling back to database", "user_id", userID.String())

	// 2. Fallback to database
	user, err = s.userRepo.FindByID(ctx, userID)
	if err != nil {
		// s.telemetry.SetSpanError(ctx, err) // Temporarily disabled
		s.logger.Error("Failed to get user from database", "error", err)
		return nil, fmt.Errorf("user not found")
	}

	// 3. Update cache with fresh data
	if err := s.userCache.CacheUser(ctx, user); err != nil {
		s.logger.Warn("Failed to update cache with user data", logging.Error(err))
		// Don't fail the operation if caching fails
	} else {
		s.logger.Info("✅ User cached successfully after database fetch", logging.String("user_id", userID.String()))
	}

	userDTO := dto.NewUserDTO(user)

	// Record business metrics
	// s.businessMetrics.RecordUserRetrieval(ctx, userID.String())

	s.logger.Info("✅ User retrieved successfully from database and cached", logging.String("user_id", userID.String()))

	// Temporarily disable span event
	// s.telemetry.AddSpanEvent(ctx, "user.retrieved.successfully",
	// 	"user_id", userID.String(),
	// 	"duration_ms", time.Since(start).Seconds()*1000,
	// 	"cache_hit", false,
	// )

	return &dto.GetUserResponse{
		User: userDTO,
	}, nil
}

// ActivateUser activates a user
func (s *UserApplicationService) ActivateUser(ctx context.Context, cmd commands.ActivateUserCommand) (*dto.ActivateUserResponse, error) {
	s.logger.Info("Activating user", logging.String("user_id", cmd.UserID))

	userID, err := valueObjects.NewUserIDFromString(cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	err = s.userDomainService.ActivateUser(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to activate user", logging.Error(err))
		return nil, err
	}

	// Get user for event
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user for event", logging.Error(err))
		return nil, err
	}

	// Publish domain event
	userActivatedEvent := events.NewUserActivated(user.ID(), user.Username(), user.Email())
	if err := s.eventBus.Publish(userActivatedEvent); err != nil {
		s.logger.Warn("Failed to publish user activated event", logging.Error(err))
	}

	s.logger.Info("User activated successfully", logging.String("user_id", cmd.UserID))

	return &dto.ActivateUserResponse{
		Message: "User activated successfully",
	}, nil
}

// ListUsers retrieves a list of users with pagination
func (s *UserApplicationService) ListUsers(ctx context.Context, page, limit int) (*dto.UserListResponse, error) {
	s.logger.Info("Listing users",
		logging.Int("page", page),
		logging.Int("limit", limit))

	offset := (page - 1) * limit

	users, err := s.userRepo.FindAll(ctx, offset, limit)
	if err != nil {
		s.logger.Error("Failed to list users", logging.Error(err))
		return nil, err
	}

	total, err := s.userRepo.Count(ctx)
	if err != nil {
		s.logger.Error("Failed to count users", logging.Error(err))
		return nil, err
	}

	userDTOs := make([]*dto.UserDTO, len(users))
	for i, user := range users {
		userDTOs[i] = dto.NewUserDTO(user)
	}

	return &dto.UserListResponse{
		Users: userDTOs,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}
