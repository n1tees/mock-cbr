package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"mock-cbr/config"
	"mock-cbr/internal/db"

	"github.com/gin-gonic/gin"
)

var (
	shuttingDown = false
)

func main() {

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
		date := c.Query("date_req")
		if date == "" {
			c.String(http.StatusBadRequest, "missing date_req")
			return
		} else {
			c.JSON(http.StatusOK, gin.H{"status": "There will be a sequel here"})
		}
	})

	// server
	serv := &http.Server{
		Addr:    addres,
		Handler: router,
	}

	go func() {
		log.Printf("mock service is running, port - %s, version - %s", addres, version)

		if err := serv.ListenAndServe(); err != nil && err != http.ErrServerClosed {

			log.Fatalf("server error - %v", err)

		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	shuttingDown = true

	ctx, cancel := context.WithTimeout(context.Background(), timeSD)
	defer cancel()

	if err := serv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown - %v", err)
	}

	log.Println("Server exited gracefully")

}
