package todos_test

import (
	"os"
	"testing"

	. "github.com/bbengfort/todos"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	reqenv := map[string]string{
		"TODOS_MODE":   "test",
		"DATABASE_URL": "postgres://benjamin@localhost:5432/todos",
		"SECRET_KEY":   "supersecretkey",
	}

	for key, val := range reqenv {
		require.NoError(t, os.Setenv(key, val))
	}

	conf, err := Config()
	require.NoError(t, err)

	// Check defaults
	require.Equal(t, "127.0.0.1:8080", conf.Addr())
	require.Equal(t, "http://localhost:8080/", conf.Endpoint())
}
