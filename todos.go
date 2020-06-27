package todos

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

// Version of the TODOs application
const Version = "1.0"

var logger = log.New(os.Stderr, "[todos] ", log.LstdFlags)

// API is the Todo server that wraps all context and variables for the handlers.
type API struct {
	sync.RWMutex
	conf    Settings     // configuration of the server
	srv     *http.Server // handle to a custom http server with specified API defaults
	router  *gin.Engine  // the http handler and associated middle ware (used for testing)
	db      *gorm.DB     // connection to the database through GORM
	healthy bool         // application state of the server
	done    chan bool    // synchronize shutdown gracefully

	// Temporary
	users  map[uint]User
	tokens map[uuid.UUID]Token
}

// New creates a Todos API server with the specified settings, fully initialized and
// ready to be run, but without causing any connections to be established.
func New(conf Settings) (api *API, err error) {
	api = &API{
		conf:    conf,
		healthy: false,
		done:    make(chan bool),
	}

	// Temporary data store
	api.users = make(map[uint]User)
	api.tokens = make(map[uuid.UUID]Token)

	// Create the router
	gin.SetMode(api.conf.Mode)
	api.router = gin.Default()
	if err = api.setupRoutes(); err != nil {
		return nil, err
	}

	// Create the http server
	api.srv = &http.Server{
		Addr:         api.conf.Addr(),
		Handler:      api.router,
		ErrorLog:     log.New(os.Stderr, "[http] ", log.LstdFlags),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	return api, nil
}

// Serve the Todos API with the internal http server and specified routes.
func (s *API) Serve() (err error) {
	s.setHealth(true)
	s.osSignals()

	logger.Printf("todo server listening on %s", s.conf.Endpoint())
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	// wait until shutdown is complete
	<-s.done
	logger.Printf("todo server stopped")
	return nil
}

// Routes returns the API router and is primarily exposed for testing purposes.
func (s *API) Routes(healthy bool) http.Handler {
	s.setHealth(healthy)
	return s.router
}

// Shutdown the API server gracefully
func (s *API) Shutdown() (err error) {
	logger.Printf("gracefully shutting down todo server")
	s.setHealth(false)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.srv.SetKeepAlivesEnabled(false)
	if err := s.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("could not gracefully shutdown server: %s", err)
	}

	close(s.done)
	return nil
}

func (s *API) setupRoutes() (err error) {
	// Middleware
	s.router.Use(s.Available())

	// Heartbeat route
	s.router.GET("/status", s.Status)

	// Authentication and user management routes
	s.router.POST("/login", s.Login)
	s.router.POST("/logout", s.Logout)
	s.router.POST("/refresh", s.Refresh)
	s.router.POST("/register", s.Register) // TODO: make registration admin only

	// Application routes
	s.router.GET("/", s.Authorize(), s.Overview)

	return nil
}

func (s *API) setHealth(health bool) {
	s.Lock()
	s.healthy = health
	s.Unlock()
}

func (s *API) osSignals() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		s.Shutdown()
	}()
}
