package todos

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

//===========================================================================
// Handlers
//===========================================================================

// TODO: remove
var useridsequence uint

// Register a new user with the specified username and password. Register is POST only
// and binds the registerUserForm to get the data. Returns an error if the username or
// email is not unique.
func (s *API) Register(c *gin.Context) {
	// Bind and parse the POST data
	form := registerUserForm{}
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Check uniqueness constraints (maybe not necessary with database)
	// if _, ok := s.users[form.Username]; ok {
	// 	err := fmt.Errorf("username %q already taken", form.Username)
	// 	c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
	// 	return
	// }

	// Create the user with a derived key password
	useridsequence++
	user := User{
		ID:        useridsequence,
		Username:  form.Username,
		Email:     form.Email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	var err error
	if user.Password, err = CreateDerivedKey(form.Password); err != nil {
		// TODO: should panic instead?
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	// Insert the user into the database
	s.users[user.ID] = user
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Login the user with the specified username and password. Login uses argon2 derived
// key comparisons to verify the user without storing the password in plain text. This
// handler binds to the loginUserForm and returns unauthorized if the password is not
// correct. On successful login, a JWT token is returned and a cookie set.
func (s *API) Login(c *gin.Context) {
	// TODO: check if already logged in and return bad request if so (must logout or use refresh)

	// Bind and parse the POST data
	form := loginUserForm{}
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Lookup the user in the database
	var ok bool
	var user User
	for _, u := range s.users {
		if u.Username == form.Username {
			user = u
			ok = true
			break
		}
	}
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false})
		return
	}

	// Verify the password
	valid, err := VerifyDerivedKey(user.Password, form.Password)
	if err != nil {
		// Panic instead?
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	// If password does not match, deny access
	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false})
		return
	}

	// Issue new JWT tokens for the user
	token, err := CreateAuthToken(user.ID)
	if err != nil {
		// Panic instead?
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
	}

	// Save the token to the database
	s.tokens[token.ID] = token

	if !form.NoCookie {
		// Store the tokens as cookies
		// TODO: get the domains from a configuration
		c.SetCookie(jwtAccessCookieName, token.accessToken, int(jwtAccessTokenDuration.Seconds()), "/", "todos.bengfort.com", false, true)
		c.SetCookie(jwtRefreshCookieName, token.refreshToken, int(jwtRefreshTokenDuration.Seconds()), "/", "todos.bengfort.com", false, true)
	}

	// Return the tokens for use by the api as Bearer headers
	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"access_token":  token.accessToken,
		"refresh_token": token.refreshToken,
	})
}

// Logout expires the user's JWT token. Note that Logout does not have the authorization
// middleware so lookups up the access token in the same manner as that middleware. If
// no access token is provided, then a bad request is returned. Revoke all will delete
// all tokens for the user with the provided access token.
func (s *API) Logout(c *gin.Context) {
	tokenString, err := FindToken(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	tokenID, err := VerifyAuthToken(tokenString, true, false)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false})
		return
	}

	// TODO: use the database to look up the token
	token, ok := s.tokens[tokenID]
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false})
		return
	}

	// Bind and parse the POST data
	form := logoutUserForm{}
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if form.RevokeAll {
		// TODO: issue the database query to do this
		for id, other := range s.tokens {
			if other.UserID == token.UserID {
				delete(s.tokens, id)
			}
		}
	} else {
		// Delete just the single token
		delete(s.tokens, tokenID)
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Refresh the access token with the refresh token if it's available and valid. The
// refresh token is essentially a time-limited one time key that will allow the user to
// reauthenticate without a username and password. If the user logs out, the refresh
// token will be revoked and no longer usable. Note that the server does not do any
// verification of the refresh token so it should be kept secret by the client in the
// same way a username and password should be kept secret. However, because the refresh
// token can be revoked and automatically expires, it is a slightly safer mechanism of
// reauthentication than resending a username and password combination.
func (s *API) Refresh(c *gin.Context) {
	// Bind and parse the POST data
	form := refreshTokenForm{}
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	tokenID, err := VerifyAuthToken(form.RefreshToken, false, true)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false})
		return
	}

	// TODO: use the database to look up the old token
	refresh, ok := s.tokens[tokenID]
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false})
		return
	}

	// Issue new JWT tokens for the user
	token, err := CreateAuthToken(refresh.UserID)
	if err != nil {
		// Panic instead?
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
	}

	// Save the token to the database
	s.tokens[token.ID] = token

	if !form.NoCookie {
		// Store the tokens as cookies
		// TODO: get the domains from a configuration
		c.SetCookie(jwtAccessCookieName, token.accessToken, int(jwtAccessTokenDuration.Seconds()), "/", "todos.bengfort.com", false, true)
		c.SetCookie(jwtRefreshCookieName, token.refreshToken, int(jwtRefreshTokenDuration.Seconds()), "/", "todos.bengfort.com", false, true)
	}

	// Revoke the old tokens
	delete(s.tokens, tokenID)

	// Return the tokens for use by the api as Bearer headers
	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"access_token":  token.accessToken,
		"refresh_token": token.refreshToken,
	})
}

