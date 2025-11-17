package interfaces

import (
	"context"
	"time"

	"backend-core/database/health"
)

// DatabaseHealthChecker defines the health checking interface
type DatabaseHealthChecker interface {
	CheckHealth(ctx context.Context) error
	GetStatus() health.HealthStatus
	GetLastHealthCheck() time.Time
	GetHealthHistory() []health.HealthCheck
}
