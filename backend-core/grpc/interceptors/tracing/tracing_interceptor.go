package tracing

import (
	"context"

	grpcmiddleware "backend-core/grpc/middleware"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryServerInterceptor returns a new unary server interceptor for tracing
// This is a simplified version - use otelgrpc for full OpenTelemetry support
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Extract trace context from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			if traceID := md.Get("x-trace-id"); len(traceID) > 0 {
				ctx = grpcmiddleware.WithCorrelationID(ctx, traceID[0])
			}
		}

		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new stream server interceptor for tracing
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()

		// Extract trace context from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			if traceID := md.Get("x-trace-id"); len(traceID) > 0 {
				ctx = grpcmiddleware.WithCorrelationID(ctx, traceID[0])
			}
		}

		return handler(srv, ss)
	}
}

// UnaryClientInterceptor returns a new unary client interceptor for tracing
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Propagate trace context
		if correlationID, ok := grpcmiddleware.GetCorrelationID(ctx); ok {
			ctx = metadata.AppendToOutgoingContext(ctx, "x-trace-id", correlationID)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StreamClientInterceptor returns a new stream client interceptor for tracing
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		// Propagate trace context
		if correlationID, ok := grpcmiddleware.GetCorrelationID(ctx); ok {
			ctx = metadata.AppendToOutgoingContext(ctx, "x-trace-id", correlationID)
		}

		return streamer(ctx, desc, cc, method, opts...)
	}
}
