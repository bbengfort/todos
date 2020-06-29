package client

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/shibukawa/configdir"
	"gopkg.in/yaml.v3"
)

const (
	vendorName      = "bbengfort"
	applicationName = "todos"
	credentialsFile = "credentials.yaml"
)

// Configuration returns the directory of the configuration file(s).
func Configuration() (string, error) {
	configDirs := configdir.New(vendorName, applicationName)
	folder := configDirs.QueryFolderContainsFile(credentialsFile)
	if folder == nil {
		return "", ErrNoConfiguration
	}
	return folder.Path, nil
}

// Credentials stores login and configuration information to connect to the todos API.
// The only required item is the endpoint to make requests to. If the username or
// password are not provided in the credentials, then they will be prompted for during
// login. The access token is used in the Bearer header to make request. After the
// NotBefore timestamp, the access token is automatically refreshed until the refresh
// token expires. If the password is stored, then automatic login occurs in this case.
// Note that the local client can only maintain one set of credentials at a time.
type Credentials struct {
	Endpoint string `yaml:"endpoint"`           // the endpoint to connect to
	Username string `yaml:"username,omitempty"` // username to login with (optional)
	Password string `yaml:"password,omitempty"` // password to login with (optional)
	Tokens   struct {
		Access    string    `yaml:"access"`     // access token to send with Bearer requests
		Refresh   string    `yaml:"refresh"`    // refresh token to obtain a new access token without login
		IssuedAt  time.Time `yaml:"issued_at"`  // timestamp of the login
		ExpiresAt time.Time `yaml:"expires_at"` // when the access token expires
		NotBefore time.Time `yaml:"not_before"` // earliest timestamp the access token can be refreshed
		RefreshBy time.Time `yaml:"refresh_by"` // when the refresh token expires
	} `yaml:"tokens,omitempty"` // access and refresh tokens for requests
}

// Dump the credentials to an OS specific configuration folder.
func (c *Credentials) Dump() (err error) {
	// Dump data to yaml
	var data []byte
	if data, err = yaml.Marshal(&c); err != nil {
		return fmt.Errorf("could not marshal yaml: %s", err)
	}

	// Find user configuration directory to write data to
	configDirs := configdir.New(vendorName, applicationName)
	folders := configDirs.QueryFolders(configdir.Global)

	if len(folders) == 0 {
		return ErrNoConfDir
	}

	// Write the data to the credentials file
	if err = folders[0].WriteFile(credentialsFile, data); err != nil {
		return fmt.Errorf("could not write %s to %s: %s", credentialsFile, folders[0].Path, err)
	}
	return nil
}

// Load the credentials from an OS specific configuration folder.
func (c *Credentials) Load() (err error) {
	// Find user configuration directory to read data from
	configDirs := configdir.New(vendorName, applicationName)
	folder := configDirs.QueryFolderContainsFile(credentialsFile)

	if folder == nil {
		return ErrNoCredentials
	}

	// Read the credentials from disk
	var data []byte
	if data, err = folder.ReadFile(credentialsFile); err != nil {
		return fmt.Errorf("could not read %s in %s: %s", credentialsFile, folder.Path, err)
	}

	// Unmarshall the credentials into the struct
	if err = yaml.Unmarshal(data, &c); err != nil {
		return fmt.Errorf("could not unmarshal yaml: %s", err)
	}

	// Validate that the required fields are available
	if c.Endpoint == "" {
		return ErrNoEndpoint
	}
	if _, err = url.Parse(c.Endpoint); err != nil {
		return err
	}

	return nil
}

// IsLoggedIn returns true if the credentials hold an access token that is still valid,
// e.g. it has not expired yet. This function does not modify the credentials file.
func (c *Credentials) IsLoggedIn() bool {
	if c.Tokens.Access != "" {
		return time.Now().Before(c.Tokens.ExpiresAt)
	}
	return false
}

// IsRefreshable returns true if the credentials hold a refresh token that is still
// valid and a refresh request can be issued at the current time.
func (c *Credentials) IsRefreshable() bool {
	if c.Tokens.Refresh != "" {
		now := time.Now()
		return now.After(c.Tokens.NotBefore) && now.Before(c.Tokens.RefreshBy)
	}
	return false
}

// GetURL constructs a complete URL to the specified location from the base endpoint
func (c *Credentials) GetURL(path string) (_ string, err error) {
	var (
		ep  *url.URL
		ref *url.URL
	)

	if ref, err = url.Parse(path); err != nil {
		return "", fmt.Errorf("could not parse %q as reference to endpoint: %s", path, err)
	}

	// TODO: cached parsed endpoint on credentials struct
	if ep, err = url.Parse(c.Endpoint); err != nil {
		return "", fmt.Errorf("could not parse %q endpoint: %s", c.Endpoint, err)
	}
	return ep.ResolveReference(ref).String(), nil
}

// MustGetURL panics if the url or path cannot be parsed
func (c *Credentials) MustGetURL(path string) string {
	url, err := c.GetURL(path)
	if err != nil {
		panic(err)
	}
	return url
}

// SetTokens on the credentials and save the credentials to disk for future use.
func (c *Credentials) SetTokens(tokens map[string]interface{}) (err error) {
	var ok bool
	var access, refresh interface{}

	if access, ok = tokens["access_token"]; !ok {
		return errors.New("response does not contain an access_token")
	}

	if refresh, ok = tokens["refresh_token"]; !ok {
		return errors.New("response does not contain an refresh_token")
	}

	// Set the tokens on the credentials
	c.Tokens.Access = access.(string)
	c.Tokens.Refresh = refresh.(string)

	// Parse the access token for expiration times
	ac, err := parseToken(c.Tokens.Access)
	if err != nil {
		return fmt.Errorf("could not parse access token: %s", err)
	}
	c.Tokens.IssuedAt = time.Unix(ac.IssuedAt, 0)
	c.Tokens.ExpiresAt = time.Unix(ac.ExpiresAt, 0)

	// Parse the refresh token for expiration times
	rc, err := parseToken(c.Tokens.Refresh)
	if err != nil {
		return fmt.Errorf("could not parse refresh token: %s", err)
	}
	c.Tokens.NotBefore = time.Unix(rc.NotBefore, 0)
	c.Tokens.RefreshBy = time.Unix(rc.ExpiresAt, 0)

	// Save the new tokens back to disk.
	return c.Dump()
}

func parseToken(tks string) (_ *jwt.StandardClaims, err error) {
	claims := &jwt.StandardClaims{}
	if _, _, err = new(jwt.Parser).ParseUnverified(tks, claims); err != nil {
		return nil, err
	}
	return claims, nil
}

// Revoke the tokens in the credentials file and overwrite the previous file.
func (c *Credentials) Revoke() (err error) {
	c.Tokens.Access = ""
	c.Tokens.Refresh = ""
	c.Tokens.IssuedAt = time.Time{}
	c.Tokens.ExpiresAt = time.Time{}
	c.Tokens.NotBefore = time.Time{}
	c.Tokens.RefreshBy = time.Time{}

	return c.Dump()
}