//===========================================================================
// Middleware
//===========================================================================

// Authorize is middleware that checks for an access token in the request and only
// allows processing to proceed if the user is valid and authorized. The middleware also
// loads the user information into the context so that it is available to downstream
// handlers.
func (s *API) Authorize() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := FindToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false})
			c.Abort()
			return
		}

		tokenID, err := VerifyAuthToken(tokenString, true, false)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false})
			c.Abort()
			return
		}

		// TODO: use the database to look up the token (and also fetch user related struct, for only single query)
		token, ok := s.tokens[tokenID]
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false})
			c.Abort()
			return
		}

		// TODO: use the database query from before to add the user
		c.Set(ctxUserKey, s.users[token.UserID])

		// Everything checks out, user is good to go
		c.Next()
	}
}

//===========================================================================
// User Methods
//===========================================================================

// SetPassword uses the Argon2 derived key algorithm to store the user password along
// with a user-specific random salt into the database. Argon2 is a modern ASIC- and
// GPU- resistant secure key derivation function that prevents password cracking.
// The password is stored with the algorithm settings + salt + hash together in the
// database in a common format to ensure cross process compatibility. Each component is
// separated by a $ and hashes are base64 encoded.
func (u User) SetPassword(password string) (_ string, err error) {
	// TODO: store in the database directly without modifying user struct
	return CreateDerivedKey(password)
}

// VerifyPassword by comparing the original derived key with derived key from the
// submitted password. This function uses the parameters from the stored password to
// compute the dervied key and compare it.
func (u User) VerifyPassword(password string) (_ bool, err error) {
	// TODO: fetch password from the database (does not use user struct)
	if u.Password == "" {
		return false, errors.New("user does not have a password set")
	}
	return VerifyDerivedKey(u.Password, password)
}

//===========================================================================
// Derived Key Algorithm
//===========================================================================

// Argon2 constants for the derived key (dk) algorithm
// See: https://cryptobook.nakov.com/mac-and-key-derivation/argon2
const (
	dkAlg  = "argon2id"        // the derived key algorithm
	dkTime = uint32(1)         // draft RFC recommends time = 1
	dkMem  = uint32(64 * 1024) // draft RFC recommends memory as ~64MB (or as much as possible)
	dkProc = uint8(2)          // can be set to the number of available CPUs
	dkSLen = 16                // the length of the salt to generate per user
	dkKLen = uint32(32)        // the length of the derived key (32 bytes is the required key size for AES-256)
)

// Argon2 variables for the derived key (dk) algorithm
var (
	dkParse = regexp.MustCompile(`^\$(?P<alg>[\w\d]+)\$v=(?P<ver>\d+)\$m=(?P<mem>\d+),t=(?P<time>\d+),p=(?P<procs>\d+)\$(?P<salt>[\+\/\=a-zA-Z0-9]+)\$(?P<key>[\+\/\=a-zA-Z0-9]+)$`)
)

// CreateDerivedKey creates an encoded derived key with a random hash for the password.
func CreateDerivedKey(password string) (_ string, err error) {
	salt := make([]byte, dkSLen)
	if _, err = rand.Read(salt); err != nil {
		return "", fmt.Errorf("could not generate %d length salt: %s", dkSLen, err)
	}

	dk := argon2.IDKey([]byte(password), salt, dkTime, dkMem, dkProc, dkKLen)
	b64salt := base64.StdEncoding.EncodeToString(salt)
	b64dk := base64.StdEncoding.EncodeToString(dk)
	return fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$%s$%s", dkAlg, argon2.Version, dkMem, dkTime, dkProc, b64salt, b64dk), nil
}

