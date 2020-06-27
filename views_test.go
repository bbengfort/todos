package todos_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bbengfort/todos"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestOverview(t *testing.T) {
	// Set gin mode
	gin.SetMode(gin.TestMode)

	// TODO: mock database
	conf, err := todos.Config()
	require.NoError(t, err)
	api, err := todos.New(conf)
	require.NoError(t, err)

	router := api.Routes(true)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	// TODO: how to handle auth-required routes?
	require.Equal(t, http.StatusUnauthorized, w.Code)
}
