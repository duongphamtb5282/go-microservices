package telemetry

import (
	"os"
	"strconv"
)

// DefaultTelemetryConfig returns default telemetry configuration
func DefaultTelemetryConfig() TelemetryConfig {
	return TelemetryConfig{
		ServiceName:    getEnv("OTEL_SERVICE_NAME", "backend-service"),
		ServiceVersion: getEnv("OTEL_SERVICE_VERSION", "1.0.0"),
		Environment:    getEnv("OTEL_ENVIRONMENT", "development"),
		JaegerEndpoint: getEnv("OTEL_EXPORTER_JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
		OTLPEndpoint:   getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"),
		PrometheusPort: getEnv("OTEL_PROMETHEUS_PORT", "8888"),
		Enabled:        getEnvBool("OTEL_ENABLED", true),
	}
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool gets boolean environment variable with default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// TelemetryConfigFromEnv creates telemetry config from environment variables
func TelemetryConfigFromEnv() TelemetryConfig {
	return TelemetryConfig{
		ServiceName:    getEnv("OTEL_SERVICE_NAME", "backend-service"),
		ServiceVersion: getEnv("OTEL_SERVICE_VERSION", "1.0.0"),
		Environment:    getEnv("OTEL_ENVIRONMENT", "development"),
		JaegerEndpoint: getEnv("OTEL_EXPORTER_JAEGER_ENDPOINT", ""),
		OTLPEndpoint:   getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		PrometheusPort: getEnv("OTEL_PROMETHEUS_PORT", "8888"),
		Enabled:        getEnvBool("OTEL_ENABLED", true),
	}
}
