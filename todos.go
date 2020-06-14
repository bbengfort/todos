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

// API is the Todo server that wraps all context and variables for the handlers.
type API struct {
	sync.RWMutex
	srv     *http.Server // handle to a custom http server with specified API defaults
	router  *gin.Engine  // the http handler and associated middle ware (used for testing)
	db      *gorm.DB     // connection to the database through GORM
	healthy bool         // application state of the server
	done    chan bool    // synchronize shutdown gracefully

	// Temporary
	users  map[string]User
	tokens map[uuid.UUID]Token
}

// New creates a Todos API server with the specified options, ready for running.
func New() (api *API, err error) {
	api = &API{
		healthy: false,
		done:    make(chan bool),
	}

	// Temporary data store
	api.users = make(map[string]User)
	api.tokens = make(map[uuid.UUID]Token)

	// Create the router
	api.router = gin.Default()
	if err = api.setupRoutes(); err != nil {
		return nil, err
	}

	// Create the http server
	api.srv = &http.Server{
		Addr:         ":8080",
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

	// TODO: log server is ready to handle requests at %s
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	// wait until shutdown is complete
	<-s.done
	// TODO: log server(s) stopped
	return nil
}

// Shutdown the API server gracefully
func (s *API) Shutdown() (err error) {
	// TODO: log shutdown message
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
	s.router.POST("/register", s.Register)

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
