package todos

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NotFound returns a JSON 404 response for the API.
func NotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "resource not found"})
}

// NotAllowed returns a JSON 405 response for the API.
func NotAllowed(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, gin.H{"success": false, "message": "method not allowed"})
}
