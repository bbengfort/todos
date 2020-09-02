package todos_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bbengfort/todos"
	"github.com/stretchr/testify/require"
)

const ExpectedVersion = "1.2"

func TestVersion(t *testing.T) {
	require.Equal(t, ExpectedVersion, todos.Version())
	require.Equal(t, "/v1", todos.VersionURL())
}

func (s *TodosTestSuite) TestRedirectVersion() {
	tt := []struct {
		header   string
		location string
		code     int
	}{
		{"", "", http.StatusNotAcceptable},
		{"*/*", "/v1", http.StatusPermanentRedirect},
		{"application/json", "/v1", http.StatusPermanentRedirect},
		{"application/*", "/v1", http.StatusPermanentRedirect},
		{"text/html", "/app", http.StatusPermanentRedirect},
		{"text/*", "/app", http.StatusPermanentRedirect},
	}

	for _, tc := range tt {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Add("Accept", tc.header)
		s.router.ServeHTTP(w, req)

		require.Equal(s.T(), tc.code, w.Code)
		require.Equal(s.T(), tc.location, w.Result().Header.Get("Location"))
	}
}
