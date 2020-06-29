package client

import "errors"

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
