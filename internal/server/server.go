package server

import (
	"context"
	"log"
	"mock-cbr/internal/config"
	"mock-cbr/internal/router"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

func Start(cfg *config.Config, rtr *gin.Engine) {

	serv := &http.Server{
		Addr:    cfg.ServPort,
		Handler: rtr,
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

	router.ShuttDown(true)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.GracePeriod)
	defer cancel()

	if err := serv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown - %v", err)
	}

	log.Println("Server exited gracefully")

}
