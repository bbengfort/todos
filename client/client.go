package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// New creates a new todos API client and prepares the credentials and configuration.
// TODO: handle TLS transport.
func New() (c *Client, err error) {
	c = &Client{
		Client: http.Client{
			Transport: &http.Transport{
				MaxIdleConns:       4,
				IdleConnTimeout:    1 * time.Minute,
				DisableCompression: false,
			},
			Jar: nil,
		},
		creds: &Credentials{},
	}

	if err = c.creds.Load(); err != nil {
		return c, err
	}
	return c, nil
}

// Client interacts with the todos API server.
type Client struct {
	http.Client
	creds *Credentials
}

// NewRequest creates an http request to the endpoint specified in the credentials and
// sets the appropriate headers for the request, including authentication if required.
func (c *Client) NewRequest(method, url string, auth bool, data interface{}) (req *http.Request, err error) {
	var body io.Reader
	if data != nil {
		var payload []byte
		if payload, err = json.Marshal(data); err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(payload)
	} else {
		body = nil
	}

	if req, err = http.NewRequest(method, c.creds.MustGetURL(url), body); err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	if auth {
		if !c.creds.IsLoggedIn() {
			return nil, ErrNotLoggedIn
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.creds.Tokens.Access))
	}

	return req, nil
}

// Status returns the current status of the todo API server.
func (c *Client) Status() (data map[string]interface{}, err error) {
	var (
		req *http.Request
		rep *http.Response
	)

	if req, err = c.NewRequest(http.MethodGet, "status", false, nil); err != nil {
		return nil, err
	}

	if rep, err = c.Do(req); err != nil {
		return nil, err
	}
	defer rep.Body.Close()

	// Read the body into json
	if err = json.NewDecoder(rep.Body).Decode(&data); err != nil {
		return nil, err
	}

	// Add status code if a non 200 status is returned
	if rep.StatusCode != http.StatusOK {
		data["status"] = rep.Status
	}

	return data, nil
}

//===========================================================================
// Authentication
//===========================================================================

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

	var rep *http.Response
	if rep, err = c.Do(req); err != nil {
		return err
	}
	defer rep.Body.Close()

	if rep.StatusCode != http.StatusOK {
		// TODO: better error handling here
		return fmt.Errorf("got a %s status", rep.Status)
	}

	// Parse the data response
	var tokens map[string]interface{}
	if err = json.NewDecoder(rep.Body).Decode(&tokens); err != nil {
		return err
	}

	// Set the tokens on the credentials and save them to disk
	if err = c.creds.SetTokens(tokens); err != nil {
		return err
	}
	return nil
}
