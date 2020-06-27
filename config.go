package todos

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/kelseyhightower/envconfig"
)

// Config creates a new Settings object, loading it from the environment, processing
// default values and validating the configuration. If the Settings cannot be loaded,
// or validated then an error is returned.
func Config() (conf Settings, err error) {
	if err = envconfig.Process("todos", &conf); err != nil {
		return Settings{}, err
	}

	// Ensure mode is a gin.Mode
	if conf.Mode != gin.DebugMode && conf.Mode != gin.ReleaseMode && conf.Mode != gin.TestMode {
		return Settings{}, fmt.Errorf("%q is an unknown mode, use %q, %q, or %q", conf.Mode, gin.DebugMode, gin.ReleaseMode, gin.TestMode)
	}

	return conf, nil
}

// Settings of the Todo API server. This is a fairly simple data structure that allows
// loading the configuration from the environment. See the Config() function for more.
type Settings struct {
	Mode        string `default:"debug"`
	UseTLS      bool   `default:"false"`
	Bind        string `default:"127.0.0.1"`
	Port        int    `default:"8080" required:"true"`
	Domain      string `default:"localhost"`
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
	SecretKey   string `envconfig:"SECRET_KEY" required:"true"`
}

// Addr returns the IPADDR:PORT to listen on
func (s Settings) Addr() string {
	return fmt.Sprintf("%s:%d", s.Bind, s.Port)
}

// Endpoint returns the URL to access the API on.
func (s Settings) Endpoint() string {
	if s.UseTLS {
		if s.Port == 443 {
			return fmt.Sprintf("https://%s/", s.Domain)
		}
		return fmt.Sprintf("https://%s:%d/", s.Domain, s.Port)
	}

	if s.Port == 80 {
		return fmt.Sprintf("http://%s/", s.Domain)
	}
	return fmt.Sprintf("http://%s:%d/", s.Domain, s.Port)
}
