package metrics

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// MetricsReporter interface for reporting metrics
type MetricsReporter interface {
	// RecordRequest records a gRPC request
	RecordRequest(method string, code string, duration time.Duration)
	// RecordMessageSent records a message sent
	RecordMessageSent(method string)
	// RecordMessageReceived records a message received
	RecordMessageReceived(method string)
}

// UnaryServerInterceptor returns a new unary server interceptor for metrics
func UnaryServerInterceptor(reporter MetricsReporter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()

		// Record message received
		reporter.RecordMessageReceived(info.FullMethod)

		// Call handler
		resp, err := handler(ctx, req)

		// Calculate duration
		duration := time.Since(startTime)

		// Get status code
		st, _ := status.FromError(err)
		code := st.Code().String()

		// Record request
		reporter.RecordRequest(info.FullMethod, code, duration)

		// Record message sent if successful
		if err == nil {
			reporter.RecordMessageSent(info.FullMethod)
		}

		return resp, err
	}
}

// StreamServerInterceptor returns a new stream server interceptor for metrics
func StreamServerInterceptor(reporter MetricsReporter) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		startTime := time.Now()

		// Wrap stream to count messages
		wrapped := &metricsServerStream{
			ServerStream: ss,
			reporter:     reporter,
			method:       info.FullMethod,
		}

		// Call handler
		err := handler(srv, wrapped)

		// Calculate duration
		duration := time.Since(startTime)

		// Get status code
		st, _ := status.FromError(err)
		code := st.Code().String()

		// Record request
		reporter.RecordRequest(info.FullMethod, code, duration)

		return err
	}
}

// metricsServerStream wraps grpc.ServerStream to count messages
type metricsServerStream struct {
	grpc.ServerStream
	reporter MetricsReporter
	method   string
}

func (s *metricsServerStream) SendMsg(m interface{}) error {
	err := s.ServerStream.SendMsg(m)
	if err == nil {
		s.reporter.RecordMessageSent(s.method)
	}
	return err
}

func (s *metricsServerStream) RecvMsg(m interface{}) error {
	err := s.ServerStream.RecvMsg(m)
	if err == nil {
		s.reporter.RecordMessageReceived(s.method)
	}
	return err
}

// UnaryClientInterceptor returns a new unary client interceptor for metrics
func UnaryClientInterceptor(reporter MetricsReporter) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		startTime := time.Now()

		// Record message sent
		reporter.RecordMessageSent(method)

		// Call invoker
		err := invoker(ctx, method, req, reply, cc, opts...)

		// Calculate duration
		duration := time.Since(startTime)

		// Get status code
		st, _ := status.FromError(err)
		code := st.Code().String()

		// Record request
		reporter.RecordRequest(method, code, duration)

		// Record message received if successful
		if err == nil {
			reporter.RecordMessageReceived(method)
		}

		return err
	}
}

// NoOpReporter is a no-op metrics reporter
type NoOpReporter struct{}

func (r *NoOpReporter) RecordRequest(method string, code string, duration time.Duration) {}
func (r *NoOpReporter) RecordMessageSent(method string)                                  {}
func (r *NoOpReporter) RecordMessageReceived(method string)                              {}