// VerifyDerivedKey checks that the submitted password matches the derived key.
func VerifyDerivedKey(dk, password string) (_ bool, err error) {
	if dk == "" || password == "" {
		return false, errors.New("cannot verify empty derived key or password")
	}

	dkb, salt, t, m, p, err := ParseDerivedKey(dk)
	if err != nil {
		return false, err
	}

	vdk := argon2.IDKey([]byte(password), salt, t, m, p, uint32(len(dkb)))
	return bytes.Equal(dkb, vdk), nil
}

// ParseDerivedKey returns the parts of the encoded derived key string.
func ParseDerivedKey(encoded string) (dk, salt []byte, time, memory uint32, threads uint8, err error) {
	if !dkParse.MatchString(encoded) {
		return nil, nil, 0, 0, 0, errors.New("cannot parse encoded derived key, does not match regular expression")
	}
	parts := dkParse.FindStringSubmatch(encoded)

	if len(parts) != 8 {
		return nil, nil, 0, 0, 0, errors.New("cannot parse encoded derived key, matched expression does not contain enough subgroups")
	}

	// check the algorithm
	if parts[1] != dkAlg {
		return nil, nil, 0, 0, 0, fmt.Errorf("current code only works with the the dk protcol %q not %q", dkAlg, parts[2])
	}

	// check the version
	if version, err := strconv.Atoi(parts[2]); err != nil || version != argon2.Version {
		return nil, nil, 0, 0, 0, fmt.Errorf("expected %s version %d got %q", dkAlg, argon2.Version, parts[2])
	}

	var (
		time64    uint64
		memory64  uint64
		threads64 uint64
	)

	if memory64, err = strconv.ParseUint(parts[3], 10, 32); err != nil {
		return nil, nil, 0, 0, 0, fmt.Errorf("could not parse memory %q: %s", parts[3], err)
	}
	memory = uint32(memory64)

	if time64, err = strconv.ParseUint(parts[4], 10, 32); err != nil {
		return nil, nil, 0, 0, 0, fmt.Errorf("could not parse time %q: %s", parts[4], err)
	}
	time = uint32(time64)

	if threads64, err = strconv.ParseUint(parts[5], 10, 8); err != nil {
		return nil, nil, 0, 0, 0, fmt.Errorf("could not parse threads %q: %s", parts[5], err)
	}
	threads = uint8(threads64)

	if salt, err = base64.StdEncoding.DecodeString(parts[6]); err != nil {
		return nil, nil, 0, 0, 0, fmt.Errorf("could not parse salt: %s", err)
	}

	if dk, err = base64.StdEncoding.DecodeString(parts[7]); err != nil {
		return nil, nil, 0, 0, 0, fmt.Errorf("could not parse derived key: %s", err)
	}

	return dk, salt, time, memory, threads, nil
}

//===========================================================================
// JWT Tokens
//===========================================================================

// JWT constants for access and refresh tokens
// See: https://www.cloudjourney.io/articles/security/jwt_in_golang-su/
const (
	// jwtAccessTokenDuration  = 4 * time.Hour
	// jwtRefreshTokenDuration = 12 * time.Hour
	// jwtAccessRefreshOverlap = -1 * time.Minute
	jwtAccessTokenDuration  = 2 * time.Minute
	jwtRefreshTokenDuration = 12 * time.Minute
	jwtAccessRefreshOverlap = -1 * time.Minute
	jwtAccessTokenAudience  = "access"
	jwtRefreshTokenAudience = "refresh"
	jwtAccessCookieName     = "access_token"
	jwtRefreshCookieName    = "refresh_token"
)

// JWT variables for access and refresh tokens
var (
	jwtKey           = []byte("supersecretkey") // TODO: fetch from the environment
	jwtSigningMethod = jwt.SigningMethodHS256   // Declared here for consistency
	jwtKeyFunc       = func(token *jwt.Token) (interface{}, error) {
		// TODO: should we have separate keys for access and refresh tokens?
		return jwtKey, nil
	}
)

// AccessClaims returns the jwt.StandardClaims for the access token.
func (t Token) AccessClaims() jwt.Claims {
	return &jwt.StandardClaims{
		Id:        t.ID.String(),
		Audience:  jwtAccessTokenAudience,
		IssuedAt:  t.IssuedAt.Unix(),
		ExpiresAt: t.ExpiresAt.Unix(),
	}
}

