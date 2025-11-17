package logging

import (
	"context"
	"fmt"
	"time"

	grpcmiddleware "backend-core/grpc/middleware"
	"backend-core/logging"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoggingConfig holds configuration for logging interceptor
type LoggingConfig struct {
	// LogPayload determines if request/response payloads should be logged
	LogPayload bool
	// LogPayloadOnError determines if payloads should be logged only on errors
	LogPayloadOnError bool
}

// UnaryServerInterceptor returns a new unary server interceptor that logs requests
func UnaryServerInterceptor(logger *logging.Logger, config *LoggingConfig) grpc.UnaryServerInterceptor {
	if config == nil {
		config = &LoggingConfig{
			LogPayload:        false,
			LogPayloadOnError: true,
		}
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()

		// Extract correlation ID or generate one
		correlationID, ok := grpcmiddleware.GetCorrelationID(ctx)
		if !ok {
			md := grpcmiddleware.ExtractMetadata(ctx)
			if id, exists := md["x-correlation-id"]; exists {
				correlationID = id
				ctx = grpcmiddleware.WithCorrelationID(ctx, correlationID)
			}
		}

		// Extract peer info
		peerAddr := grpcmiddleware.ExtractPeerInfo(ctx)

		// Log request
		if userID, ok := grpcmiddleware.GetUserID(ctx); ok {
			logger.Info("gRPC request started",
				"grpc_method", info.FullMethod,
				"peer", peerAddr,
				"correlation_id", correlationID,
				"user_id", userID)
		} else {
			logger.Info("gRPC request started",
				"grpc_method", info.FullMethod,
				"peer", peerAddr,
				"correlation_id", correlationID)
		}

		// Call handler
		resp, err := handler(ctx, req)

		// Calculate duration
		duration := time.Since(startTime)

		// Determine status code
		st, _ := status.FromError(err)
		code := st.Code()

		// Log response
		if userID, ok := grpcmiddleware.GetUserID(ctx); ok {
			if err != nil {
				if config.LogPayloadOnError {
					logger.Error("gRPC request failed",
						"grpc_method", info.FullMethod,
						"grpc_code", code.String(),
						"duration", duration,
						"correlation_id", correlationID,
						"user_id", userID,
						"error", err.Error(),
						"request", req)
				} else {
					logger.Error("gRPC request failed",
						"grpc_method", info.FullMethod,
						"grpc_code", code.String(),
						"duration", duration,
						"correlation_id", correlationID,
						"user_id", userID,
						"error", err.Error())
				}
			} else {
				if config.LogPayload {
					logger.Info("gRPC request completed",
						"grpc_method", info.FullMethod,
						"grpc_code", code.String(),
						"duration", duration,
						"correlation_id", correlationID,
						"user_id", userID,
						"response", resp)
				} else {
					logger.Info("gRPC request completed",
						"grpc_method", info.FullMethod,
						"grpc_code", code.String(),
						"duration", duration,
						"correlation_id", correlationID,
						"user_id", userID)
				}
			}
		} else {
			if err != nil {
				if config.LogPayloadOnError {
					logger.Error("gRPC request failed",
						"grpc_method", info.FullMethod,
						"grpc_code", code.String(),
						"duration", duration,
						"correlation_id", correlationID,
						"error", err.Error(),
						"request", req)
				} else {
					logger.Error("gRPC request failed",
						"grpc_method", info.FullMethod,
						"grpc_code", code.String(),
						"duration", duration,
						"correlation_id", correlationID,
						"error", err.Error())
				}
			} else {
				if config.LogPayload {
					logger.Info("gRPC request completed",
						"grpc_method", info.FullMethod,
						"grpc_code", code.String(),
						"duration", duration,
						"correlation_id", correlationID,
						"response", resp)
				} else {
					logger.Info("gRPC request completed",
						"grpc_method", info.FullMethod,
						"grpc_code", code.String(),
						"duration", duration,
						"correlation_id", correlationID)
				}
			}
		}

		return resp, err
	}
}

// StreamServerInterceptor returns a new stream server interceptor that logs streams
func StreamServerInterceptor(logger *logging.Logger, config *LoggingConfig) grpc.StreamServerInterceptor {
	if config == nil {
		config = &LoggingConfig{
			LogPayload:        false,
			LogPayloadOnError: true,
		}
	}

	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		startTime := time.Now()
		ctx := ss.Context()

		// Extract correlation ID
		correlationID, ok := grpcmiddleware.GetCorrelationID(ctx)
		if !ok {
			md := grpcmiddleware.ExtractMetadata(ctx)
			if id, exists := md["x-correlation-id"]; exists {
				correlationID = id
			}
		}

		// Extract peer info
		peerAddr := grpcmiddleware.ExtractPeerInfo(ctx)

		// Log stream start
		if userID, ok := grpcmiddleware.GetUserID(ctx); ok {
			logger.Info("gRPC stream started",
				"grpc_method", info.FullMethod,
				"peer", peerAddr,
				"correlation_id", correlationID,
				"is_client_stream", info.IsClientStream,
				"is_server_stream", info.IsServerStream,
				"user_id", userID)
		} else {
			logger.Info("gRPC stream started",
				"grpc_method", info.FullMethod,
				"peer", peerAddr,
				"correlation_id", correlationID,
				"is_client_stream", info.IsClientStream,
				"is_server_stream", info.IsServerStream)
		}

		// Call handler
		err := handler(srv, ss)

		// Calculate duration
		duration := time.Since(startTime)

		// Determine status code
		st, _ := status.FromError(err)
		code := st.Code()

		// Log stream end
		if userID, ok := grpcmiddleware.GetUserID(ctx); ok {
			if err != nil {
				logger.Error("gRPC stream failed",
					"grpc_method", info.FullMethod,
					"grpc_code", code.String(),
					"duration", duration,
					"correlation_id", correlationID,
					"user_id", userID,
					"error", err.Error())
			} else {
				logger.Info("gRPC stream completed",
					"grpc_method", info.FullMethod,
					"grpc_code", code.String(),
					"duration", duration,
					"correlation_id", correlationID,
					"user_id", userID)
			}
		} else {
			if err != nil {
				logger.Error("gRPC stream failed",
					"grpc_method", info.FullMethod,
					"grpc_code", code.String(),
					"duration", duration,
					"correlation_id", correlationID,
					"error", err.Error())
			} else {
				logger.Info("gRPC stream completed",
					"grpc_method", info.FullMethod,
					"grpc_code", code.String(),
					"duration", duration,
					"correlation_id", correlationID)
			}
		}

		return err
	}
}

