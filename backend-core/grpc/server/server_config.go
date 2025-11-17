package server

import "time"

// ServerConfig holds gRPC server configuration
type ServerConfig struct {
	// Host is the server host
	Host string
	// Port is the server port
	Port string
	// MaxMessageSize is the maximum message size in bytes
	MaxMessageSize int
	// ConnectionTimeout is the connection timeout
	ConnectionTimeout time.Duration
	// KeepaliveTime is the keepalive time
	KeepaliveTime time.Duration
	// KeepaliveTimeout is the keepalive timeout
	KeepaliveTimeout time.Duration
	// MaxConcurrentStreams is the maximum concurrent streams
	MaxConcurrentStreams uint32
}

// DefaultServerConfig returns default server configuration
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Host:                 "0.0.0.0",
		Port:                 "50051",
		MaxMessageSize:       10 * 1024 * 1024, // 10MB
		ConnectionTimeout:    120 * time.Second,
		KeepaliveTime:        30 * time.Second,
		KeepaliveTimeout:     10 * time.Second,
		MaxConcurrentStreams: 100,
	}
}
