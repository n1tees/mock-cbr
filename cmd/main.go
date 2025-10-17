package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"mock-cbr/internal/config"
	"mock-cbr/internal/db"
	"mock-cbr/internal/handlers"

	"github.com/gin-gonic/gin"
)

var (
	shuttingDown = false
)

func main() {

	// config
	config.LoadEnv()
	cfg := config.GetConfig()

	// database
	database, err := db.InitDB(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize DB: %v", err)
	}
	defer database.Close()

	// routes
	router := gin.Default()

	router.Use(func(c *gin.Context) {
		if shuttingDown {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "server is shutting down",
			})
			return
		}
		c.Next()
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "working"})
	})

	router.GET("/scripts/XML_daily.asp", func(c *gin.Context) {
		handlers.RateHandler(c, database)
	})

	// server
	serv := &http.Server{
		Addr:    cfg.ServPort,
		Handler: router,
	}

	go func() {
		log.Printf("mock service is running, port - %s, version - %s", cfg.ServPort, cfg.ServVersion)

		if err := serv.ListenAndServe(); err != nil && err != http.ErrServerClosed {

			log.Fatalf("server error - %v", err)

		}
	}()

	// gracefull shutt down
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	shuttingDown = true

	ctx, cancel := context.WithTimeout(context.Background(), cfg.GracePeriod)
	defer cancel()

	if err := serv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown - %v", err)
	}

	log.Println("Server exited gracefully")

}
