package todos

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Context keys for middleware lookups
const (
	ctxUserKey = "user"
)

// Overview returns the state of the todos (e.g. the number of tasks and lists for the
// given user). This request must be authenticated.
func (s *API) Overview(c *gin.Context) {
	user := c.Value(ctxUserKey).(User)
	c.JSON(http.StatusOK, gin.H{"tasks": 0, "lists": 0, "user": user.Username})
}
