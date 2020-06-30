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
	"github.com/jinzhu/gorm"

	// Load database dialects for use with gorm
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

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
}

// New creates a Todos API server with the specified settings, fully initialized and
// ready to be run. Note that this function will attempt to connect to the database and
// migrate the latest schema to it.
func New(conf Settings) (api *API, err error) {
	api = &API{
		conf:    conf,
		healthy: false,
		done:    make(chan bool),
	}

	// Connect to the database
	if err = api.setupDatabase(); err != nil {
		return nil, err
	}

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
	s.SetHealth(true)
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

// Shutdown the API server gracefully
func (s *API) Shutdown() (err error) {
	logger.Printf("gracefully shutting down todo server")
	s.SetHealth(false)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.srv.SetKeepAlivesEnabled(false)
	if err := s.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("could not gracefully shutdown server: %s", err)
	}

	if s.db != nil {
		if err = s.db.Close(); err != nil {
			log.Printf("could not close connection to database: %s\n", err)
		}
	}

	close(s.done)
	return nil
}

// SetHealth sets the health status on the API server, putting it into maintenance mode
// if health is false, and removing maintenance mode if health is true. Here primarily
// for testing purposes since it is unlikely an outside caller can access this.
func (s *API) SetHealth(health bool) {
	s.Lock()
	s.healthy = health
	s.Unlock()
}

// Routes returns the API router and is primarily exposed for testing purposes.
func (s *API) Routes() http.Handler {
	return s.router
}

// DB returns the gorm database and is primarily exposed for testing purposes.
func (s *API) DB() *gorm.DB {
	return s.db
}

func (s *API) setupRoutes() (err error) {
	// Middleware
	s.router.Use(s.Available())
	authorize := s.Authorize()
	administrative := s.Administrative()

	// Redirect the root to the current version root
	s.router.GET("/", s.RedirectVersion)

	// V1 API
	v1 := s.router.Group(VersionURL())
	{
		// Heartbeat route
		v1.GET("/status", s.Status)

		// Authentication and user management routes
		v1.POST("/login", s.Login)
		v1.POST("/logout", s.Logout)
		v1.POST("/refresh", s.Refresh)
		v1.POST("/register", authorize, administrative, s.Register)

		// Application routes
		v1.GET("/", authorize, s.Overview)
		todos := v1.Group("/todos", authorize)
		{
			todos.GET("", s.FindTodos)
			todos.POST("", s.CreateTodo)
			todos.GET("/:id", s.DetailTodo)
			todos.PUT("/:id", s.UpdateTodo)
			todos.DELETE("/:id", s.DeleteTodo)
		}

		lists := v1.Group("/lists", authorize)
		{
			lists.GET("", s.FindLists)
			lists.POST("", s.CreateList)
			lists.GET("/:id", s.DetailList)
			lists.PUT("/:id", s.UpdateList)
			lists.DELETE("/:id", s.DeleteList)
		}
	}

	// NotFound and NotAllowed requests
	s.router.NoRoute(NotFound)
	s.router.NoMethod(NotAllowed)

	return nil
}

func (s *API) setupDatabase() (err error) {
	var dialect string
	if dialect, err = s.conf.DBDialect(); err != nil {
		return err
	}

	if s.db, err = gorm.Open(dialect, s.conf.DatabaseURL); err != nil {
		return err
	}

	// Disable logger unless we're in debug mode
	if s.conf.Mode != gin.DebugMode {
		s.db.LogMode(false)
	}

	// Migrate the database to the latest schema
	if err = Migrate(s.db); err != nil {
		return err
	}

	return nil
}

func (s *API) osSignals() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		s.Shutdown()
	}()
}
