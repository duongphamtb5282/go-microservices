package grpc

import (
	"context"
	"fmt"
	"time"

	"admin-service/src/applications/services"
	"admin-service/src/domain"
	"backend-core/logging"
	pb "backend-shared/proto/admin"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AdminServer implements the gRPC AdminService
type AdminServer struct {
	pb.UnimplementedAdminServiceServer
	userEventService *services.UserEventService
	logger           *logging.Logger
}

// NewAdminServer creates a new admin gRPC server
func NewAdminServer(userEventService *services.UserEventService, logger *logging.Logger) *AdminServer {
	return &AdminServer{
		userEventService: userEventService,
		logger:           logger,
	}
}

// RecordUserCreated records a user creation event
func (s *AdminServer) RecordUserCreated(ctx context.Context, req *pb.UserCreatedRequest) (*pb.UserCreatedResponse, error) {
	s.logger.Info("Received RecordUserCreated request",
		logging.String("user_id", req.UserId),
		logging.String("email", req.Email),
		logging.String("service_name", req.ServiceName))

	// Parse user ID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		s.logger.Error("Invalid user ID", logging.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	// Determine performed by
	performedBy := req.CreatedBy
	if performedBy == "" {
		performedBy = "system"
	}

	// Record the event
	event, err := s.userEventService.RecordUserCreated(
		ctx,
		userID,
		req.Email,
		req.Username,
		req.FirstName,
		req.LastName,
		req.ServiceName,
		performedBy,
	)
	if err != nil {
		s.logger.Error("Failed to record user created event", logging.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to record event: %v", err)
	}

	return &pb.UserCreatedResponse{
		EventId:    event.ID.String(),
		UserId:     event.UserID.String(),
		Success:    true,
		Message:    "User creation event recorded successfully",
		RecordedAt: timestamppb.New(event.CreatedAt),
	}, nil
}

// RecordUserUpdated records a user update event
func (s *AdminServer) RecordUserUpdated(ctx context.Context, req *pb.UserUpdatedRequest) (*pb.UserUpdatedResponse, error) {
	s.logger.Info("Received RecordUserUpdated request",
		logging.String("user_id", req.UserId),
		logging.String("email", req.Email))

	// Parse user ID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		s.logger.Error("Invalid user ID", logging.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	// Determine performed by
	performedBy := req.UpdatedBy
	if performedBy == "" {
		performedBy = "system"
	}

	// Record the event
	event, err := s.userEventService.RecordUserUpdated(
		ctx,
		userID,
		req.Email,
		req.Username,
		req.FirstName,
		req.LastName,
		req.ServiceName,
		performedBy,
		req.ChangedFields,
	)
	if err != nil {
		s.logger.Error("Failed to record user updated event", logging.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to record event: %v", err)
	}

	return &pb.UserUpdatedResponse{
		EventId:    event.ID.String(),
		UserId:     event.UserID.String(),
		Success:    true,
		Message:    "User update event recorded successfully",
		RecordedAt: timestamppb.New(event.CreatedAt),
	}, nil
}

// RecordUserDeleted records a user deletion event
func (s *AdminServer) RecordUserDeleted(ctx context.Context, req *pb.UserDeletedRequest) (*pb.UserDeletedResponse, error) {
	s.logger.Info("Received RecordUserDeleted request",
		logging.String("user_id", req.UserId),
		logging.String("email", req.Email))

	// Parse user ID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		s.logger.Error("Invalid user ID", logging.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	// Determine performed by
	performedBy := req.DeletedBy
	if performedBy == "" {
		performedBy = "system"
	}

	// Record the event
	event, err := s.userEventService.RecordUserDeleted(
		ctx,
		userID,
		req.Email,
		req.ServiceName,
		performedBy,
		req.Reason,
	)
	if err != nil {
		s.logger.Error("Failed to record user deleted event", logging.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to record event: %v", err)
	}

	return &pb.UserDeletedResponse{
		EventId:    event.ID.String(),
		UserId:     event.UserID.String(),
		Success:    true,
		Message:    "User deletion event recorded successfully",
		RecordedAt: timestamppb.New(event.CreatedAt),
	}, nil
}

// GetUserEvents retrieves user events
func (s *AdminServer) GetUserEvents(ctx context.Context, req *pb.GetUserEventsRequest) (*pb.GetUserEventsResponse, error) {
	s.logger.Info("Received GetUserEvents request",
		logging.String("user_id", req.UserId),
		logging.String("event_type", req.EventType))

	// Parse user ID if provided
	var userID uuid.UUID
	var err error
	if req.UserId != "" {
		userID, err = uuid.Parse(req.UserId)
		if err != nil {
			s.logger.Error("Invalid user ID", logging.Error(err))
			return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
		}
	}

	// Parse event type if provided
	var eventType domain.EventType
	if req.EventType != "" {
		eventType = domain.EventType(req.EventType)
	}

	// Parse dates if provided
	var fromDate, toDate time.Time
	if req.FromDate != nil {
		fromDate = req.FromDate.AsTime()
	}
	if req.ToDate != nil {
		toDate = req.ToDate.AsTime()
	}

	// Set pagination defaults
	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get events
	events, total, err := s.userEventService.GetUserEvents(
		ctx,
		userID,
		eventType,
		fromDate,
		toDate,
		page,
		pageSize,
	)
	if err != nil {
		s.logger.Error("Failed to get user events", logging.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get events: %v", err)
	}

	// Convert to proto events
	pbEvents := make([]*pb.UserEvent, len(events))
	for i, event := range events {
		// Convert metadata to map[string]string
		metadata := make(map[string]string)
		for k, v := range event.Metadata {
			if str, ok := v.(string); ok {
				metadata[k] = str
			} else {
				metadata[k] = fmt.Sprint(v)
			}
		}

		pbEvents[i] = &pb.UserEvent{
			EventId:     event.ID.String(),
			UserId:      event.UserID.String(),
			EventType:   string(event.EventType),
			Email:       event.Email,
			Username:    event.Username,
			ServiceName: event.ServiceName,
			PerformedBy: event.PerformedBy,
			EventTime:   timestamppb.New(event.EventTime),
			Metadata:    metadata,
		}
	}

	return &pb.GetUserEventsResponse{
		Events:     pbEvents,
		TotalCount: int32(total),
		Page:       int32(page),
		PageSize:   int32(pageSize),
	}, nil
}