// AccessToken returns the cached access token or generates it from the claims.
func (t Token) AccessToken() (token string, err error) {
	// Return the cached access token if available
	if t.accessToken != "" {
		return t.accessToken, nil
	}

	// Generate the access token (but does not cache)
	at := jwt.NewWithClaims(jwtSigningMethod, t.AccessClaims())
	if token, err = at.SignedString(jwtKey); err != nil {
		return "", fmt.Errorf("could not generate access token: %s", err)
	}
	return token, nil
}

// RefreshClaims returns the jwt.StandardClaims for the refresh token. Note that a
// refresh token cannot be used until one minute within the access token expiration.
func (t Token) RefreshClaims() jwt.Claims {
	return &jwt.StandardClaims{
		Id:        t.ID.String(),
		Audience:  jwtRefreshTokenAudience,
		IssuedAt:  t.IssuedAt.Unix(),
		ExpiresAt: t.RefreshBy.Unix(),
		NotBefore: t.ExpiresAt.Add(jwtAccessRefreshOverlap).Unix(),
	}
}

// RefreshToken returns the cached refresh token or generates it from the claims.
func (t Token) RefreshToken() (token string, err error) {
	// Return the cached refresh token if available
	if t.refreshToken != "" {
		return t.refreshToken, nil
	}

	// Generate the access token (but does not cache)
	rt := jwt.NewWithClaims(jwtSigningMethod, t.RefreshClaims())
	if token, err = rt.SignedString(jwtKey); err != nil {
		return "", fmt.Errorf("could not generate refresh token: %s", err)
	}
	return token, nil
}

// CreateAuthToken generates acccess and refresh tokens for API authorization using a
// cookie or Bearer header and stores them in the database. A single user can create
// multiple auth tokens and each of them are assigned a unique uuid for lookup.
func CreateAuthToken(user uint) (token Token, err error) {
	// Create the token record in the database
	now := time.Now()
	token = Token{
		ID:        uuid.New(),
		UserID:    user,
		IssuedAt:  now,
		ExpiresAt: now.Add(jwtAccessTokenDuration),
		RefreshBy: now.Add(jwtRefreshTokenDuration),
	}

	// Sign and generate the accessToken (caching it and ensuring no errors)
	if token.accessToken, err = token.AccessToken(); err != nil {
		return Token{}, err
	}

	// Sign and generate the refreshToken
	if token.refreshToken, err = token.RefreshToken(); err != nil {
		return Token{}, err
	}

	return token, nil
}

// VerifyAuthToken validates an access or refresh token string with its signature and claims
// fields and verifies the token is an access or refresh token if required by the input.
// If the token is valid, the database token id is returned without error, otherwise an
// error is returned to indicate that the token is no longer valid.
func VerifyAuthToken(tokenString string, access, refresh bool) (id uuid.UUID, err error) {
	var token *jwt.Token
	claims := &jwt.StandardClaims{}
	if token, err = jwt.ParseWithClaims(tokenString, claims, jwtKeyFunc); err != nil {
		return uuid.Nil, err
	}

	if !token.Valid {
		// It is likely that we will never reach this line of code
		return uuid.Nil, errors.New("token is invalid or has expired")
	}

	if access && !claims.VerifyAudience(jwtAccessTokenAudience, true) {
		return uuid.Nil, errors.New("token is not an access token")
	}

	if refresh && !claims.VerifyAudience(jwtRefreshTokenAudience, true) {
		return uuid.Nil, errors.New("token is not a refresh token")
	}

	if id, err = uuid.Parse(claims.Id); err != nil {
		return uuid.Nil, fmt.Errorf("could not parse token id: %s", err)
	}

	return id, nil
}

// FindToken uses the gin context to look up the access token in the Bearer header, in
// the cookies of the request, or as a url request parameter. It returns an error if it
// cannot find the token string.
func FindToken(c *gin.Context) (token string, err error) {
	// Check the Bearer header for the token (usual place)
	bearer := c.GetHeader("Authorization")
	if bearer != "" {
		parts := strings.Split(bearer, "Bearer ")
		if len(parts) == 2 {
			return strings.TrimSpace(parts[1]), nil
		}
		return "", errors.New("could not parse Bearer authorization header")
	}

	// Check the cookie for the token
	if cookie, err := c.Cookie(jwtAccessCookieName); err == nil {
		return cookie, nil
	}

	// Check the request parameters for the token
	if param := c.Request.URL.Query().Get("token"); param != "" {
		return param, nil
	}

	return "", errors.New("no access token found in header, cookie, or request")
}
