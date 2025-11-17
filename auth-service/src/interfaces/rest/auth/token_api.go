package auth

import (
	"auth-service/src/interfaces/rest/handlers"
	"backend-core/logging"

	"github.com/gin-gonic/gin"
)

type TokenApi struct {
	tokenHandler *handlers.TokenHandler
	logger       *logging.Logger
}

func NewTokenApi(tokenHandler *handlers.TokenHandler, logger *logging.Logger) *TokenApi {
	return &TokenApi{
		tokenHandler: tokenHandler,
		logger:       logger,
	}
}

// GenerateToken handles token generation
func (t *TokenApi) GenerateToken(c *gin.Context) {
	t.tokenHandler.GenerateToken(c)
}

// RevokeToken handles token revocation
func (t *TokenApi) RevokeToken(c *gin.Context) {
	t.tokenHandler.RevokeToken(c)
}

// RevokeAllTokens handles revoking all user tokens
func (t *TokenApi) RevokeAllTokens(c *gin.Context) {
	t.tokenHandler.RevokeAllTokens(c)
}

// ValidateToken handles token validation
func (t *TokenApi) ValidateToken(c *gin.Context) {
	t.tokenHandler.ValidateToken(c)
}

// GetTokenList handles getting user token list
func (t *TokenApi) GetTokenList(c *gin.Context) {
	t.tokenHandler.GetTokenList(c)
}

// GetTokenInfo handles getting token information
func (t *TokenApi) GetTokenInfo(c *gin.Context) {
	t.tokenHandler.GetTokenInfo(c)
}

// GetTokenStats handles getting token statistics
func (t *TokenApi) GetTokenStats(c *gin.Context) {
	t.tokenHandler.GetTokenStats(c)
}
