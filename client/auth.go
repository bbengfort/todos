package client

import (
	"fmt"
	"net/http"
)

// Login to the todos API, saving the access tokens to disk for use during other
// sessions. If the password is in the credentials, login executes directly, otherwise
// it prompts the user for the password.
func (c *Client) Login() (err error) {
	// If we're already logged in, return an error (must logout first)
	if c.creds.IsLoggedIn() {
		return ErrLoggedIn
	}

	// Build data request
	data := make(map[string]interface{})
	data["username"] = c.creds.Username
	data["password"] = c.creds.Password
	data["no_cookie"] = true

	if c.creds.Username == "" {
		data["username"] = Prompt("username", "")
	}

	if c.creds.Password == "" {
		if data["password"], err = PromptPassword("password", true, false); err != nil {
			return err
		}
	}

	// Execute the data request
	var req *http.Request
	if req, err = c.NewRequest(http.MethodPost, "/login", false, data); err != nil {
		return err
	}

	var status int
	var tokens map[string]interface{}
	if tokens, status, err = c.Do(req); err != nil {
		return err
	}

	if status != http.StatusOK {
		// TODO: better error handling here
		return fmt.Errorf("got a %s status", http.StatusText(status))
	}

	// Set the tokens on the credentials and save them to disk
	if err = c.creds.SetTokens(tokens); err != nil {
		return err
	}
	return nil
}

// Logout issues a logout request to the server then clears cached tokens locally.
// If revokeAll is true, then the server will remove all outstanding tokens, not just
// the token posted by the current client.
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
	data := make(map[string]interface{})
	data["revoke_all"] = revokeAll

	var req *http.Request
	if req, err = c.NewRequest(http.MethodPost, "/logout", true, data); err != nil {
		return err
	}

	// Execute the logout request
	var status int
	if _, status, err = c.Do(req); err != nil {
		return err
	}

	// If a bad status code is given, then return an error
	if !(status == http.StatusOK || status == http.StatusNoContent || status == http.StatusUnauthorized) {
		return fmt.Errorf("could not logout received status %q", http.StatusText(status))
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
		tokens map[string]interface{}
	)

	// Build request
	data := make(map[string]interface{})
	data["refresh_token"] = c.creds.Tokens.Refresh
	data["no_cookie"] = true
	if req, err = c.NewRequest(http.MethodPost, "/refresh", false, data); err != nil {
		return err
	}

	// Execute the request
	if tokens, status, err = c.Do(req); err != nil {
		return err
	}

	if status != http.StatusOK {
		// TODO: better error handling here
		return fmt.Errorf("got a %s status", http.StatusText(status))
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