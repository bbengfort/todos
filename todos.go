package todos

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// Version of the TODOs application
const Version = "1.0"

// Temporary variables for testing the server
var (
	users map[string]User
)

// Serve the Todos API server. Just a quick hack to get started
func Serve() (err error) {

	// temporary
	users = make(map[string]User)

	router := gin.Default()
	router.POST("/login", Login)
	router.POST("/logout", Logout)
	router.POST("/register", Register)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown:", err)
	}

	log.Println("server exiting")
	return nil
}
