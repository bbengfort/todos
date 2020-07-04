package client

import (
	"fmt"
	"net/http"

	"github.com/bbengfort/todos"
)

// Login to the todos API, saving the access tokens to disk for use during other
// sessions. If the password is in the credentials, login executes directly, otherwise
// it prompts the user for the password. This is not a standard API client request, e.g.
// it does not take a LoginRequest and return a LoginResponse. Instead this method
// entirely manages the login process on behalf of the user.
func (c *Client) Login() (err error) {
	// If we're already logged in, return an error (must logout first)
	if c.creds.IsLoggedIn() {
		return ErrLoggedIn
	}

	// Build data request
	data := &todos.LoginRequest{
		Username: c.creds.Username,
		Password: c.creds.Password,
		NoCookie: true,
	}

	if data.Username == "" {
		data.Username = Prompt("username", "")
	}

	if data.Password == "" {
		if data.Password, err = PromptPassword("password", true, false); err != nil {
			return err
		}
	}

	// Execute the data request
	var req *http.Request
	if req, err = c.NewRequest(http.MethodPost, "/login", false, data); err != nil {
		return err
	}

	var status int
	var tokens *todos.LoginResponse
	if status, err = c.Do(req, &tokens); err != nil {
		return err
	}

	// Handle the error if we don't get an ok or a success message
	if status != http.StatusOK || !tokens.Success {
		return StatusError(status, tokens.Error)
	}

	// Set the tokens on the credentials and save them to disk
	if err = c.creds.SetTokens(tokens); err != nil {
		return err
	}
	return nil
}

// Logout issues a logout request to the server then clears cached tokens locally.
// If revokeAll is true, then the server will remove all outstanding tokens, not just
// the token posted by the current client. If the logout succeeds, then the cached
// tokens are revoked, but they are not deleted if the request fails.
func (c *Client) Logout(revokeAll bool) (err error) {
	if !c.creds.IsLoggedIn() {
		if c.creds.IsRefreshable() {
			// We have to refresh the access token in order to perform the log out
			if err = c.Refresh(); err != nil {
				return fmt.Errorf("could not refresh token to log it out: %s", err)
			}
		} else {
			// Can't log out if we aren't logged in
			return ErrNotLoggedIn
		}

	}

	// Create the logout request
	data := &todos.LogoutRequest{
		RevokeAll: revokeAll,
	}

	var req *http.Request
	if req, err = c.NewRequest(http.MethodPost, "/logout", true, data); err != nil {
		return err
	}

	// Execute the logout request
	var (
		status int
		rep    *todos.Response
	)
	if status, err = c.Do(req, &rep); err != nil {
		return err
	}

	// If a bad status code is given, then return an error
	if !(status == http.StatusOK || status == http.StatusNoContent || status == http.StatusUnauthorized) || !rep.Success {
		if rep.Error == "" {
			rep.Error = "could not logout the user"
		}
		return StatusError(status, rep.Error)
	}

	// Revoke certificates from the credentials
	if err = c.creds.Revoke(); err != nil {
		return err
	}
	return nil
}

// Refresh uses the refresh token to get a new access token without having to login.
func (c *Client) Refresh() (err error) {
	if !c.creds.IsRefreshable() {
		return ErrNotRefreshable
	}

	var (
		status int
		req    *http.Request
		tokens *todos.LoginResponse
	)

	// Build request
	data := &todos.RefreshRequest{
		RefreshToken: c.creds.Tokens.Refresh,
		NoCookie:     true,
	}
	if req, err = c.NewRequest(http.MethodPost, "/refresh", false, data); err != nil {
		return err
	}

	// Execute the request
	if status, err = c.Do(req, tokens); err != nil {
		return err
	}

	if status != http.StatusOK || !tokens.Success {
		return StatusError(status, tokens.Error)
	}

	// Set the tokens on the credentials and save them to disk
	if err = c.creds.SetTokens(tokens); err != nil {
		return err
	}
	return nil
}

// CheckLogin ensures that the user is ready to make an authenticated request by
// verifying that a non-expired access token exists. If the access token is expired but
// the refresh token is not, it refreshes the token automatically. Otherwise, it runs
// the login command to get an access token (which may prompt the user for a password).
func (c *Client) CheckLogin() (err error) {
	if !c.creds.IsLoggedIn() {
		if c.creds.IsRefreshable() {
			return c.Refresh()
		}
		return c.Login()
	}
	return nil
}
