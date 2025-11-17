package providers

import (
	"backend-core/security"
)

// SecurityManagerProvider creates a security manager
func SecurityManagerProvider() *security.SecurityManager {
	return security.NewSecurityManager()
}
