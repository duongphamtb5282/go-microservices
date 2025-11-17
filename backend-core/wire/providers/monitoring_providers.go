package providers

import (
	"backend-core/monitoring"
)

// MonitoringManagerProvider creates a monitoring manager
func MonitoringManagerProvider() *monitoring.MonitoringManager {
	return monitoring.NewMonitoringManager()
}
