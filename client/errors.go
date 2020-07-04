package client

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Errors for standardized error handling
var (
	ErrNoConfiguration = errors.New("no configuration directory created yet")
	ErrNoConfDir       = errors.New("could not find a directory to write credentials to")
	ErrNoCredentials   = errors.New("could not find credentials, run configure")
	ErrNoEndpoint      = errors.New("no endpoint specified in credentials, run configure")
	ErrLoggedIn        = errors.New("already logged in, logout before logging in again")
	ErrNotLoggedIn     = errors.New("not logged in, run the login command first")
	ErrNotRefreshable  = errors.New("cannot refresh tokens, please login again")
)

// StatusError creates an error for the status and the error text in the reply or uses
// the normal http status text for the error message if needed.
func StatusError(status int, text string) error {
	if text != "" {
		return fmt.Errorf("[%d] %s", status, text)
	}
	return fmt.Errorf("[%d] %s", status, strings.ToLower(http.StatusText(status)))
}
