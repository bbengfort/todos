package todos_test

import (
	"os"
	"testing"

	. "github.com/bbengfort/todos"
	"github.com/stretchr/testify/require"
)

var requiredEnv = map[string]string{
	"TODOS_MODE":   "test",
	"DATABASE_URL": "postgres://benjamin@localhost:5432/todos",
	"SECRET_KEY":   "supersecretkey",
}

func unsetEnv(t *testing.T) func() {
	return func() {
		for key := range requiredEnv {
			require.NoError(t, os.Unsetenv(key))
		}
	}
}

func TestConfig(t *testing.T) {
	for key, val := range requiredEnv {
		require.NoError(t, os.Setenv(key, val))
	}

	defer unsetEnv(t)()

	conf, err := Config()
	require.NoError(t, err)

	// Check defaults
	require.Equal(t, "127.0.0.1:8080", conf.Addr())
	require.Equal(t, "http://localhost:8080/", conf.Endpoint())
}

func TestBadConfigs(t *testing.T) {
	defer unsetEnv(t)()

	_, err := Config()
	require.EqualError(t, err, "required key SECRET_KEY missing value")
	require.NoError(t, os.Setenv("SECRET_KEY", "theeaglefliesatmidnight"))

	_, err = Config()
	require.EqualError(t, err, "required key DATABASE_URL missing value")
	require.NoError(t, os.Setenv("DATABASE_URL", "postgres://benjamin@localhost:5432/todos"))

	// All defaults set correctly at this point
	_, err = Config()
	require.NoError(t, err)

	require.NoError(t, os.Setenv("TODOS_MODE", "fakemode"))
	_, err = Config()
	require.EqualError(t, err, "\"fakemode\" is an unknown mode, use \"debug\", \"release\", or \"test\"")
}

func TestEndpoint(t *testing.T) {
	s := Settings{
		UseTLS: true,
		Domain: "api.todos.bengfort.com",
		Port:   443,
	}

	require.Equal(t, "https://api.todos.bengfort.com/", s.Endpoint())

	s.Port = 5356
	require.Equal(t, "https://api.todos.bengfort.com:5356/", s.Endpoint())

	s.UseTLS = false
	require.Equal(t, "http://api.todos.bengfort.com:5356/", s.Endpoint())

	s.Port = 80
	require.Equal(t, "http://api.todos.bengfort.com/", s.Endpoint())
}
