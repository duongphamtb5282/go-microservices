package wire

import (
	"backend-core/logging"
	"backend-core/wire/composition"
)

// InitializeCoreInfrastructure creates the core infrastructure with proper dependency injection
func InitializeCoreInfrastructure() *composition.CoreComposition {
	// Use composition to create the core infrastructure
	return composition.ComposeCoreInfrastructure()
}

// GetLogger creates a logger for external services
func GetLogger() *logging.Logger {
	coreComposition := InitializeCoreInfrastructure()
	return coreComposition.GetLogger()
}
