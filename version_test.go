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
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	s.router.ServeHTTP(w, req)

	require.Equal(s.T(), http.StatusPermanentRedirect, w.Code)
	require.Equal(s.T(), "/v1", w.Result().Header.Get("Location"))
}
