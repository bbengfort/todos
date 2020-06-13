package todos

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Status is an unauthenticated endpoint that returns the status of the api server and
// can be used for heartbeats and liveness checks.
func (s *API) Status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339Nano),
		"version":   Version,
	})
}

// Available is middleware that uses the healthy boolean to return a service unavailable
// http status code if the server is shutting down. It does this before all routes to
// ensure that complex handling doesn't bog down the server.
func (s *API) Available() gin.HandlerFunc {
	return func(c *gin.Context) {
		s.RLock()
		healthy := s.healthy
		s.RUnlock()

		if !healthy {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"success": false, "error": "service unavailable",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
