package providers

import (
	"auth-service/src/domain/repositories"
	domainServices "auth-service/src/domain/services"
	"backend-core/logging"
)

// UserDomainServiceProvider creates a user domain service
func UserDomainServiceProvider(userRepo repositories.UserRepository, logger *logging.Logger) *domainServices.UserDomainService {
	return domainServices.NewUserDomainService(userRepo)
}
