package todos_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/bbengfort/todos"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/suite"
)

const (
	userUsername  = "jane"
	userEmail     = "jane@example.com"
	userPassword  = "specialsnowflake"
	adminUsername = "admin"
	adminEmail    = "server@example.com"
	adminPassword = "hackme42"
)

// TodosTestSuite mocks the database and gin/http requests for testing endpoints.
type TodosTestSuite struct {
	suite.Suite
	api              *API
	conf             Settings
	router           http.Handler
	adminAccessToken string
	userAccessToken  string
}

func (s *TodosTestSuite) SetupSuite() {
	var err error
	gin.SetMode(gin.TestMode)

	// Create test configuration for mocked database and server
	s.conf = Settings{
		Mode:        gin.TestMode,
		UseTLS:      false,
		Bind:        "127.0.0.1",
		Port:        8080,
		Domain:      "localhost",
		DatabaseURL: "file::memory:?cache=shared",
		SecretKey:   "supersecretkey",
	}

	// Create the api, which will setup both the routes and the database
	s.api, err = New(s.conf)
	s.NoError(err)

	// Get the routes from the server
	s.router = s.api.Routes()

	// Set the server as healthy
	s.api.SetHealth(true)
}

func TestTodos(t *testing.T) {
	suite.Run(t, new(TodosTestSuite))
}

func (s *TodosTestSuite) TestGinMode() {
	s.Equal(gin.TestMode, gin.Mode())
}

func (s *TodosTestSuite) RequireUser() {
	var (
		user User
		err  error
	)

	db := s.api.DB()
	err = db.Where(User{Username: userUsername}).First(&user).Error
	if gorm.IsRecordNotFoundError(err) {
		user = User{
			Username: userUsername,
			Email:    userEmail,
		}
		user.Password, err = user.SetPassword(userPassword)
		s.NoError(err)

		err = db.Create(&user).Error
		s.NoError(err)
	}
}

func (s *TodosTestSuite) RequireAdmin() {
	var (
		user User
		err  error
	)

	db := s.api.DB()
	err = db.Where(User{Username: adminUsername}).First(&user).Error
	if gorm.IsRecordNotFoundError(err) {
		user = User{
			Username: adminUsername,
			Email:    adminEmail,
		}
		user.Password, err = user.SetPassword(adminPassword)
		s.NoError(err)

		err = db.Create(&user).Error
		s.NoError(err)
	}
}

func (s *TodosTestSuite) Login(admin bool) (token string) {
	if admin {
		if s.adminAccessToken != "" {
			return s.adminAccessToken
		}
	} else {
		if s.userAccessToken != "" {
			return s.userAccessToken
		}
	}

	form := make(map[string]interface{})
	form["no_cookie"] = true

	if admin {
		s.RequireAdmin()
		form["username"] = adminUsername
		form["password"] = adminPassword
	} else {
		s.RequireUser()
		form["username"] = userUsername
		form["password"] = userPassword
	}

	data, err := json.Marshal(form)
	s.NoError(err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/login", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var tokens map[string]interface{}
	err = json.NewDecoder(w.Result().Body).Decode(&tokens)
	s.NoError(err)

	access := tokens["access_token"].(string)
	s.NotEmpty(access)

	if admin {
		s.adminAccessToken = access
	} else {
		s.userAccessToken = access
	}

	return access
}
