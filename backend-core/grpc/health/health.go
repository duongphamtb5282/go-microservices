package health

import (
	"context"
	"sync"

	"google.golang.org/grpc/health/grpc_health_v1"
)

// HealthServer implements the gRPC health checking protocol
type HealthServer struct {
	grpc_health_v1.UnimplementedHealthServer
	mu        sync.RWMutex
	statusMap map[string]grpc_health_v1.HealthCheckResponse_ServingStatus
}

// NewHealthServer creates a new health server
func NewHealthServer() *HealthServer {
	return &HealthServer{
		statusMap: make(map[string]grpc_health_v1.HealthCheckResponse_ServingStatus),
	}
}

// Check implements the health check
func (s *HealthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	service := req.GetService()
	status, ok := s.statusMap[service]
	if !ok {
		status = grpc_health_v1.HealthCheckResponse_SERVICE_UNKNOWN
	}

	return &grpc_health_v1.HealthCheckResponse{
		Status: status,
	}, nil
}

// Watch implements the health watch (server streaming)
func (s *HealthServer) Watch(req *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
	service := req.GetService()

	// Send initial status
	s.mu.RLock()
	status, ok := s.statusMap[service]
	s.mu.RUnlock()

	if !ok {
		status = grpc_health_v1.HealthCheckResponse_SERVICE_UNKNOWN
	}

	if err := stream.Send(&grpc_health_v1.HealthCheckResponse{Status: status}); err != nil {
		return err
	}

	// Keep connection alive (simplified implementation)
	<-stream.Context().Done()
	return stream.Context().Err()
}

// SetServingStatus sets the serving status of a service
func (s *HealthServer) SetServingStatus(service string, status grpc_health_v1.HealthCheckResponse_ServingStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.statusMap[service] = status
}

// SetServingStatusForAll sets the serving status for all services
func (s *HealthServer) SetServingStatusForAll(status grpc_health_v1.HealthCheckResponse_ServingStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for service := range s.statusMap {
		s.statusMap[service] = status
	}
}

// Shutdown marks all services as not serving
func (s *HealthServer) Shutdown() {
	s.SetServingStatusForAll(grpc_health_v1.HealthCheckResponse_NOT_SERVING)
}

// Resume marks all services as serving
func (s *HealthServer) Resume() {
	s.SetServingStatusForAll(grpc_health_v1.HealthCheckResponse_SERVING)
}