// UnaryClientInterceptor returns a new unary client interceptor that logs requests
func UnaryClientInterceptor(logger *logging.Logger) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		startTime := time.Now()

		// Extract correlation ID
		correlationID, _ := grpcmiddleware.GetCorrelationID(ctx)

		logger.Info("gRPC client request started",
			logging.String("grpc_method", method),
			logging.String("correlation_id", correlationID),
			logging.String("target", cc.Target()))

		// Call invoker
		err := invoker(ctx, method, req, reply, cc, opts...)

		// Calculate duration
		duration := time.Since(startTime)

		// Determine status code
		st, _ := status.FromError(err)
		code := st.Code()

		if err != nil {
			logger.Error("gRPC client request failed",
				logging.String("grpc_method", method),
				logging.String("grpc_code", code.String()),
				logging.Duration("duration", duration),
				logging.String("correlation_id", correlationID),
				logging.Error(err))
		} else {
			logger.Info("gRPC client request completed",
				logging.String("grpc_method", method),
				logging.String("grpc_code", code.String()),
				logging.Duration("duration", duration),
				logging.String("correlation_id", correlationID))
		}

		return err
	}
}

// ExtractFields extracts fields from context for logging
func ExtractFields(ctx context.Context) []logging.Field {
	fields := []logging.Field{}

	if userID, ok := grpcmiddleware.GetUserID(ctx); ok {
		fields = append(fields, logging.String("user_id", userID))
	}

	if username, ok := grpcmiddleware.GetUsername(ctx); ok {
		fields = append(fields, logging.String("username", username))
	}

	if correlationID, ok := grpcmiddleware.GetCorrelationID(ctx); ok {
		fields = append(fields, logging.String("correlation_id", correlationID))
	}

	if clientIP, ok := grpcmiddleware.GetClientIP(ctx); ok {
		fields = append(fields, logging.String("client_ip", clientIP))
	}

	return fields
}

// CodeToLevel maps gRPC codes to log levels
func CodeToLevel(code codes.Code) string {
	switch code {
	case codes.OK:
		return "info"
	case codes.Canceled:
		return "info"
	case codes.Unknown:
		return "error"
	case codes.InvalidArgument:
		return "warn"
	case codes.DeadlineExceeded:
		return "warn"
	case codes.NotFound:
		return "warn"
	case codes.AlreadyExists:
		return "warn"
	case codes.PermissionDenied:
		return "warn"
	case codes.Unauthenticated:
		return "warn"
	case codes.ResourceExhausted:
		return "warn"
	case codes.FailedPrecondition:
		return "warn"
	case codes.Aborted:
		return "warn"
	case codes.OutOfRange:
		return "warn"
	case codes.Unimplemented:
		return "error"
	case codes.Internal:
		return "error"
	case codes.Unavailable:
		return "warn"
	case codes.DataLoss:
		return "error"
	default:
		return "error"
	}
}

// DefaultMessageProducer produces a default log message
func DefaultMessageProducer(ctx context.Context, msg string, level string, code codes.Code, err error, duration time.Duration) string {
	return fmt.Sprintf("%s [%s] duration=%s", msg, code.String(), duration)
}
