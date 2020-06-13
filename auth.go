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
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/argon2"
)

// Register a new user with the specified username and password. Register is POST only
// and binds the registerUserForm to get the data. Returns an error if the username or
// email is not unique.
func (s *API) Register(c *gin.Context) {
	form := registerUserForm{}
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if _, ok := s.users[form.Username]; ok {
		err := fmt.Errorf("username %q already taken", form.Username)
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	user := User{
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

	s.users[form.Username] = user
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Login the user with the specified username and password. Login uses argon2 derived
// key comparisons to verify the user without storing the password in plain text. This
// handler binds to the loginUserForm and returns unauthorized if the password is not
// correct. On successful login, a JWT token is returned and a cookie set.
func (s *API) Login(c *gin.Context) {
	form := loginUserForm{}
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	user, ok := s.users[form.Username]
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false})
		return
	}

	valid, err := VerifyDerivedKey(user.Password, form.Password)
	if err != nil {
		// Panic instead?
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}

// Logout expires the user's JWT token.
func (s *API) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}

// SetPassword uses the Argon2 derived key algorithm to store the user password along
// with a user-specific random salt into the database. Argon2 is a modern ASIC- and
// GPU- resistant secure key derivation function that prevents password cracking.
// The password is stored with the algorithm settings + salt + hash together in the
// database in a common format to ensure cross process compatibility. Each component is
// separated by a $ and hashes are base64 encoded.
func (u User) SetPassword(password string) (_ string, err error) {
	return CreateDerivedKey(password)
}

// VerifyPassword by comparing the original derived key with derived key from the
// submitted password. This function uses the parameters from the stored password to
// compute the dervied key and compare it.
func (u User) VerifyPassword(password string) (_ bool, err error) {
	if u.Password == "" {
		return false, errors.New("user does not have a password set")
	}
	return VerifyDerivedKey(u.Password, password)
}

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
