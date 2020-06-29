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

// Do the http request and parse the JSON response returning the data and code.
func (c *Client) Do(req *http.Request) (map[string]interface{}, int, error) {
	rep, err := c.Client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer rep.Body.Close()

	if ct := rep.Header.Get("Content-Type"); ct != "application/json; charset=utf-8" {
		return nil, rep.StatusCode, fmt.Errorf("unexpected content type: %s", ct)
	}

	data := make(map[string]interface{})
	if err = json.NewDecoder(rep.Body).Decode(&data); err != nil {
		return nil, rep.StatusCode, err
	}

	return data, rep.StatusCode, err
}

//===========================================================================
// Status Methods
//===========================================================================

// Status returns the current status of the todo API server.
func (c *Client) Status() (data map[string]interface{}, err error) {
	var req *http.Request
	if req, err = c.NewRequest(http.MethodGet, "/status", false, nil); err != nil {
		return nil, err
	}

	var status int
	if data, status, err = c.Do(req); err != nil {
		return nil, err
	}

	// Add status code if a non 200 status is returned
	if status != http.StatusOK {
		data["status_code"] = http.StatusText(status)
	}

	return data, nil
}

// Overview returns the user's current todo listing and m ust be authenticated.
func (c *Client) Overview() (data map[string]interface{}, err error) {
	var req *http.Request
	if req, err = c.NewRequest(http.MethodGet, "/", true, nil); err != nil {
		return nil, err
	}

	var status int
	if data, status, err = c.Do(req); err != nil {
		return nil, err
	}

	// Add status code if a non 200 status is returned
	if status != http.StatusOK {
		data["status_code"] = http.StatusText(status)
	}

	return data, nil
}
