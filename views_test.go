package todos_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/require"
)

func (s *TodosTestSuite) TestOverview() {
	// Test overview requires authorized login
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/", nil)
	s.router.ServeHTTP(w, req)

	require.Equal(s.T(), http.StatusUnauthorized, w.Code)

	// Test overview with authenticated user
	access := s.Login(false)
	require.NotZero(s.T(), access)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/v1/", nil)
	req.Header.Set("Authorization", "Bearer "+access)
	s.router.ServeHTTP(w, req)

	result := w.Result()
	require.Equal(s.T(), http.StatusOK, w.Code)
	require.Equal(s.T(), "application/json; charset=utf-8", result.Header.Get("Content-Type"))

	var data map[string]interface{}
	err := json.NewDecoder(result.Body).Decode(&data)
	require.NoError(s.T(), err)
}
