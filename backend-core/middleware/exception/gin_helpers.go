package exception

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GinErrorResponse sends a standardized error response using Gin
func GinErrorResponse(c *gin.Context, err error, statusCode int) {
	errorResponse := NewErrorResponse(
		http.StatusText(statusCode),
		err.Error(),
		"validation_error",
		map[string]interface{}{
			"field": "request",
		},
	)

	c.JSON(statusCode, errorResponse)
}

// GinSuccessResponse sends a standardized success response using Gin
func GinSuccessResponse(c *gin.Context, data interface{}, message string) {
	response := map[string]interface{}{
		"success": true,
		"message": message,
		"data":    data,
	}

	c.JSON(http.StatusOK, response)
}
