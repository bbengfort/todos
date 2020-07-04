package todos

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Version components for detailed version helpers
const (
	VersionMajor  = 1
	VersionMinor  = 2
	VersionPatch  = 0
	VersionSerial = 3
)

// Version returns the human readable version of the package
func Version() string {
	if VersionPatch > 0 {
		return fmt.Sprintf("%d.%d.%d", VersionMajor, VersionMinor, VersionPatch)
	}
	return fmt.Sprintf("%d.%d", VersionMajor, VersionMinor)
}

// VersionURL returns the URL prefix for the API at the current version
func VersionURL() string {
	return fmt.Sprintf("/v%d", VersionMajor)
}

// RedirectVersion sends the caller to the root of the current version
func (s *API) RedirectVersion(c *gin.Context) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Redirect(http.StatusPermanentRedirect, VersionURL())
}
