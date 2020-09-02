package todos

import (
	"fmt"
	"net/http"
	"strings"

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

// RedirectVersion sends the caller to the root of the current version and performs some
// lightweight content negotiation since it is the root of the application. Content
// negotiation is purposefully simple - if the Accept header starts with application/*
// then redirect to the API. If it starts with text/* then redirect to the web
// application. If the Accept header is */* then we default to the application endpoint
//(this is primarily an API). Otherwise, we respond with 406.
func (s *API) RedirectVersion(c *gin.Context) {
	accepts := negotiateFormat(c.GetHeader("Accept"))
	switch accepts {
	case "application", "*":
		c.Redirect(http.StatusPermanentRedirect, VersionURL())
	case "text":
		c.Redirect(http.StatusPermanentRedirect, "/app")
	default:
		c.AbortWithStatus(http.StatusNotAcceptable)
	}
}

// Returns "application", "text", "*", or "" by parsing the accept header and looking
// for the first matching acceptable content-type. If no acceptable content-type is
// available then an empty string is returned. This is a lightweight method that intends
// to capture as much as possible from the accept header.
func negotiateFormat(acceptHeader string) string {
	// Get accepts header and split on comma.
	accepts := strings.Split(acceptHeader, ",")
	for _, accept := range accepts {
		if accept = strings.TrimSpace(strings.Split(accept, ";")[0]); accept != "" {
			if accept = strings.ToLower(strings.Split(accept, "/")[0]); accept != "" {
				if accept == "application" || accept == "text" || accept == "*" {
					return accept
				}
			}
		}
	}
	return ""
}
